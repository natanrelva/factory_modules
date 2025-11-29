package silence

import (
	"math"
	"testing"
	"testing/quick"
)

// Unit Tests

func TestNewSilenceDetector(t *testing.T) {
	detector := NewSilenceDetector(0.01, 100)
	
	if detector == nil {
		t.Fatal("NewSilenceDetector returned nil")
	}
	
	if detector.threshold != 0.01 {
		t.Errorf("Expected threshold 0.01, got %f", detector.threshold)
	}
}

func TestIsSilence_EmptyAudio(t *testing.T) {
	detector := NewSilenceDetector(0.01, 100)
	
	samples := []float32{}
	
	if !detector.IsSilence(samples) {
		t.Error("Expected empty audio to be classified as silence")
	}
}

func TestIsSilence_ZeroAudio(t *testing.T) {
	detector := NewSilenceDetector(0.01, 100)
	
	// All zeros
	samples := make([]float32, 1000)
	
	if !detector.IsSilence(samples) {
		t.Error("Expected zero audio to be classified as silence")
	}
}

func TestIsSilence_LowEnergy(t *testing.T) {
	detector := NewSilenceDetector(0.01, 100)
	
	// Very low amplitude
	samples := make([]float32, 1000)
	for i := range samples {
		samples[i] = 0.001 // Very quiet
	}
	
	if !detector.IsSilence(samples) {
		t.Error("Expected low energy audio to be classified as silence")
	}
}

func TestIsSilence_HighEnergy(t *testing.T) {
	detector := NewSilenceDetector(0.01, 100)
	
	// High amplitude (speech-like)
	samples := make([]float32, 1000)
	for i := range samples {
		samples[i] = 0.5 * float32(math.Sin(float64(i)*0.1))
	}
	
	if detector.IsSilence(samples) {
		t.Error("Expected high energy audio to NOT be classified as silence")
	}
}

func TestGetEnergy_Zero(t *testing.T) {
	detector := NewSilenceDetector(0.01, 100)
	
	samples := make([]float32, 1000)
	energy := detector.GetEnergy(samples)
	
	if energy != 0 {
		t.Errorf("Expected energy 0, got %f", energy)
	}
}

func TestGetEnergy_Calculation(t *testing.T) {
	detector := NewSilenceDetector(0.01, 100)
	
	// Known values
	samples := []float32{0.1, 0.2, 0.3}
	energy := detector.GetEnergy(samples)
	
	// Expected: sqrt((0.1^2 + 0.2^2 + 0.3^2) / 3) = sqrt(0.14/3) = sqrt(0.0467) ≈ 0.216
	expected := float32(math.Sqrt((0.01 + 0.04 + 0.09) / 3))
	
	diff := energy - expected
	if diff < 0 {
		diff = -diff
	}
	
	if diff > 0.001 {
		t.Errorf("Expected energy %f, got %f", expected, energy)
	}
}

func TestGetStats(t *testing.T) {
	detector := NewSilenceDetector(0.01, 100)
	
	// Process some samples
	silence := make([]float32, 1000)
	speech := make([]float32, 1000)
	for i := range speech {
		speech[i] = 0.5
	}
	
	detector.IsSilence(silence)
	detector.IsSilence(speech)
	detector.IsSilence(silence)
	
	stats := detector.GetStats()
	
	if stats.TotalChecks != 3 {
		t.Errorf("Expected 3 checks, got %d", stats.TotalChecks)
	}
	
	if stats.SilenceDetected != 2 {
		t.Errorf("Expected 2 silence detections, got %d", stats.SilenceDetected)
	}
	
	if stats.SpeechDetected != 1 {
		t.Errorf("Expected 1 speech detection, got %d", stats.SpeechDetected)
	}
	
	expectedRate := 2.0 / 3.0
	diff := stats.SilenceRate - expectedRate
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.001 {
		t.Errorf("Expected silence rate %f, got %f", expectedRate, stats.SilenceRate)
	}
}

func TestThresholdConfiguration(t *testing.T) {
	// Test different thresholds
	thresholds := []float32{0.001, 0.01, 0.1}
	
	samples := make([]float32, 1000)
	for i := range samples {
		samples[i] = 0.05 // Medium amplitude
	}
	
	for _, threshold := range thresholds {
		detector := NewSilenceDetector(threshold, 100)
		result := detector.IsSilence(samples)
		
		// With amplitude 0.05, energy will be ~0.05
		// Should be silence if threshold > 0.05
		// Should be speech if threshold < 0.05
		energy := detector.GetEnergy(samples)
		
		if energy < threshold && !result {
			t.Errorf("Threshold %f: Expected silence (energy %f < threshold)", threshold, energy)
		}
		if energy >= threshold && result {
			t.Errorf("Threshold %f: Expected speech (energy %f >= threshold)", threshold, energy)
		}
	}
}

// Property-Based Tests

// Property 1: Silence Detection Accuracy
// For any audio with RMS energy < threshold, it should be classified as silence
func TestProperty_SilenceDetectionAccuracy(t *testing.T) {
	f := func(amplitude float32) bool {
		// Clamp amplitude to reasonable range
		if amplitude < 0 {
			amplitude = -amplitude
		}
		if amplitude > 1.0 {
			amplitude = 1.0
		}
		
		threshold := float32(0.01)
		detector := NewSilenceDetector(threshold, 100)
		
		// Generate samples with given amplitude
		samples := make([]float32, 100)
		for i := range samples {
			samples[i] = amplitude
		}
		
		energy := detector.GetEnergy(samples)
		isSilence := detector.IsSilence(samples)
		
		// Property: energy < threshold => silence
		if energy < threshold {
			return isSilence
		}
		// Property: energy >= threshold => not silence
		return !isSilence
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property 2: Energy Monotonicity
// For any audio, doubling all amplitudes should increase energy
func TestProperty_EnergyMonotonicity(t *testing.T) {
	f := func(samples []float32) bool {
		if len(samples) == 0 {
			return true
		}
		
		// Clamp samples
		for i := range samples {
			if samples[i] > 1.0 {
				samples[i] = 1.0
			}
			if samples[i] < -1.0 {
				samples[i] = -1.0
			}
		}
		
		detector := NewSilenceDetector(0.01, 100)
		
		energy1 := detector.GetEnergy(samples)
		
		// Double all samples
		doubled := make([]float32, len(samples))
		for i := range samples {
			doubled[i] = samples[i] * 2
			if doubled[i] > 1.0 {
				doubled[i] = 1.0
			}
			if doubled[i] < -1.0 {
				doubled[i] = -1.0
			}
		}
		
		energy2 := detector.GetEnergy(doubled)
		
		// Property: doubled energy should be >= original (or both zero)
		return energy2 >= energy1 || (energy1 == 0 && energy2 == 0)
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property 3: Silence Rate Calculation
// For any sequence of checks, silence rate = silence / total
func TestProperty_SilenceRateCalculation(t *testing.T) {
	f := func(amplitudes []float32) bool {
		if len(amplitudes) == 0 {
			return true
		}
		
		detector := NewSilenceDetector(0.01, 100)
		
		silenceCount := 0
		for _, amp := range amplitudes {
			// Clamp
			if amp < 0 {
				amp = -amp
			}
			if amp > 1.0 {
				amp = 1.0
			}
			
			samples := make([]float32, 10)
			for i := range samples {
				samples[i] = amp
			}
			
			if detector.IsSilence(samples) {
				silenceCount++
			}
		}
		
		stats := detector.GetStats()
		
		if stats.TotalChecks != int64(len(amplitudes)) {
			return false
		}
		
		if stats.SilenceDetected != int64(silenceCount) {
			return false
		}
		
		expectedRate := float64(silenceCount) / float64(len(amplitudes))
		diff := stats.SilenceRate - expectedRate
		if diff < 0 {
			diff = -diff
		}
		
		return diff < 0.001
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property 4: Consistency
// For any audio, calling IsSilence multiple times should return same result
func TestProperty_Consistency(t *testing.T) {
	f := func(samples []float32) bool {
		if len(samples) == 0 {
			return true
		}
		
		// Clamp samples
		for i := range samples {
			if samples[i] > 1.0 {
				samples[i] = 1.0
			}
			if samples[i] < -1.0 {
				samples[i] = -1.0
			}
		}
		
		detector := NewSilenceDetector(0.01, 100)
		
		result1 := detector.IsSilence(samples)
		result2 := detector.IsSilence(samples)
		result3 := detector.IsSilence(samples)
		
		return result1 == result2 && result2 == result3
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Benchmark Tests

func BenchmarkIsSilence_Silence(b *testing.B) {
	detector := NewSilenceDetector(0.01, 100)
	samples := make([]float32, 16000) // 1 second at 16kHz
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.IsSilence(samples)
	}
}

func BenchmarkIsSilence_Speech(b *testing.B) {
	detector := NewSilenceDetector(0.01, 100)
	samples := make([]float32, 16000)
	for i := range samples {
		samples[i] = 0.5 * float32(math.Sin(float64(i)*0.1))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.IsSilence(samples)
	}
}

func BenchmarkGetEnergy(b *testing.B) {
	detector := NewSilenceDetector(0.01, 100)
	samples := make([]float32, 16000)
	for i := range samples {
		samples[i] = 0.5
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.GetEnergy(samples)
	}
}
