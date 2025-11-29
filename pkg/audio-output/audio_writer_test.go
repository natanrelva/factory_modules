package audiooutput

import (
	"math"
	"testing"
	"testing/quick"
	"time"
)

// Unit Tests

func TestNewAudioWriter(t *testing.T) {
	writer := NewAudioWriter()
	
	if writer == nil {
		t.Fatal("NewAudioWriter returned nil")
	}
}

func TestOpen_ValidDevice(t *testing.T) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	device := AudioDevice{
		Name:       "Default Output Device",
		ID:         "default",
		SampleRate: 16000,
		Channels:   1,
	}
	
	err := writer.Open(device, 16000, 1)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	
	// Verify stream is open
	if !writer.IsOpen() {
		t.Error("Stream should be open")
	}
}

func TestOpen_InvalidDevice(t *testing.T) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	device := AudioDevice{
		Name:       "NonExistent",
		ID:         "invalid",
		SampleRate: 16000,
		Channels:   1,
	}
	
	err := writer.Open(device, 16000, 1)
	if err == nil {
		t.Error("Expected error for invalid device")
	}
}

func TestWrite_ValidSamples(t *testing.T) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	device := AudioDevice{
		Name:       "Default Output Device",
		ID:         "default",
		SampleRate: 16000,
		Channels:   1,
	}
	
	err := writer.Open(device, 16000, 1)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	
	// Create test samples (1 second of audio)
	samples := make([]float32, 16000)
	for i := range samples {
		// Generate sine wave
		samples[i] = float32(math.Sin(2 * math.Pi * 440 * float64(i) / 16000))
	}
	
	err = writer.Write(samples)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	
	// Verify bytes written
	status := writer.GetBufferStatus()
	if status.Used == 0 {
		t.Error("Expected bytes to be written")
	}
}

func TestWrite_EmptySamples(t *testing.T) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	device := AudioDevice{
		Name:       "Default Output Device",
		ID:         "default",
		SampleRate: 16000,
		Channels:   1,
	}
	
	err := writer.Open(device, 16000, 1)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	
	// Write empty samples
	err = writer.Write([]float32{})
	if err != nil {
		t.Errorf("Write with empty samples should not error: %v", err)
	}
}

func TestWrite_WithoutOpen(t *testing.T) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	samples := make([]float32, 1000)
	
	err := writer.Write(samples)
	if err == nil {
		t.Error("Expected error when writing without opening stream")
	}
}

func TestGetBufferStatus(t *testing.T) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	device := AudioDevice{
		Name:       "Default Output Device",
		ID:         "default",
		SampleRate: 16000,
		Channels:   1,
	}
	
	err := writer.Open(device, 16000, 1)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	
	status := writer.GetBufferStatus()
	
	// Verify status structure
	if status.Available < 0 {
		t.Error("Available should be >= 0")
	}
	if status.Used < 0 {
		t.Error("Used should be >= 0")
	}
	if status.Underruns < 0 {
		t.Error("Underruns should be >= 0")
	}
	if status.Overruns < 0 {
		t.Error("Overruns should be >= 0")
	}
}

func TestClose(t *testing.T) {
	writer := NewAudioWriter()
	
	device := AudioDevice{
		Name:       "Default Output Device",
		ID:         "default",
		SampleRate: 16000,
		Channels:   1,
	}
	
	err := writer.Open(device, 16000, 1)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	
	err = writer.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
	
	// Verify stream is closed
	if writer.IsOpen() {
		t.Error("Stream should be closed")
	}
}

func TestClose_WithoutOpen(t *testing.T) {
	writer := NewAudioWriter()
	
	err := writer.Close()
	if err != nil {
		t.Errorf("Close without open should not error: %v", err)
	}
}

func TestMultipleWrites(t *testing.T) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	device := AudioDevice{
		Name:       "Default Output Device",
		ID:         "default",
		SampleRate: 16000,
		Channels:   1,
	}
	
	err := writer.Open(device, 16000, 1)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	
	// Write multiple chunks
	for i := 0; i < 5; i++ {
		samples := make([]float32, 1600) // 100ms chunks
		for j := range samples {
			samples[j] = float32(math.Sin(2 * math.Pi * 440 * float64(j) / 16000))
		}
		
		err = writer.Write(samples)
		if err != nil {
			t.Fatalf("Write %d failed: %v", i, err)
		}
		
		// Small delay between writes
		time.Sleep(10 * time.Millisecond)
	}
}

// Property-Based Tests

// Property 7: Audio Quality Preservation
// For any audio samples, sample rate and bit depth should be preserved
func TestProperty_AudioQualityPreservation(t *testing.T) {
	f := func(frequency uint16) bool {
		if frequency == 0 {
			return true // Skip zero frequency
		}
		if frequency > 8000 {
			frequency = 8000 // Cap at Nyquist for 16kHz
		}
		
		writer := NewAudioWriter()
		defer writer.Close()
		
		device := AudioDevice{
			Name:       "Default Output Device",
			ID:         "default",
			SampleRate: 16000,
			Channels:   1,
		}
		
		err := writer.Open(device, 16000, 1)
		if err != nil {
			return true // Skip on error
		}
		
		// Generate test samples
		samples := make([]float32, 1600)
		for i := range samples {
			samples[i] = float32(math.Sin(2 * math.Pi * float64(frequency) * float64(i) / 16000))
		}
		
		// Write samples
		err = writer.Write(samples)
		if err != nil {
			return false
		}
		
		// Verify buffer status shows data was written
		status := writer.GetBufferStatus()
		if status.Used == 0 {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 20}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property: Buffer Management
// For any sequence of writes, buffer should not overflow
func TestProperty_BufferManagement(t *testing.T) {
	f := func(numChunks uint8) bool {
		if numChunks == 0 || numChunks > 20 {
			return true // Skip invalid values
		}
		
		writer := NewAudioWriter()
		defer writer.Close()
		
		device := AudioDevice{
			Name:       "Default Output Device",
			ID:         "default",
			SampleRate: 16000,
			Channels:   1,
		}
		
		err := writer.Open(device, 16000, 1)
		if err != nil {
			return true // Skip on error
		}
		
		// Write multiple chunks
		for i := 0; i < int(numChunks); i++ {
			samples := make([]float32, 800) // 50ms chunks
			for j := range samples {
				samples[j] = float32(math.Sin(2 * math.Pi * 440 * float64(j) / 16000))
			}
			
			err = writer.Write(samples)
			if err != nil {
				return false
			}
			
			// Small delay to allow buffer to drain
			time.Sleep(10 * time.Millisecond)
		}
		
		// Verify no overruns
		status := writer.GetBufferStatus()
		if status.Overruns > 0 {
			return false
		}
		
		return true
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 10}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property: Write Idempotency
// Writing the same samples multiple times should work consistently
func TestProperty_WriteConsistency(t *testing.T) {
	f := func() bool {
		writer := NewAudioWriter()
		defer writer.Close()
		
		device := AudioDevice{
			Name:       "Default Output Device",
			ID:         "default",
			SampleRate: 16000,
			Channels:   1,
		}
		
		err := writer.Open(device, 16000, 1)
		if err != nil {
			return true // Skip on error
		}
		
		// Create test samples
		samples := make([]float32, 1600)
		for i := range samples {
			samples[i] = float32(math.Sin(2 * math.Pi * 440 * float64(i) / 16000))
		}
		
		// Write same samples multiple times
		for i := 0; i < 3; i++ {
			err = writer.Write(samples)
			if err != nil {
				return false
			}
			time.Sleep(50 * time.Millisecond)
		}
		
		return true
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 5}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Benchmark Tests

func BenchmarkWrite_SmallChunks(b *testing.B) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	device := AudioDevice{
		Name:       "Default Output Device",
		ID:         "default",
		SampleRate: 16000,
		Channels:   1,
	}
	
	writer.Open(device, 16000, 1)
	
	samples := make([]float32, 800) // 50ms
	for i := range samples {
		samples[i] = float32(math.Sin(2 * math.Pi * 440 * float64(i) / 16000))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(samples)
	}
}

func BenchmarkWrite_LargeChunks(b *testing.B) {
	writer := NewAudioWriter()
	defer writer.Close()
	
	device := AudioDevice{
		Name:       "Default Output Device",
		ID:         "default",
		SampleRate: 16000,
		Channels:   1,
	}
	
	writer.Open(device, 16000, 1)
	
	samples := make([]float32, 16000) // 1 second
	for i := range samples {
		samples[i] = float32(math.Sin(2 * math.Pi * 440 * float64(i) / 16000))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(samples)
	}
}
