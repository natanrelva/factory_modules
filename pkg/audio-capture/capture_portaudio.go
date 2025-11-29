// +build portaudio

package audiocapture

import (
	"fmt"
	"log"
	
	"github.com/gordonklaus/portaudio"
)

// tryInitPortAudio initializes PortAudio library
func (c *AudioCapture) tryInitPortAudio() error {
	// Initialize PortAudio
	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize PortAudio: %w", err)
	}
	
	log.Println("✓ PortAudio initialized")
	
	// Get default input device
	device, err := portaudio.DefaultInputDevice()
	if err != nil {
		portaudio.Terminate()
		return fmt.Errorf("failed to get default input device: %w", err)
	}
	
	log.Printf("  Using device: %s", device.Name)
	log.Printf("  Max input channels: %d", device.MaxInputChannels)
	log.Printf("  Default sample rate: %.0f Hz", device.DefaultSampleRate)
	
	// Create stream parameters
	streamParams := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   device,
			Channels: c.channels,
			Latency:  device.DefaultLowInputLatency,
		},
		SampleRate:      float64(c.sampleRate),
		FramesPerBuffer: 512, // Small buffer for low latency
	}
	
	// Open stream
	stream, err := portaudio.OpenStream(streamParams, c.audioCallback)
	if err != nil {
		portaudio.Terminate()
		return fmt.Errorf("failed to open stream: %w", err)
	}
	
	c.stream = stream
	
	return nil
}

// startPortAudioRecording starts the PortAudio stream
func (c *AudioCapture) startPortAudioRecording() error {
	if c.stream == nil {
		return fmt.Errorf("stream not initialized")
	}
	
	if err := c.stream.Start(); err != nil {
		return fmt.Errorf("failed to start stream: %w", err)
	}
	
	log.Println("🎙️  PortAudio stream started - recording from microphone...")
	
	return nil
}

// stopPortAudioRecording stops the PortAudio stream
func (c *AudioCapture) stopPortAudioRecording() error {
	if c.stream == nil {
		return nil
	}
	
	if err := c.stream.Stop(); err != nil {
		return fmt.Errorf("failed to stop stream: %w", err)
	}
	
	log.Println("⏹️  PortAudio stream stopped")
	
	return nil
}

// closePortAudio closes PortAudio resources
func (c *AudioCapture) closePortAudio() error {
	if c.stream != nil {
		c.stream.Close()
		c.stream = nil
	}
	
	if err := portaudio.Terminate(); err != nil {
		return fmt.Errorf("failed to terminate PortAudio: %w", err)
	}
	
	log.Println("✓ PortAudio terminated")
	
	return nil
}

// audioCallback is called by PortAudio when audio data is available
func (c *AudioCapture) audioCallback(in []float32) {
	if len(in) == 0 {
		return
	}
	
	// Add samples to buffer
	c.bufferLock.Lock()
	c.buffer = append(c.buffer, in...)
	
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
	c.totalSamples += int64(len(in))
	c.mu.Unlock()
}
