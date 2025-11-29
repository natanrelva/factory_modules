package integration

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// TTSInterface provides integration with TTS (Text-to-Speech) module
type TTSInterface struct {
	mu           sync.RWMutex
	textChannel  chan string
	frameChannel chan types.PCMFrame
	latency      time.Duration
	running      bool
	textsSent    int64
	framesRecv   int64
}

// TTSMetadata contains metadata for TTS synthesis
type TTSMetadata struct {
	Language     string    // e.g., "en-US"
	VoiceID      string    // Voice identifier
	SpeakerEmbed []float32 // Speaker embedding vector
	ProsodyHints ProsodyInfo
}

// ProsodyInfo contains prosody control information
type ProsodyInfo struct {
	RelativeDuration float64       // 0.8 = faster, 1.2 = slower
	EmphasisLevel    int           // 0 = none, 1 = moderate, 2 = strong
	PauseAfter       time.Duration // Pause duration after synthesis
}

// NewTTSInterface creates a new TTS interface
func NewTTSInterface() *TTSInterface {
	return &TTSInterface{
		textChannel:  make(chan string, 5),
		frameChannel: make(chan types.PCMFrame, 10),
		latency:      200 * time.Millisecond, // Default TTS latency
	}
}

// Start begins the TTS interface
func (tts *TTSInterface) Start() error {
	tts.mu.Lock()
	defer tts.mu.Unlock()

	if tts.running {
		return fmt.Errorf("TTS interface already running")
	}

	tts.running = true
	return nil
}

// Stop stops the TTS interface
func (tts *TTSInterface) Stop() error {
	tts.mu.Lock()
	defer tts.mu.Unlock()

	if !tts.running {
		return nil
	}

	tts.running = false
	return nil
}

// SendText sends text to the TTS module for synthesis
func (tts *TTSInterface) SendText(text string) error {
	tts.mu.RLock()
	if !tts.running {
		tts.mu.RUnlock()
		return fmt.Errorf("TTS interface not running")
	}
	tts.mu.RUnlock()

	select {
	case tts.textChannel <- text:
		tts.mu.Lock()
		tts.textsSent++
		tts.mu.Unlock()
		return nil
	default:
		return fmt.Errorf("TTS text channel full")
	}
}

// SendTextWithMetadata sends text with metadata to the TTS module
func (tts *TTSInterface) SendTextWithMetadata(text string, metadata TTSMetadata) error {
	// For now, just send the text
	// In a real implementation, metadata would be sent separately or encoded
	return tts.SendText(text)
}

// ReceiveFrame returns the channel for receiving synthesized audio frames
func (tts *TTSInterface) ReceiveFrame() <-chan types.PCMFrame {
	return tts.frameChannel
}

// GetLatency returns the current TTS processing latency
func (tts *TTSInterface) GetLatency() time.Duration {
	tts.mu.RLock()
	defer tts.mu.RUnlock()
	return tts.latency
}

// SetLatency sets the TTS processing latency
func (tts *TTSInterface) SetLatency(latency time.Duration) {
	tts.mu.Lock()
	defer tts.mu.Unlock()
	tts.latency = latency
}

// GetTextChannel returns the channel for receiving text (for TTS module to consume)
func (tts *TTSInterface) GetTextChannel() <-chan string {
	return tts.textChannel
}

// SendFrame sends a synthesized audio frame (used by TTS module)
func (tts *TTSInterface) SendFrame(frame types.PCMFrame) error {
	select {
	case tts.frameChannel <- frame:
		tts.mu.Lock()
		tts.framesRecv++
		tts.mu.Unlock()
		return nil
	default:
		return fmt.Errorf("TTS frame channel full")
	}
}

// GetStats returns TTS interface statistics
func (tts *TTSInterface) GetStats() (textsSent, framesRecv int64) {
	tts.mu.RLock()
	defer tts.mu.RUnlock()
	return tts.textsSent, tts.framesRecv
}

// Reset resets statistics
func (tts *TTSInterface) Reset() {
	tts.mu.Lock()
	defer tts.mu.Unlock()
	tts.textsSent = 0
	tts.framesRecv = 0
}

// Close closes the TTS interface and releases resources
func (tts *TTSInterface) Close() error {
	tts.Stop()
	
	tts.mu.Lock()
	defer tts.mu.Unlock()
	
	close(tts.textChannel)
	close(tts.frameChannel)
	
	return nil
}

// IsRunning returns whether the TTS interface is running
func (tts *TTSInterface) IsRunning() bool {
	tts.mu.RLock()
	defer tts.mu.RUnlock()
	return tts.running
}
