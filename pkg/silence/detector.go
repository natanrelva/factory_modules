package silence

import (
	"math"
	"sync"
	"time"
)

// SilenceDetector detects silence in audio based on energy threshold
type SilenceDetector struct {
	threshold float32       // Energy threshold (0.01 recommended)
	minLength time.Duration // Minimum silence duration
	
	// Statistics
	mu              sync.RWMutex
	totalChecks     int64
	silenceDetected int64
	speechDetected  int64
}

// SilenceStats holds silence detection statistics
type SilenceStats struct {
	TotalChecks     int64
	SilenceDetected int64
	SpeechDetected  int64
	SilenceRate     float64
}

// NewSilenceDetector creates a new silence detector
func NewSilenceDetector(threshold float32, minLengthMs int) *SilenceDetector {
	return &SilenceDetector{
		threshold: threshold,
		minLength: time.Duration(minLengthMs) * time.Millisecond,
	}
}

// IsSilence checks if audio samples represent silence
func (sd *SilenceDetector) IsSilence(samples []float32) bool {
	// Empty audio is silence
	if len(samples) == 0 {
		sd.recordSilence()
		return true
	}
	
	// Calculate energy
	energy := sd.GetEnergy(samples)
	
	// Compare with threshold
	isSilence := energy < sd.threshold
	
	// Record statistics
	if isSilence {
		sd.recordSilence()
	} else {
		sd.recordSpeech()
	}
	
	return isSilence
}

// GetEnergy calculates RMS energy of audio samples
func (sd *SilenceDetector) GetEnergy(samples []float32) float32 {
	if len(samples) == 0 {
		return 0
	}
	
	// Calculate RMS (Root Mean Square)
	var sum float32
	for _, sample := range samples {
		sum += sample * sample
	}
	
	rms := float32(math.Sqrt(float64(sum / float32(len(samples)))))
	
	return rms
}

// GetStats returns silence detection statistics
func (sd *SilenceDetector) GetStats() SilenceStats {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	
	silenceRate := 0.0
	if sd.totalChecks > 0 {
		silenceRate = float64(sd.silenceDetected) / float64(sd.totalChecks)
	}
	
	return SilenceStats{
		TotalChecks:     sd.totalChecks,
		SilenceDetected: sd.silenceDetected,
		SpeechDetected:  sd.speechDetected,
		SilenceRate:     silenceRate,
	}
}

// ResetStats resets statistics
func (sd *SilenceDetector) ResetStats() {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	
	sd.totalChecks = 0
	sd.silenceDetected = 0
	sd.speechDetected = 0
}

// SetThreshold updates the energy threshold
func (sd *SilenceDetector) SetThreshold(threshold float32) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	
	sd.threshold = threshold
}

// GetThreshold returns the current threshold
func (sd *SilenceDetector) GetThreshold() float32 {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	
	return sd.threshold
}

// Internal methods

func (sd *SilenceDetector) recordSilence() {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	
	sd.totalChecks++
	sd.silenceDetected++
}

func (sd *SilenceDetector) recordSpeech() {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	
	sd.totalChecks++
	sd.speechDetected++
}
