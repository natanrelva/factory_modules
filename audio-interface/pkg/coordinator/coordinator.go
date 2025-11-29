package coordinator

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/capture"
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
	mu                sync.RWMutex
	capture           interfaces.AudioCapture
	playback          interfaces.AudioPlayback
	synchronizer      interfaces.StreamSynchronizer
	latencyManager    interfaces.LatencyManager
	metricsCollector  interfaces.MetricsCollector
	config            types.AudioConfig
	running           bool
	stopChan          chan struct{}
	captureWorkerDone chan struct{}
	playbackWorkerDone chan struct{}
	monitorWorkerDone chan struct{}
}

// NewAudioInterfaceCoordinator creates a new coordinator
func NewAudioInterfaceCoordinator(config types.AudioConfig) *AudioInterfaceCoordinator {
	return &AudioInterfaceCoordinator{
		capture:           capture.NewWindowsAudioCapture(),
		playback:          playback.NewWindowsAudioPlayback(),
		synchronizer:      streamsync.NewStreamSynchronizer(),
		latencyManager:    latency.NewLatencyManager(100 * time.Millisecond), // Default 100ms target
		metricsCollector:  metrics.NewMetricsCollector(),
		config:            config,
		stopChan:          make(chan struct{}),
		captureWorkerDone: make(chan struct{}),
		playbackWorkerDone: make(chan struct{}),
		monitorWorkerDone: make(chan struct{}),
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

			// Record capture latency
			captureLatency := c.capture.GetCaptureLatency()
			c.metricsCollector.RecordLatency("capture", captureLatency)

			// Sync timestamps
			c.synchronizer.SyncCapturePlayback(frame.Timestamp, time.Now())

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

// monitorWorker monitors overall system health
func (c *AudioInterfaceCoordinator) monitorWorker() {
	defer close(c.monitorWorkerDone)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return

		case <-ticker.C:
			// Update latency manager
			captureLatency := c.capture.GetCaptureLatency()
			playbackLatency := c.playback.GetPlaybackLatency()
			
			if lm, ok := c.latencyManager.(*latency.LatencyManager); ok {
				lm.UpdateLatency(captureLatency, playbackLatency)

				// Check if optimization needed
				if !lm.IsWithinTarget() {
					// Attempt buffer optimization
					cpuLoad := 0.5 // TODO: Get actual CPU load
					if err := lm.OptimizeBuffers(cpuLoad); err == nil {
						// Optimization successful
						c.metricsCollector.RecordLatency("optimization", time.Since(time.Now()))
					}
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
