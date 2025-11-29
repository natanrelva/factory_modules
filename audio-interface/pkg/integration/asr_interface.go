package integration

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// ASRInterface provides integration with ASR (Automatic Speech Recognition) module
type ASRInterface struct {
	mu            sync.RWMutex
	frameChannel  chan types.PCMFrame
	resultChannel chan ASRResult
	latency       time.Duration
	running       bool
	framesSent    int64
	resultsRecv   int64
}

// ASRResult represents a speech recognition result
type ASRResult struct {
	Text       string
	Confidence float64
	Timestamp  time.Time
	IsFinal    bool
	Language   string // e.g., "pt-BR"
}

// NewASRInterface creates a new ASR interface
func NewASRInterface() *ASRInterface {
	return &ASRInterface{
		frameChannel:  make(chan types.PCMFrame, 10),
		resultChannel: make(chan ASRResult, 5),
		latency:       200 * time.Millisecond, // Default ASR latency
	}
}

// Start begins the ASR interface
func (asr *ASRInterface) Start() error {
	asr.mu.Lock()
	defer asr.mu.Unlock()

	if asr.running {
		return fmt.Errorf("ASR interface already running")
	}

	asr.running = true
	return nil
}

// Stop stops the ASR interface
func (asr *ASRInterface) Stop() error {
	asr.mu.Lock()
	defer asr.mu.Unlock()

	if !asr.running {
		return nil
	}

	asr.running = false
	return nil
}

// SendFrame sends a PCM frame to the ASR module
func (asr *ASRInterface) SendFrame(frame types.PCMFrame) error {
	asr.mu.RLock()
	if !asr.running {
		asr.mu.RUnlock()
		return fmt.Errorf("ASR interface not running")
	}
	asr.mu.RUnlock()

	select {
	case asr.frameChannel <- frame:
		asr.mu.Lock()
		asr.framesSent++
		asr.mu.Unlock()
		return nil
	default:
		return fmt.Errorf("ASR frame channel full")
	}
}

// ReceiveResult returns the channel for receiving ASR results
func (asr *ASRInterface) ReceiveResult() <-chan ASRResult {
	return asr.resultChannel
}

// GetLatency returns the current ASR processing latency
func (asr *ASRInterface) GetLatency() time.Duration {
	asr.mu.RLock()
	defer asr.mu.RUnlock()
	return asr.latency
}

// SetLatency sets the ASR processing latency
func (asr *ASRInterface) SetLatency(latency time.Duration) {
	asr.mu.Lock()
	defer asr.mu.Unlock()
	asr.latency = latency
}

// GetFrameChannel returns the channel for receiving frames (for ASR module to consume)
func (asr *ASRInterface) GetFrameChannel() <-chan types.PCMFrame {
	return asr.frameChannel
}

// SendResult sends a recognition result (used by ASR module)
func (asr *ASRInterface) SendResult(result ASRResult) error {
	select {
	case asr.resultChannel <- result:
		asr.mu.Lock()
		asr.resultsRecv++
		asr.mu.Unlock()
		return nil
	default:
		return fmt.Errorf("ASR result channel full")
	}
}

// GetStats returns ASR interface statistics
func (asr *ASRInterface) GetStats() (framesSent, resultsRecv int64) {
	asr.mu.RLock()
	defer asr.mu.RUnlock()
	return asr.framesSent, asr.resultsRecv
}

// Reset resets statistics
func (asr *ASRInterface) Reset() {
	asr.mu.Lock()
	defer asr.mu.Unlock()
	asr.framesSent = 0
	asr.resultsRecv = 0
}

// Close closes the ASR interface and releases resources
func (asr *ASRInterface) Close() error {
	asr.Stop()
	
	asr.mu.Lock()
	defer asr.mu.Unlock()
	
	close(asr.frameChannel)
	close(asr.resultChannel)
	
	return nil
}

// IsRunning returns whether the ASR interface is running
func (asr *ASRInterface) IsRunning() bool {
	asr.mu.RLock()
	defer asr.mu.RUnlock()
	return asr.running
}
