package audiocapture

import (
	"fmt"
	"log"
	"sync"
	"time"
	
	audiocapturepython "github.com/user/audio-dubbing-system/pkg/audio-capture-python"
)

// AudioCapture provides real-time audio capture from microphone
type AudioCapture struct {
	deviceName string
	sampleRate int
	channels   int
	
	// Buffer for captured audio
	buffer     []float32
	bufferLock sync.Mutex
	
	// PortAudio stream (when available)
	stream interface{} // *portaudio.Stream when portaudio build tag is used
	
	// PyAudio capture (Python-based)
	pythonCapture *audiocapturepython.PythonAudioCapture
	
	// Statistics
	mu              sync.RWMutex
	chunksRecorded  int64
	totalSamples    int64
	isRecording     bool
	usePortAudio    bool
	usePyAudio      bool
}

// Config holds audio capture configuration
type Config struct {
	DeviceName string // Device name or empty for default
	SampleRate int    // 16000, 44100, 48000
	Channels   int    // 1 (mono) or 2 (stereo)
	BufferSize int    // Buffer size in samples
}

// NewAudioCapture creates a new audio capture instance
func NewAudioCapture(config Config) (*AudioCapture, error) {
	if config.SampleRate == 0 {
		config.SampleRate = 16000
	}
	
	if config.Channels == 0 {
		config.Channels = 1 // Mono
	}
	
	capture := &AudioCapture{
		deviceName: config.DeviceName,
		sampleRate: config.SampleRate,
		channels:   config.Channels,
		buffer:     make([]float32, 0, config.SampleRate*5), // 5 seconds buffer
	}
	
	// Try to initialize PyAudio first (Python-based, easier to install)
	pythonCapture, err := audiocapturepython.NewPythonAudioCapture(config.DeviceName, config.SampleRate)
	if err == nil {
		log.Println("✓ Using PyAudio for real microphone capture")
		capture.pythonCapture = pythonCapture
		capture.usePyAudio = true
		capture.usePortAudio = false
	} else {
		log.Printf("⚠️  PyAudio not available: %v", err)
		
		// Try to initialize PortAudio (if available)
		// This will only work if compiled with: go build -tags portaudio
		if err := capture.tryInitPortAudio(); err != nil {
			log.Printf("⚠️  PortAudio not available: %v", err)
			log.Println("Using simulated audio capture (for testing)")
			capture.usePortAudio = false
			capture.usePyAudio = false
		} else {
			log.Println("✓ Using PortAudio for real microphone capture")
			capture.usePortAudio = true
			capture.usePyAudio = false
		}
	}
	
	log.Printf("✓ Audio Capture initialized")
	log.Printf("  Device: %s", getDeviceName(config.DeviceName))
	log.Printf("  Sample Rate: %d Hz", config.SampleRate)
	log.Printf("  Channels: %d", config.Channels)
	log.Printf("  Mode: %s", getMode(capture.usePyAudio, capture.usePortAudio))
	
	return capture, nil
}

func getMode(usePyAudio, usePortAudio bool) string {
	if usePyAudio {
		return "Real (PyAudio)"
	}
	if usePortAudio {
		return "Real (PortAudio)"
	}
	return "Simulated (for testing)"
}

// StartRecording starts capturing audio from microphone
func (c *AudioCapture) StartRecording() error {
	c.mu.Lock()
	if c.isRecording {
		c.mu.Unlock()
		return fmt.Errorf("already recording")
	}
	c.isRecording = true
	c.mu.Unlock()
	
	log.Println("🎙️  Started recording from microphone...")
	
	if c.usePyAudio {
		// Use PyAudio (Python-based)
		// Recording happens on-demand in GetChunk()
		return nil
	}
	
	if c.usePortAudio {
		// Start PortAudio stream
		return c.startPortAudioRecording()
	}
	
	// Fallback: simulate recording
	go c.simulateRecording()
	
	return nil
}

// StopRecording stops capturing audio
func (c *AudioCapture) StopRecording() error {
	c.mu.Lock()
	c.isRecording = false
	c.mu.Unlock()
	
	log.Println("⏹️  Stopped recording")
	
	if c.usePortAudio {
		return c.stopPortAudioRecording()
	}
	
	return nil
}

// GetChunk retrieves the last N seconds of audio
func (c *AudioCapture) GetChunk(durationSeconds float64) []float32 {
	// If using PyAudio, capture in real-time
	if c.usePyAudio && c.pythonCapture != nil {
		samples, err := c.pythonCapture.Capture(durationSeconds)
		if err != nil {
			log.Printf("⚠️  PyAudio capture failed: %v", err)
			// Fallback to simulated
			return c.getSimulatedChunk(durationSeconds)
		}
		
		c.mu.Lock()
		c.chunksRecorded++
		c.totalSamples += int64(len(samples))
		c.mu.Unlock()
		
		return samples
	}
	
	// Otherwise use buffer-based approach
	c.bufferLock.Lock()
	defer c.bufferLock.Unlock()
	
	numSamples := int(durationSeconds * float64(c.sampleRate))
	
	// Get last N samples from buffer
	if len(c.buffer) < numSamples {
		numSamples = len(c.buffer)
	}
	
	if numSamples == 0 {
		return []float32{}
	}
	
	// Copy last N samples
	chunk := make([]float32, numSamples)
	startIdx := len(c.buffer) - numSamples
	copy(chunk, c.buffer[startIdx:])
	
	return chunk
}

// getSimulatedChunk generates simulated audio chunk
func (c *AudioCapture) getSimulatedChunk(durationSeconds float64) []float32 {
	numSamples := int(durationSeconds * float64(c.sampleRate))
	samples := make([]float32, numSamples)
	
	// Generate white noise
	for i := range samples {
		samples[i] = (float32(i%100) / 100.0) * 0.01 // Very low amplitude noise
	}
	
	return samples
}

// ClearBuffer clears the audio buffer
func (c *AudioCapture) ClearBuffer() {
	c.bufferLock.Lock()
	defer c.bufferLock.Unlock()
	
	c.buffer = c.buffer[:0]
}

// Close releases audio capture resources
func (c *AudioCapture) Close() error {
	c.StopRecording()
	log.Println("Closing audio capture...")
	
	if c.usePortAudio {
		return c.closePortAudio()
	}
	
	return nil
}

// GetStats returns capture statistics
func (c *AudioCapture) GetStats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	return Stats{
		ChunksRecorded: c.chunksRecorded,
		TotalSamples:   c.totalSamples,
		IsRecording:    c.isRecording,
	}
}

// Stats holds capture statistics
type Stats struct {
	ChunksRecorded int64
	TotalSamples   int64
	IsRecording    bool
}

// simulateRecording simulates audio recording (placeholder)
// TODO: Replace with real PortAudio implementation
func (c *AudioCapture) simulateRecording() {
	ticker := time.NewTicker(100 * time.Millisecond) // 100ms chunks
	defer ticker.Stop()
	
	for {
		c.mu.RLock()
		recording := c.isRecording
		c.mu.RUnlock()
		
		if !recording {
			break
		}
		
		<-ticker.C
		
		// Simulate capturing audio samples
		// In real implementation, this would come from PortAudio callback
		chunkSize := c.sampleRate / 10 // 100ms of audio
		samples := make([]float32, chunkSize)
		
		// Generate some noise to simulate audio
		// TODO: Replace with real microphone input
		for i := range samples {
			// Very quiet noise
			samples[i] = (float32(i%100) / 10000.0) - 0.005
		}
		
		// Add to buffer
		c.bufferLock.Lock()
		c.buffer = append(c.buffer, samples...)
		
		// Keep buffer size manageable (max 10 seconds)
		maxBufferSize := c.sampleRate * 10
		if len(c.buffer) > maxBufferSize {
			// Remove old samples
			excess := len(c.buffer) - maxBufferSize
			c.buffer = c.buffer[excess:]
		}
		c.bufferLock.Unlock()
		
		// Update stats
		c.mu.Lock()
		c.chunksRecorded++
		c.totalSamples += int64(len(samples))
		c.mu.Unlock()
	}
}

func getDeviceName(name string) string {
	if name == "" {
		return "Default Microphone"
	}
	return name
}


// tryInitPortAudio attempts to initialize PortAudio
// Returns error if PortAudio is not available (not compiled with portaudio tag)
func (c *AudioCapture) tryInitPortAudio() error {
	// This function is overridden in capture_portaudio.go when built with portaudio tag
	return fmt.Errorf("PortAudio not available - compile with: go build -tags portaudio")
}

// startPortAudioRecording starts PortAudio recording
// This is a stub - real implementation is in capture_portaudio.go
func (c *AudioCapture) startPortAudioRecording() error {
	return fmt.Errorf("PortAudio not available")
}

// stopPortAudioRecording stops PortAudio recording
// This is a stub - real implementation is in capture_portaudio.go
func (c *AudioCapture) stopPortAudioRecording() error {
	return nil
}

// closePortAudio closes PortAudio
// This is a stub - real implementation is in capture_portaudio.go
func (c *AudioCapture) closePortAudio() error {
	return nil
}
