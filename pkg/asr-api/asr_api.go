package asrapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// ASRAPI provides speech recognition using Whisper API
type ASRAPI struct {
	apiKey   string
	language string
	endpoint string
	
	// Statistics
	mu              sync.RWMutex
	chunksProcessed int64
	totalLatency    time.Duration
	errorCount      int64
}

// Config holds ASR API configuration
type Config struct {
	APIKey   string
	Language string // "pt" for Portuguese
	Endpoint string // Whisper API endpoint
}

// NewASRAPI creates a new ASR instance using API
func NewASRAPI(config Config) (*ASRAPI, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	
	if config.Language == "" {
		config.Language = "pt"
	}
	
	if config.Endpoint == "" {
		// Default to OpenAI Whisper API
		config.Endpoint = "https://api.openai.com/v1/audio/transcriptions"
	}

	asr := &ASRAPI{
		apiKey:   config.APIKey,
		language: config.Language,
		endpoint: config.Endpoint,
	}

	log.Printf("✓ ASR API initialized (language: %s)", config.Language)

	return asr, nil
}

// Transcribe converts audio samples to Portuguese text using API
func (a *ASRAPI) Transcribe(audioSamples []float32) (string, error) {
	if len(audioSamples) == 0 {
		return "", nil
	}

	start := time.Now()

	// Convert float32 samples to WAV format
	wavData, err := samplesToWAV(audioSamples, 16000)
	if err != nil {
		a.recordError()
		return "", fmt.Errorf("failed to convert to WAV: %w", err)
	}

	// Call Whisper API
	text, err := a.callWhisperAPI(wavData)
	if err != nil {
		a.recordError()
		return "", fmt.Errorf("API call failed: %w", err)
	}

	elapsed := time.Since(start)
	a.recordLatency(elapsed)

	if text != "" {
		log.Printf("ASR API: '%s' (%v)", text, elapsed)
	}

	return text, nil
}

// callWhisperAPI makes HTTP request to Whisper API
func (a *ASRAPI) callWhisperAPI(wavData []byte) (string, error) {
	// TODO: Implement proper multipart form with audio file
	// For now, return mock response
	// In real implementation, this would call the actual API
	
	// Prevent unused variable warnings
	_ = wavData
	_ = bytes.Buffer{}
	_ = json.Marshal
	_ = io.Reader(nil)
	_ = http.Client{}
	
	/*
	// Create multipart form data
	body := &bytes.Buffer{}
	
	req, err := http.NewRequest("POST", a.endpoint, body)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "multipart/form-data")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	var result struct {
		Text string `json:"text"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	
	return result.Text, nil
	*/
	
	// Mock response for now
	return "[API: Texto transcrito via Whisper API]", nil
}

// samplesToWAV converts float32 samples to WAV format
func samplesToWAV(samples []float32, sampleRate int) ([]byte, error) {
	// TODO: Implement proper WAV encoding
	// For now, return empty bytes
	// In real implementation, this would create a proper WAV file
	
	buf := &bytes.Buffer{}
	
	// WAV header (44 bytes)
	// RIFF chunk
	buf.WriteString("RIFF")
	// File size (will be updated)
	buf.Write([]byte{0, 0, 0, 0})
	buf.WriteString("WAVE")
	
	// fmt chunk
	buf.WriteString("fmt ")
	buf.Write([]byte{16, 0, 0, 0}) // Chunk size
	buf.Write([]byte{1, 0})         // Audio format (PCM)
	buf.Write([]byte{1, 0})         // Num channels (mono)
	// Sample rate
	buf.Write([]byte{
		byte(sampleRate),
		byte(sampleRate >> 8),
		byte(sampleRate >> 16),
		byte(sampleRate >> 24),
	})
	// Byte rate
	byteRate := sampleRate * 2 // 16-bit mono
	buf.Write([]byte{
		byte(byteRate),
		byte(byteRate >> 8),
		byte(byteRate >> 16),
		byte(byteRate >> 24),
	})
	buf.Write([]byte{2, 0})  // Block align
	buf.Write([]byte{16, 0}) // Bits per sample
	
	// data chunk
	buf.WriteString("data")
	dataSize := len(samples) * 2
	buf.Write([]byte{
		byte(dataSize),
		byte(dataSize >> 8),
		byte(dataSize >> 16),
		byte(dataSize >> 24),
	})
	
	// Convert float32 to int16 and write
	for _, sample := range samples {
		// Clamp to [-1.0, 1.0]
		if sample > 1.0 {
			sample = 1.0
		}
		if sample < -1.0 {
			sample = -1.0
		}
		
		// Convert to int16
		intSample := int16(sample * 32767)
		buf.Write([]byte{
			byte(intSample),
			byte(intSample >> 8),
		})
	}
	
	// Update file size in header
	data := buf.Bytes()
	fileSize := len(data) - 8
	data[4] = byte(fileSize)
	data[5] = byte(fileSize >> 8)
	data[6] = byte(fileSize >> 16)
	data[7] = byte(fileSize >> 24)
	
	return data, nil
}

// Close releases ASR resources
func (a *ASRAPI) Close() error {
	log.Println("Closing ASR API...")
	return nil
}

// GetStats returns ASR statistics
func (a *ASRAPI) GetStats() Stats {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	avgLatency := time.Duration(0)
	if a.chunksProcessed > 0 {
		avgLatency = a.totalLatency / time.Duration(a.chunksProcessed)
	}
	
	return Stats{
		ChunksProcessed: a.chunksProcessed,
		AverageLatency:  avgLatency,
		ErrorCount:      a.errorCount,
	}
}

// recordLatency records processing latency
func (a *ASRAPI) recordLatency(latency time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.chunksProcessed++
	a.totalLatency += latency
}

// recordError records an error
func (a *ASRAPI) recordError() {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.errorCount++
}

// Stats holds ASR statistics
type Stats struct {
	ChunksProcessed int64
	AverageLatency  time.Duration
	ErrorCount      int64
}
