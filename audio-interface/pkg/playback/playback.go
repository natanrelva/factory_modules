package playback

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/buffer"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// WindowsAudioPlayback implements audio playback for Windows using WASAPI
type WindowsAudioPlayback struct {
	mu              sync.RWMutex
	config          types.AudioConfig
	ringBuffer      *buffer.RingBuffer
	running         bool
	initialized     bool
	stopChan        chan struct{}
	latency         time.Duration
	bufferFillLevel float64
	playbackErrors  int
	underruns       int
}

// NewWindowsAudioPlayback creates a new Windows audio playback instance
func NewWindowsAudioPlayback() *WindowsAudioPlayback {
	return &WindowsAudioPlayback{
		stopChan:        make(chan struct{}),
		latency:         30 * time.Millisecond, // Default latency
		bufferFillLevel: 0.0,
	}
}

// Initialize sets up the audio playback with the given configuration
func (w *WindowsAudioPlayback) Initialize(config types.AudioConfig) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.initialized {
		return fmt.Errorf("playback already initialized")
	}

	// Validate configuration
	if config.SampleRate <= 0 {
		return fmt.Errorf("invalid sample rate: %d", config.SampleRate)
	}
	if config.Channels <= 0 || config.Channels > 2 {
		return fmt.Errorf("invalid channels: %d (must be 1 or 2)", config.Channels)
	}
	if config.FrameSize <= 0 {
		return fmt.Errorf("invalid frame size: %d", config.FrameSize)
	}

	w.config = config

	// Create adaptive jitter buffer (40-80ms capacity)
	// Calculate buffer capacity based on frame duration
	frameDurationMs := float64(config.FrameSize) / float64(config.SampleRate) * 1000.0
	minBufferCapacity := int(40.0 / frameDurationMs)  // 40ms minimum
	maxBufferCapacity := int(80.0 / frameDurationMs)  // 80ms maximum
	
	bufferCapacity := config.BufferSize
	if bufferCapacity <= 0 {
		bufferCapacity = (minBufferCapacity + maxBufferCapacity) / 2 // Default to middle
	}
	if bufferCapacity < minBufferCapacity {
		bufferCapacity = minBufferCapacity
	}

	w.ringBuffer = buffer.NewRingBuffer(bufferCapacity)

	w.initialized = true
	return nil
}

// Start begins audio playback
func (w *WindowsAudioPlayback) Start() error {
	w.mu.Lock()
	if !w.initialized {
		w.mu.Unlock()
		return fmt.Errorf("playback not initialized")
	}
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("playback already running")
	}
	w.running = true
	w.stopChan = make(chan struct{})
	w.mu.Unlock()

	// Start playback goroutine
	go w.playbackLoop()

	return nil
}

// Stop stops audio playback
func (w *WindowsAudioPlayback) Stop() error {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return nil
	}
	w.running = false
	w.mu.Unlock()

	// Signal stop
	close(w.stopChan)

	return nil
}

// WriteFrame writes a frame to the playback buffer
func (w *WindowsAudioPlayback) WriteFrame(frame types.PCMFrame) error {
	w.mu.RLock()
	if !w.running {
		w.mu.RUnlock()
		return fmt.Errorf("playback not running")
	}
	w.mu.RUnlock()

	// Try to write to ring buffer
	err := w.ringBuffer.Write(frame)
	if err != nil {
		w.mu.Lock()
		w.playbackErrors++
		w.mu.Unlock()
		return err
	}

	// Update buffer fill level
	w.mu.Lock()
	w.bufferFillLevel = w.ringBuffer.FillLevel()
	w.mu.Unlock()

	return nil
}

// GetPlaybackLatency returns the current playback latency
func (w *WindowsAudioPlayback) GetPlaybackLatency() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.latency
}

// GetBufferFillLevel returns the current buffer fill level (0.0 - 1.0)
func (w *WindowsAudioPlayback) GetBufferFillLevel() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.bufferFillLevel
}

// Close releases all resources
func (w *WindowsAudioPlayback) Close() error {
	w.Stop()

	w.mu.Lock()
	defer w.mu.Unlock()

	w.initialized = false

	return nil
}

// playbackLoop is the main playback loop (runs in goroutine)
func (w *WindowsAudioPlayback) playbackLoop() {
	ticker := time.NewTicker(w.calculateFrameDuration())
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.playFrame()
		}
	}
}

// playFrame plays a single audio frame
func (w *WindowsAudioPlayback) playFrame() {
	startTime := time.Now()

	// Try to read from ring buffer
	frame, ok := w.ringBuffer.TryRead()
	if !ok {
		// Buffer underrun - insert silence
		w.mu.Lock()
		w.underruns++
		w.mu.Unlock()
		
		// Simulate playing silence
		w.playSilence()
		return
	}

	// Simulate audio playback (in real implementation, this would call WASAPI)
	w.playAudioData(frame)

	// Update latency measurement
	w.mu.Lock()
	w.latency = time.Since(startTime)
	w.bufferFillLevel = w.ringBuffer.FillLevel()
	w.mu.Unlock()
}

// playAudioData simulates playing audio data
func (w *WindowsAudioPlayback) playAudioData(frame types.PCMFrame) {
	// In real implementation, this would send data to WASAPI
	// For now, just simulate the time it takes
	time.Sleep(time.Microsecond * 100) // Simulate hardware latency
}

// playSilence simulates playing silence when buffer is empty
func (w *WindowsAudioPlayback) playSilence() {
	// In real implementation, this would send silence to WASAPI
	time.Sleep(time.Microsecond * 100)
}

// calculateFrameDuration calculates the duration of a single frame
func (w *WindowsAudioPlayback) calculateFrameDuration() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.config.SampleRate == 0 {
		return 20 * time.Millisecond // Default
	}

	// Duration = (FrameSize / SampleRate) seconds
	durationMs := float64(w.config.FrameSize) / float64(w.config.SampleRate) * 1000.0
	return time.Duration(durationMs) * time.Millisecond
}

// GetStats returns playback statistics
func (w *WindowsAudioPlayback) GetStats() (playbackErrors, underruns int, bufferFillLevel float64) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.playbackErrors, w.underruns, w.bufferFillLevel
}

// ResetStats resets playback statistics
func (w *WindowsAudioPlayback) ResetStats() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.playbackErrors = 0
	w.underruns = 0
	if w.ringBuffer != nil {
		w.ringBuffer.ResetStats()
	}
}

// AdjustBufferSize dynamically adjusts the buffer size based on performance
func (w *WindowsAudioPlayback) AdjustBufferSize(targetLatency time.Duration) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.initialized {
		return fmt.Errorf("playback not initialized")
	}

	// Calculate new buffer capacity based on target latency
	frameDurationMs := float64(w.config.FrameSize) / float64(w.config.SampleRate) * 1000.0
	newCapacity := int(targetLatency.Milliseconds() / int64(frameDurationMs))

	if newCapacity < 2 {
		newCapacity = 2 // Minimum buffer size
	}

	// Create new buffer with adjusted capacity
	// Note: In production, this would need to preserve existing data
	w.ringBuffer = buffer.NewRingBuffer(newCapacity)

	return nil
}
