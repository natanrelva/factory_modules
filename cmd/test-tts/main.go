package main

import (
	"fmt"
	"log"
	"os"
	"time"

	espeak "github.com/user/audio-dubbing-system/pkg/tts-espeak"
)

func main() {
	fmt.Println("🧪 Testing eSpeak TTS Integration")
	fmt.Println("==================================\n")

	// Initialize TTS
	config := espeak.Config{
		Voice:      "en-us",
		Speed:      175, // Normal speed
		Pitch:      50,  // Normal pitch
		SampleRate: 16000,
	}

	tts, err := espeak.NewESpeakTTS(config)
	if err != nil {
		log.Fatalf("Failed to initialize eSpeak: %v", err)
	}
	defer tts.Close()

	fmt.Println()

	// Test cases
	tests := []string{
		"Hello world",
		"Good morning",
		"How are you",
		"I like to program",
		"Thank you",
		"See you later",
	}

	fmt.Println("📝 Running TTS tests...\n")

	successCount := 0
	totalTime := time.Duration(0)
	totalSamples := 0

	for i, text := range tests {
		fmt.Printf("Test %d: '%s'\n", i+1, text)

		start := time.Now()
		samples, err := tts.Synthesize(text)
		elapsed := time.Since(start)
		totalTime += elapsed

		if err != nil {
			fmt.Printf("  ❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("  ✓ Generated: %d samples\n", len(samples))
		fmt.Printf("  ⏱️  Time: %v\n", elapsed)
		
		// Calculate duration
		duration := float64(len(samples)) / 16000.0
		fmt.Printf("  🎵 Duration: %.2fs\n\n", duration)

		if len(samples) > 0 {
			successCount++
			totalSamples += len(samples)
		}
	}

	// Print statistics
	fmt.Println("📊 Statistics")
	fmt.Println("=============")
	fmt.Printf("Tests passed: %d/%d\n", successCount, len(tests))
	fmt.Printf("Average time: %v\n", totalTime/time.Duration(len(tests)))
	fmt.Printf("Total samples: %d\n", totalSamples)
	fmt.Printf("Total audio: %.2fs\n", float64(totalSamples)/16000.0)

	stats := tts.GetStats()
	fmt.Printf("\nTTS stats:\n")
	fmt.Printf("  Sentences synthesized: %d\n", stats.SentencesSynthesized)
	fmt.Printf("  Average latency: %v\n", stats.AverageLatency)
	fmt.Printf("  Errors: %d\n", stats.ErrorCount)

	if successCount == len(tests) {
		fmt.Println("\n✅ All tests passed!")
		fmt.Println("eSpeak TTS integration is working correctly.")
		
		// Optionally save a sample to file
		fmt.Println("\n💾 Saving sample audio to test_output.wav...")
		if err := saveSampleAudio(tts); err != nil {
			fmt.Printf("Warning: Could not save sample: %v\n", err)
		} else {
			fmt.Println("✓ Sample saved! Play with: aplay test_output.wav (Linux) or afplay test_output.wav (macOS)")
		}
	} else {
		fmt.Printf("\n⚠️  %d/%d tests passed\n", successCount, len(tests))
		fmt.Println("Some syntheses may have failed.")
	}
}

func saveSampleAudio(tts *espeak.ESpeakTTS) error {
	// Generate sample audio
	samples, err := tts.Synthesize("This is a test of the eSpeak text to speech system")
	if err != nil {
		return err
	}

	// Convert to WAV format
	wavData := samplesToWAV(samples, 16000)

	// Save to file
	return os.WriteFile("test_output.wav", wavData, 0644)
}

func samplesToWAV(samples []float32, sampleRate int) []byte {
	// WAV header
	header := make([]byte, 44)
	
	// RIFF chunk
	copy(header[0:4], "RIFF")
	dataSize := len(samples) * 2
	fileSize := 36 + dataSize
	header[4] = byte(fileSize)
	header[5] = byte(fileSize >> 8)
	header[6] = byte(fileSize >> 16)
	header[7] = byte(fileSize >> 24)
	copy(header[8:12], "WAVE")
	
	// fmt chunk
	copy(header[12:16], "fmt ")
	header[16] = 16 // Chunk size
	header[20] = 1  // Audio format (PCM)
	header[22] = 1  // Num channels (mono)
	// Sample rate
	header[24] = byte(sampleRate)
	header[25] = byte(sampleRate >> 8)
	header[26] = byte(sampleRate >> 16)
	header[27] = byte(sampleRate >> 24)
	// Byte rate
	byteRate := sampleRate * 2
	header[28] = byte(byteRate)
	header[29] = byte(byteRate >> 8)
	header[30] = byte(byteRate >> 16)
	header[31] = byte(byteRate >> 24)
	header[32] = 2  // Block align
	header[34] = 16 // Bits per sample
	
	// data chunk
	copy(header[36:40], "data")
	header[40] = byte(dataSize)
	header[41] = byte(dataSize >> 8)
	header[42] = byte(dataSize >> 16)
	header[43] = byte(dataSize >> 24)
	
	// Combine header and audio data
	wavData := make([]byte, 44+dataSize)
	copy(wavData, header)
	
	// Convert float32 to int16
	for i, sample := range samples {
		// Clamp
		if sample > 1.0 {
			sample = 1.0
		}
		if sample < -1.0 {
			sample = -1.0
		}
		
		intSample := int16(sample * 32767)
		offset := 44 + i*2
		wavData[offset] = byte(intSample)
		wavData[offset+1] = byte(intSample >> 8)
	}
	
	return wavData
}
