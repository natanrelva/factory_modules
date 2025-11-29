package coordinator

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/adaptive"
	"github.com/dubbing-system/audio-interface/pkg/backpressure"
	"github.com/dubbing-system/audio-interface/pkg/capture"
	"github.com/dubbing-system/audio-interface/pkg/integration"
	"github.com/dubbing-system/audio-interface/pkg/interfaces"
	"github.com/dubbing-system/audio-interface/pkg/latency"
	"github.com/dubbing-system/audio-interface/pkg/metrics"
	"github.com/dubbing-system/audio-interface/pkg/playback"
	streamsync "github.com/dubbing-system/audio-interface/pkg/sync"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// AudioInterfaceCoordinator manages all audio interface components
type AudioInterfaceCoordinator struct {
	mu                     sync.RWMutex
	capture                interfaces.AudioCapture
	playback               interfaces.AudioPlayback
	synchronizer           interfaces.StreamSynchronizer
	latencyManager         interfaces.LatencyManager
	metricsCollector       interfaces.MetricsCollector
	config                 types.AudioConfig
	running                bool
	stopChan               chan struct{}
	captureWorkerDone      chan struct{}
	playbackWorkerDone     chan struct{}
	monitorWorkerDone      chan struct{}
	
	// V2.0 Components
	backpressureController *backpressure.BackpressureController
	asrInterface           *integration.ASRInterface
	ttsInterface           *integration.TTSInterface
	adaptivePolicy         *adaptive.AdaptivePolicy
}

// NewAudioInterfaceCoordinator creates a new coordinator
func NewAudioInterfaceCoordinator(config types.AudioConfig) *AudioInterfaceCoordinator {
	return &AudioInterfaceCoordinator{
		capture:                capture.NewWindowsAudioCapture(),
		playback:               playback.NewWindowsAudioPlayback(),
		synchronizer:           streamsync.NewStreamSynchronizer(),
		latencyManager:         latency.NewLatencyManager(100 * time.Millisecond), // Default 100ms target
		metricsCollector:       metrics.NewMetricsCollector(),
		config:                 config,
		stopChan:               make(chan struct{}),
		captureWorkerDone:      make(chan struct{}),
		playbackWorkerDone:     make(chan struct{}),
		monitorWorkerDone:      make(chan struct{}),
		
		// V2.0 Components
		backpressureController: backpressure.NewBackpressureController(),
		adaptivePolicy:         adaptive.NewAdaptivePolicy(),
	}
}

// Initialize initializes all components
func (c *AudioInterfaceCoordinator) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Initialize capture
	if err := c.capture.Initialize(c.config); err != nil {
		c.metricsCollector.RecordError(types.ErrorInfo{
			Module:    "Coordinator",
			Operation: "Initialize",
			Err:       err,
			Context:   "Failed to initialize capture",
		})
		return fmt.Errorf("failed to initialize capture: %w", err)
	}

	// Initialize playback
	if err := c.playback.Initialize(c.config); err != nil {
		c.metricsCollector.RecordError(types.ErrorInfo{
			Module:    "Coordinator",
			Operation: "Initialize",
			Err:       err,
			Context:   "Failed to initialize playback",
		})
		return fmt.Errorf("failed to initialize playback: %w", err)
	}

	return nil
}

// Start starts all components and workers
func (c *AudioInterfaceCoordinator) Start() error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("coordinator already running")
	}
	c.running = true
	c.stopChan = make(chan struct{})
	c.captureWorkerDone = make(chan struct{})
	c.playbackWorkerDone = make(chan struct{})
	c.monitorWorkerDone = make(chan struct{})
	c.mu.Unlock()

	// Start capture
	if err := c.capture.Start(); err != nil {
		c.mu.Lock()
		c.running = false
		c.mu.Unlock()
		return fmt.Errorf("failed to start capture: %w", err)
	}

	// Start playback
	if err := c.playback.Start(); err != nil {
		c.capture.Stop()
		c.mu.Lock()
		c.running = false
		c.mu.Unlock()
		return fmt.Errorf("failed to start playback: %w", err)
	}

	// Start workers
	go c.captureWorker()
	go c.playbackWorker()
	go c.monitorWorker()

	return nil
}

// Stop stops all components gracefully
func (c *AudioInterfaceCoordinator) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false
	c.mu.Unlock()

	// Signal stop
	close(c.stopChan)

	// Wait for workers to finish (with timeout)
	timeout := time.After(5 * time.Second)
	workersFinished := make(chan struct{})

	go func() {
		<-c.captureWorkerDone
		<-c.playbackWorkerDone
		<-c.monitorWorkerDone
		close(workersFinished)
	}()

	select {
	case <-workersFinished:
		// All workers finished
	case <-timeout:
		// Timeout - force stop
	}

	// Stop components
	c.capture.Stop()
	c.playback.Stop()

	return nil
}

// Close releases all resources
func (c *AudioInterfaceCoordinator) Close() error {
	c.Stop()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Close components
	if err := c.capture.Close(); err != nil {
		return fmt.Errorf("failed to close capture: %w", err)
	}

	if err := c.playback.Close(); err != nil {
		return fmt.Errorf("failed to close playback: %w", err)
	}

	return nil
}

// captureWorker handles captured audio frames
func (c *AudioInterfaceCoordinator) captureWorker() {
	defer close(c.captureWorkerDone)

	frameChannel := c.capture.GetFrameChannel()

	for {
		select {
		case <-c.stopChan:
			return

		case frame, ok := <-frameChannel:
			if !ok {
				return
			}

			// Check backpressure
			fillLevel := c.playback.GetBufferFillLevel()
			if c.backpressureController.ShouldApplyBackpressure(fillLevel) {
				throttle := c.backpressureController.GetThrottleDuration()
				if throttle > 0 {
					time.Sleep(throttle)
					c.backpressureController.RecordThrottling(throttle)
					c.metricsCollector.RecordLatency("backpressure", throttle)
				}
			}

			// Record capture latency
			captureLatency := c.capture.GetCaptureLatency()
			c.metricsCollector.RecordLatency("capture", captureLatency)

			// Sync timestamps
			c.synchronizer.SyncCapturePlayback(frame.Timestamp, time.Now())

			// Send to ASR if connected
			if c.asrInterface != nil && c.asrInterface.IsRunning() {
				if err := c.asrInterface.SendFrame(frame); err != nil {
					c.metricsCollector.RecordError(types.ErrorInfo{
						Module:    "Coordinator",
						Operation: "CaptureWorker",
						Err:       err,
						Context:   "Failed to send frame to ASR",
					})
				}
			}

			// Forward frame to playback (with backpressure handling)
			if err := c.playback.WriteFrame(frame); err != nil {
				c.metricsCollector.RecordError(types.ErrorInfo{
					Module:    "Coordinator",
					Operation: "CaptureWorker",
					Err:       err,
					Context:   "Failed to write frame to playback",
				})
			}
		}
	}
}

// playbackWorker monitors playback status
func (c *AudioInterfaceCoordinator) playbackWorker() {
	defer close(c.playbackWorkerDone)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return

		case <-ticker.C:
			// Record playback latency
			playbackLatency := c.playback.GetPlaybackLatency()
			c.metricsCollector.RecordLatency("playback", playbackLatency)

			// Check buffer fill level
			fillLevel := c.playback.GetBufferFillLevel()
			if fillLevel < 0.2 {
				// Buffer running low - potential underrun
				c.metricsCollector.RecordError(types.ErrorInfo{
					Module:    "Coordinator",
					Operation: "PlaybackWorker",
					Context:   fmt.Sprintf("Low buffer fill level: %.2f", fillLevel),
				})
			}
		}
	}
}

// monitorWorker monitors overall system health and applies adaptive policies
func (c *AudioInterfaceCoordinator) monitorWorker() {
	defer close(c.monitorWorkerDone)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return

		case <-ticker.C:
			// Collect metrics
			metrics := c.collectAllMetrics()
			
			// Update latency manager
			captureLatency := c.capture.GetCaptureLatency()
			playbackLatency := c.playback.GetPlaybackLatency()
			
			if lm, ok := c.latencyManager.(*latency.LatencyManager); ok {
				lm.UpdateLatency(captureLatency, playbackLatency)
			}

			// Evaluate adaptive policies
			cpuLoad := 0.5 // TODO: Get actual CPU load
			actions := c.adaptivePolicy.Evaluate(metrics, metrics.Underruns, cpuLoad)
			
			// Apply actions
			for _, action := range actions {
				if err := c.applyAction(action); err != nil {
					c.metricsCollector.RecordError(types.ErrorInfo{
						Module:    "Coordinator",
						Operation: "MonitorWorker",
						Err:       err,
						Context:   fmt.Sprintf("Failed to apply action: %s", action.Type.String()),
					})
				}
			}

			// Record sync metrics
			if ss, ok := c.synchronizer.(*streamsync.StreamSynchronizer); ok {
				if !ss.IsAligned() {
					c.metricsCollector.RecordError(types.ErrorInfo{
						Module:    "Coordinator",
						Operation: "MonitorWorker",
						Context:   "Streams not aligned",
					})
				}
			}
			
			// Record monitoring latency
			c.metricsCollector.RecordLatency("monitor", time.Since(time.Now()))
		}
	}
}

// GetMetrics returns current metrics
func (c *AudioInterfaceCoordinator) GetMetrics() types.LatencyMetrics {
	return c.latencyManager.MonitorLatency()
}

// GetMetricsSummary returns comprehensive metrics summary
func (c *AudioInterfaceCoordinator) GetMetricsSummary() metrics.MetricsSummary {
	return c.metricsCollector.(*metrics.MetricsCollector).GetSummary()
}

// GetLatencyStats returns detailed latency statistics
func (c *AudioInterfaceCoordinator) GetLatencyStats() latency.LatencyStats {
	return c.latencyManager.(*latency.LatencyManager).GetStats()
}

// GetSyncStats returns synchronization statistics
func (c *AudioInterfaceCoordinator) GetSyncStats() streamsync.SyncStats {
	return c.synchronizer.(*streamsync.StreamSynchronizer).GetStats()
}

// IsRunning returns whether the coordinator is running
func (c *AudioInterfaceCoordinator) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// SetTargetLatency sets the target end-to-end latency
func (c *AudioInterfaceCoordinator) SetTargetLatency(target time.Duration) error {
	if lm, ok := c.latencyManager.(*latency.LatencyManager); ok {
		return lm.SetTargetLatency(target)
	}
	return fmt.Errorf("latency manager does not support SetTargetLatency")
}

// GetConfig returns the current audio configuration
func (c *AudioInterfaceCoordinator) GetConfig() types.AudioConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// ConnectASR connects an ASR interface to the coordinator
func (c *AudioInterfaceCoordinator) ConnectASR(asr *integration.ASRInterface) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.running {
		return fmt.Errorf("cannot connect ASR while coordinator is running")
	}
	
	c.asrInterface = asr
	return nil
}

// ConnectTTS connects a TTS interface to the coordinator
func (c *AudioInterfaceCoordinator) ConnectTTS(tts *integration.TTSInterface) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.running {
		return fmt.Errorf("cannot connect TTS while coordinator is running")
	}
	
	c.ttsInterface = tts
	return nil
}

// GetBackpressureStats returns backpressure statistics
func (c *AudioInterfaceCoordinator) GetBackpressureStats() (events int64, duration time.Duration) {
	return c.backpressureController.GetStats()
}

// GetAdaptivePolicyStats returns adaptive policy statistics
func (c *AudioInterfaceCoordinator) GetAdaptivePolicyStats() adaptive.PolicyStats {
	return c.adaptivePolicy.GetStats()
}

// collectAllMetrics collects metrics from all components
func (c *AudioInterfaceCoordinator) collectAllMetrics() types.LatencyMetrics {
	captureLatency := c.capture.GetCaptureLatency()
	playbackLatency := c.playback.GetPlaybackLatency()
	fillLevel := c.playback.GetBufferFillLevel()
	
	// Get underruns from playback stats
	_, underruns, _ := c.playback.(*playback.WindowsAudioPlayback).GetStats()
	
	return types.LatencyMetrics{
		CaptureLatency:  captureLatency,
		PlaybackLatency: playbackLatency,
		BufferFillLevel: fillLevel,
		Underruns:       underruns,
		Timestamp:       time.Now(),
	}
}

// applyAction applies an adaptive policy action
func (c *AudioInterfaceCoordinator) applyAction(action adaptive.Action) error {
	switch action.Type {
	case adaptive.ReduceBuffer:
		// Reduce buffer size
		step := action.Value.(int)
		currentLatency := c.capture.GetCaptureLatency() + c.playback.GetPlaybackLatency()
		targetLatency := currentLatency - time.Duration(step*10)*time.Millisecond
		
		if targetLatency < 20*time.Millisecond {
			targetLatency = 20 * time.Millisecond
		}
		
		return c.synchronizer.AdjustBufferSize(targetLatency)
		
	case adaptive.IncreaseBuffer:
		// Increase buffer size
		step := action.Value.(int)
		currentLatency := c.capture.GetCaptureLatency() + c.playback.GetPlaybackLatency()
		targetLatency := currentLatency + time.Duration(step*10)*time.Millisecond
		
		if targetLatency > 200*time.Millisecond {
			targetLatency = 200 * time.Millisecond
		}
		
		return c.synchronizer.AdjustBufferSize(targetLatency)
		
	case adaptive.SwitchToExclusiveMode:
		// Switch to exclusive WASAPI mode
		if lm, ok := c.latencyManager.(*latency.LatencyManager); ok {
			mode, err := lm.SelectOperationMode()
			if err == nil && mode == types.Exclusive {
				// Mode switched successfully
				return nil
			}
			return err
		}
		return fmt.Errorf("latency manager does not support mode selection")
		
	case adaptive.SwitchToSharedMode:
		// Switch to shared WASAPI mode
		if lm, ok := c.latencyManager.(*latency.LatencyManager); ok {
			mode, err := lm.SelectOperationMode()
			if err == nil && mode == types.Shared {
				// Mode switched successfully
				return nil
			}
			return err
		}
		return fmt.Errorf("latency manager does not support mode selection")
		
	case adaptive.ApplyDriftCompensation:
		// Apply drift compensation
		drift := c.synchronizer.GetDriftCompensation()
		if drift > 5*time.Millisecond || drift < -5*time.Millisecond {
			// Significant drift detected - adjust
			return c.synchronizer.AdjustBufferSize(50 * time.Millisecond)
		}
		return nil
		
	default:
		return fmt.Errorf("unknown action type: %v", action.Type)
	}
}

// GetASRInterface returns the connected ASR interface
func (c *AudioInterfaceCoordinator) GetASRInterface() *integration.ASRInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.asrInterface
}

// GetTTSInterface returns the connected TTS interface
func (c *AudioInterfaceCoordinator) GetTTSInterface() *integration.TTSInterface {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ttsInterface
}

// IsASRConnected returns whether an ASR interface is connected
func (c *AudioInterfaceCoordinator) IsASRConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.asrInterface != nil
}

// IsTTSConnected returns whether a TTS interface is connected
func (c *AudioInterfaceCoordinator) IsTTSConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ttsInterface != nil
}
