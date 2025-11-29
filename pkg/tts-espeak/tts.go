package ttsespeak

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

// ESpeakTTS provides text-to-speech using eSpeak
type ESpeakTTS struct {
	voice      string
	speed      int // Words per minute (default: 175)
	pitch      int // 0-99 (default: 50)
	sampleRate int
	
	// Statistics
	mu                   sync.RWMutex
	sentencesSynthesized int64
	totalLatency         time.Duration
	errorCount           int64
}

// Config holds eSpeak TTS configuration
type Config struct {
	Voice      string // "en-us", "en-gb", "en"
	Speed      int    // Words per minute (80-450, default: 175)
	Pitch      int    // 0-99 (default: 50)
	SampleRate int    // 16000, 22050, 44100
}

// NewESpeakTTS creates a new eSpeak TTS instance
func NewESpeakTTS(config Config) (*ESpeakTTS, error) {
	if config.Voice == "" {
		config.Voice = "en-us"
	}
	
	if config.Speed == 0 {
		config.Speed = 175 // Default speed
	}
	
	if config.Pitch == 0 {
		config.Pitch = 50 // Default pitch
	}
	
	if config.SampleRate == 0 {
		config.SampleRate = 16000
	}

	tts := &ESpeakTTS{
		voice:      config.Voice,
		speed:      config.Speed,
		pitch:      config.Pitch,
		sampleRate: config.SampleRate,
	}

	// Check if eSpeak is installed
	if err := checkESpeakInstalled(); err != nil {
		return nil, fmt.Errorf("eSpeak not found: %w\nPlease install: sudo apt-get install espeak (Linux) or brew install espeak (macOS)", err)
	}

	log.Printf("✓ eSpeak TTS initialized (voice: %s, speed: %d wpm, pitch: %d)", 
		config.Voice, config.Speed, config.Pitch)

	return tts, nil
}

// Synthesize converts English text to audio samples
func (t *ESpeakTTS) Synthesize(textEN string) ([]float32, error) {
	if textEN == "" {
		return nil, nil
	}

	start := time.Now()

	// Call eSpeak to generate WAV audio
	wavData, err := t.callESpeak(textEN)
	if err != nil {
		t.recordError()
		return nil, fmt.Errorf("eSpeak synthesis failed: %w", err)
	}

	// Convert WAV to float32 samples
	samples, err := wavToSamples(wavData)
	if err != nil {
		t.recordError()
		return nil, fmt.Errorf("WAV conversion failed: %w", err)
	}

	elapsed := time.Since(start)
	t.recordLatency(elapsed)

	log.Printf("eSpeak TTS: Synthesized '%s' → %d samples (%v)", textEN, len(samples), elapsed)

	return samples, nil
}

// callESpeak executes eSpeak command and returns WAV data
func (t *ESpeakTTS) callESpeak(text string) ([]byte, error) {
	// Build eSpeak command
	// espeak -v en-us -s 175 -p 50 --stdout "text"
	args := []string{
		"-v", t.voice,
		"-s", fmt.Sprintf("%d", t.speed),
		"-p", fmt.Sprintf("%d", t.pitch),
		"--stdout", // Output WAV to stdout
		text,
	}

	cmd := exec.Command("espeak", args...)
	
	// Capture stdout (WAV data)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("eSpeak command failed: %w, stderr: %s", err, stderr.String())
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		return nil, fmt.Errorf("eSpeak command timed out")
	}

	return stdout.Bytes(), nil
}

// checkESpeakInstalled verifies eSpeak is available
func checkESpeakInstalled() error {
	cmd := exec.Command("espeak", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("eSpeak not found in PATH")
	}
	return nil
}

// wavToSamples converts WAV data to float32 samples
func wavToSamples(wavData []byte) ([]float32, error) {
	if len(wavData) < 44 {
		return nil, fmt.Errorf("WAV data too short (< 44 bytes)")
	}

	// Skip WAV header (44 bytes)
	audioData := wavData[44:]

	// Convert int16 PCM to float32
	numSamples := len(audioData) / 2
	samples := make([]float32, numSamples)

	for i := 0; i < numSamples; i++ {
		// Read int16 (little-endian)
		offset := i * 2
		if offset+1 >= len(audioData) {
			break
		}
		
		intSample := int16(audioData[offset]) | (int16(audioData[offset+1]) << 8)
		
		// Convert to float32 [-1.0, 1.0]
		samples[i] = float32(intSample) / 32768.0
	}

	return samples, nil
}

// Close releases eSpeak resources
func (t *ESpeakTTS) Close() error {
	log.Println("Closing eSpeak TTS...")
	return nil
}

// GetStats returns TTS statistics
func (t *ESpeakTTS) GetStats() Stats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	avgLatency := time.Duration(0)
	if t.sentencesSynthesized > 0 {
		avgLatency = t.totalLatency / time.Duration(t.sentencesSynthesized)
	}
	
	return Stats{
		SentencesSynthesized: t.sentencesSynthesized,
		AverageLatency:       avgLatency,
		ErrorCount:           t.errorCount,
	}
}

// recordLatency records processing latency
func (t *ESpeakTTS) recordLatency(latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.sentencesSynthesized++
	t.totalLatency += latency
}

// recordError records an error
func (t *ESpeakTTS) recordError() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.errorCount++
}

// Stats holds TTS statistics
type Stats struct {
	SentencesSynthesized int64
	AverageLatency       time.Duration
	ErrorCount           int64
}
