package main

import (
	"fmt"
	"log"
	"os"
	"time"

	ttswindows "github.com/user/audio-dubbing-system/pkg/tts-windows"
)

func main() {
	fmt.Println("🧪 Testing Windows TTS Integration")
	fmt.Println("===================================")
	fmt.Println()

	// Inicializar Windows TTS
	tts, err := ttswindows.NewWindowsTTS("en-us", 175, 16000)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Windows TTS: %v", err)
	}
	defer tts.Close()

	fmt.Println()
	fmt.Println("📝 Running TTS tests...")
	fmt.Println()

	// Casos de teste
	testCases := []string{
		"Hello world",
		"Good morning",
		"How are you?",
		"I like programming",
		"This is a test of the Windows text to speech system",
	}

	successCount := 0
	totalTime := time.Duration(0)

	for i, text := range testCases {
		fmt.Printf("Test %d: '%s'\n", i+1, text)

		start := time.Now()
		samples, err := tts.Synthesize(text)
		elapsed := time.Since(start)
		totalTime += elapsed

		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			continue
		}

		duration := float64(len(samples)) / 16000.0
		fmt.Printf("  ✓ Generated: %d samples\n", len(samples))
		fmt.Printf("  ⏱️  Time: %v\n", elapsed)
		fmt.Printf("  🎵 Duration: %.2fs\n", duration)
		fmt.Println()

		successCount++
	}

	// Resumo
	fmt.Println("📊 Test Summary")
	fmt.Println("===============")
	fmt.Printf("✓ Passed: %d/%d\n", successCount, len(testCases))
	fmt.Printf("⏱️  Total time: %v\n", totalTime)
	fmt.Printf("⏱️  Average time: %v\n", totalTime/time.Duration(len(testCases)))
	fmt.Println()

	if successCount == len(testCases) {
		fmt.Println("✅ All tests passed!")
		fmt.Println("Windows TTS integration is working correctly.")
		os.Exit(0)
	} else {
		fmt.Printf("⚠️  %d/%d tests failed\n", len(testCases)-successCount, len(testCases))
		os.Exit(1)
	}
}
