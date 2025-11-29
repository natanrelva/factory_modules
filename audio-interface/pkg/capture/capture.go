package capture

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/buffer"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// WindowsAudioCapture implements audio capture for Windows using WASAPI
type WindowsAudioCapture struct {
	mu            sync.RWMutex
	config        types.AudioConfig
	ringBuffer    *buffer.RingBuffer
	frameChannel  chan types.PCMFrame
	running       bool
	initialized   bool
	stopChan      chan struct{}
	latency       time.Duration
	captureErrors int
}

// NewWindowsAudioCapture creates a new Windows audio capture instance
func NewWindowsAudioCapture() *WindowsAudioCapture {
	return &WindowsAudioCapture{
		frameChannel: make(chan types.PCMFrame, 10),
		stopChan:     make(chan struct{}),
		latency:      20 * time.Millisecond, // Default latency
	}
}

// Initialize sets up the audio capture with the given configuration
func (w *WindowsAudioCapture) Initialize(config types.AudioConfig) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.initialized {
		return fmt.Errorf("capture already initialized")
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

	// Create ring buffer for internal buffering
	bufferCapacity := config.BufferSize
	if bufferCapacity <= 0 {
		bufferCapacity = 10 // Default buffer size
	}
	w.ringBuffer = buffer.NewRingBuffer(bufferCapacity)

	w.initialized = true
	return nil
}

// Start begins audio capture
func (w *WindowsAudioCapture) Start() error {
	w.mu.Lock()
	if !w.initialized {
		w.mu.Unlock()
		return fmt.Errorf("capture not initialized")
	}
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("capture already running")
	}
	w.running = true
	w.stopChan = make(chan struct{})
	w.mu.Unlock()

	// Start capture goroutine
	go w.captureLoop()

	return nil
}

// Stop stops audio capture
func (w *WindowsAudioCapture) Stop() error {
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

// GetFrameChannel returns the channel for receiving captured frames
func (w *WindowsAudioCapture) GetFrameChannel() <-chan types.PCMFrame {
	return w.frameChannel
}

// GetCaptureLatency returns the current capture latency
func (w *WindowsAudioCapture) GetCaptureLatency() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.latency
}

// Close releases all resources
func (w *WindowsAudioCapture) Close() error {
	w.Stop()
	
	w.mu.Lock()
	defer w.mu.Unlock()
	
	close(w.frameChannel)
	w.initialized = false
	
	return nil
}

// captureLoop is the main capture loop (runs in goroutine)
func (w *WindowsAudioCapture) captureLoop() {
	ticker := time.NewTicker(w.calculateFrameDuration())
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.captureFrame()
		}
	}
}

// captureFrame captures a single audio frame
func (w *WindowsAudioCapture) captureFrame() {
	w.mu.RLock()
	config := w.config
	w.mu.RUnlock()

	startTime := time.Now()

	// Simulate audio capture (in real implementation, this would call WASAPI)
	// For now, create a frame with silence
	frame := types.PCMFrame{
		Data:       make([]int16, config.FrameSize*config.Channels),
		SampleRate: config.SampleRate,
		Channels:   config.Channels,
		Timestamp:  startTime,
		Duration:   w.calculateFrameDuration(),
		IsSpeech:   false,
	}

	// Try to write to ring buffer
	if !w.ringBuffer.TryWrite(frame) {
		w.mu.Lock()
		w.captureErrors++
		w.mu.Unlock()
		return
	}

	// Try to send to channel (non-blocking)
	select {
	case w.frameChannel <- frame:
		// Update latency measurement
		w.mu.Lock()
		w.latency = time.Since(startTime)
		w.mu.Unlock()
	default:
		// Channel full, skip this frame
		w.mu.Lock()
		w.captureErrors++
		w.mu.Unlock()
	}
}

// calculateFrameDuration calculates the duration of a single frame
func (w *WindowsAudioCapture) calculateFrameDuration() time.Duration {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	if w.config.SampleRate == 0 {
		return 20 * time.Millisecond // Default
	}
	
	// Duration = (FrameSize / SampleRate) seconds
	durationMs := float64(w.config.FrameSize) / float64(w.config.SampleRate) * 1000.0
	return time.Duration(durationMs) * time.Millisecond
}

// GetStats returns capture statistics
func (w *WindowsAudioCapture) GetStats() (captureErrors int, bufferFillLevel float64) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	fillLevel := 0.0
	if w.ringBuffer != nil {
		fillLevel = w.ringBuffer.FillLevel()
	}
	
	return w.captureErrors, fillLevel
}

// ResetStats resets capture statistics
func (w *WindowsAudioCapture) ResetStats() {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.captureErrors = 0
	if w.ringBuffer != nil {
		w.ringBuffer.ResetStats()
	}
}
