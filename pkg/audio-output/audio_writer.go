package audiooutput

import (
	"fmt"
	"sync"
	"time"
)

// AudioWriter handles low-level audio writing to output devices
type AudioWriter struct {
	mu sync.RWMutex
	
	// Stream state
	device     AudioDevice
	sampleRate int
	channels   int
	isOpen     bool
	
	// Buffer management
	buffer     []float32
	bufferSize int
	
	// Statistics
	bytesWritten int64
	underruns    int64
	overruns     int64
	
	// Simulated playback
	lastWrite time.Time
}

// NewAudioWriter creates a new audio writer
func NewAudioWriter() *AudioWriter {
	return &AudioWriter{
		buffer:     make([]float32, 0),
		bufferSize: 16000 * 2, // 2 seconds buffer at 16kHz
		lastWrite:  time.Now(),
	}
}

// Open opens an audio stream for the specified device
func (aw *AudioWriter) Open(device AudioDevice, sampleRate int, channels int) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	
	if aw.isOpen {
		return fmt.Errorf("stream already open")
	}
	
	// Validate device (simple check for mock implementation)
	if device.ID == "invalid" || device.Name == "NonExistent" {
		return fmt.Errorf("invalid device: %s", device.Name)
	}
	
	// Validate parameters
	if sampleRate <= 0 {
		return fmt.Errorf("invalid sample rate: %d", sampleRate)
	}
	if channels <= 0 {
		return fmt.Errorf("invalid channels: %d", channels)
	}
	
	// TODO: Replace with actual PortAudio stream opening
	// For now, just store configuration
	aw.device = device
	aw.sampleRate = sampleRate
	aw.channels = channels
	aw.isOpen = true
	aw.buffer = make([]float32, 0, aw.bufferSize)
	aw.bytesWritten = 0
	aw.underruns = 0
	aw.overruns = 0
	aw.lastWrite = time.Now()
	
	return nil
}

// Write writes audio samples to the output device
func (aw *AudioWriter) Write(samples []float32) error {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	
	if !aw.isOpen {
		return fmt.Errorf("stream not open")
	}
	
	if len(samples) == 0 {
		return nil // Nothing to write
	}
	
	// TODO: Replace with actual PortAudio writing
	// For now, simulate writing by adding to buffer
	
	// Check for buffer overflow
	if len(aw.buffer)+len(samples) > aw.bufferSize {
		aw.overruns++
		// Drop oldest samples to make room
		overflow := (len(aw.buffer) + len(samples)) - aw.bufferSize
		if overflow < len(aw.buffer) {
			aw.buffer = aw.buffer[overflow:]
		} else {
			aw.buffer = aw.buffer[:0]
		}
	}
	
	// Add samples to buffer
	aw.buffer = append(aw.buffer, samples...)
	aw.bytesWritten += int64(len(samples) * 4) // 4 bytes per float32
	aw.lastWrite = time.Now()
	
	// Simulate buffer draining (playback)
	go aw.simulatePlayback()
	
	return nil
}

// simulatePlayback simulates audio playback by draining the buffer
func (aw *AudioWriter) simulatePlayback() {
	// Calculate playback duration
	aw.mu.Lock()
	if len(aw.buffer) == 0 {
		aw.mu.Unlock()
		return
	}
	
	// Drain some samples based on time elapsed
	elapsed := time.Since(aw.lastWrite)
	samplesToDrain := int(float64(aw.sampleRate) * elapsed.Seconds())
	
	if samplesToDrain > len(aw.buffer) {
		samplesToDrain = len(aw.buffer)
	}
	
	if samplesToDrain > 0 {
		aw.buffer = aw.buffer[samplesToDrain:]
	}
	aw.mu.Unlock()
}

// GetBufferStatus returns the current buffer status
func (aw *AudioWriter) GetBufferStatus() BufferStatus {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	
	return BufferStatus{
		Available: aw.bufferSize - len(aw.buffer),
		Used:      len(aw.buffer),
		Underruns: int(aw.underruns),
		Overruns:  int(aw.overruns),
	}
}

// IsOpen returns whether the stream is open
func (aw *AudioWriter) IsOpen() bool {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	
	return aw.isOpen
}

// Close closes the audio stream
func (aw *AudioWriter) Close() error {
	aw.mu.Lock()
	defer aw.mu.Unlock()
	
	if !aw.isOpen {
		return nil // Already closed
	}
	
	// TODO: Replace with actual PortAudio stream closing
	// For now, just clear state
	aw.isOpen = false
	aw.buffer = nil
	
	return nil
}

// GetDevice returns the current device
func (aw *AudioWriter) GetDevice() AudioDevice {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	
	return aw.device
}

// GetSampleRate returns the current sample rate
func (aw *AudioWriter) GetSampleRate() int {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	
	return aw.sampleRate
}

// GetChannels returns the number of channels
func (aw *AudioWriter) GetChannels() int {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	
	return aw.channels
}

// GetBytesWritten returns total bytes written
func (aw *AudioWriter) GetBytesWritten() int64 {
	aw.mu.RLock()
	defer aw.mu.RUnlock()
	
	return aw.bytesWritten
}
