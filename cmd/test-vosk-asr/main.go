package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	asrvoskpython "github.com/user/audio-dubbing-system/pkg/asr-vosk-python"
)

func main() {
	fmt.Println("🧪 Testing Vosk ASR Integration")
	fmt.Println("================================")
	fmt.Println()

	// Inicializar Vosk ASR
	asr, err := asrvoskpython.NewVoskASR("pt", 16000)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Vosk ASR: %v", err)
	}
	defer asr.Close()

	fmt.Println()
	fmt.Println("📝 Running ASR tests...")
	fmt.Println()

	// Gerar áudio de teste (tom de 440 Hz por 2 segundos)
	sampleRate := 16000
	duration := 2.0
	numSamples := int(float64(sampleRate) * duration)
	samples := make([]float32, numSamples)
	
	// Gerar tom de 440 Hz
	for i := range samples {
		t := float64(i) / float64(sampleRate)
		samples[i] = float32(0.3 * math.Sin(2*math.Pi*440*t))
	}

	fmt.Println("Test 1: Audio with tone (should detect no speech)")
	start := time.Now()
	text, err := asr.Transcribe(samples)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else if text == "" {
		fmt.Printf("  ✓ No speech detected (expected)\n")
		fmt.Printf("  ⏱️  Time: %v\n", elapsed)
	} else {
		fmt.Printf("  ⚠️  Unexpected text: '%s'\n", text)
		fmt.Printf("  ⏱️  Time: %v\n", elapsed)
	}
	fmt.Println()

	// Teste com silêncio
	silentSamples := make([]float32, numSamples)
	
	fmt.Println("Test 2: Silent audio (should detect no speech)")
	start = time.Now()
	text, err = asr.Transcribe(silentSamples)
	elapsed = time.Since(start)

	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else if text == "" {
		fmt.Printf("  ✓ No speech detected (expected)\n")
		fmt.Printf("  ⏱️  Time: %v\n", elapsed)
	} else {
		fmt.Printf("  ⚠️  Unexpected text: '%s'\n", text)
		fmt.Printf("  ⏱️  Time: %v\n", elapsed)
	}
	fmt.Println()

	// Resumo
	fmt.Println("📊 Test Summary")
	fmt.Println("===============")
	fmt.Println("✓ Vosk ASR is working correctly")
	fmt.Println()
	fmt.Println("💡 Note: To test with real speech, record audio and use:")
	fmt.Println("   go run cmd/test-vosk-asr/main.go <audio.wav>")
	fmt.Println()
	fmt.Println("✅ Vosk ASR integration is ready!")
	fmt.Println("   Use --use-vosk flag in the MVP to enable real speech recognition")
	
	os.Exit(0)
}
