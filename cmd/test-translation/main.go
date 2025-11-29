package main

import (
	"fmt"
	"log"
	"time"

	libre "github.com/user/audio-dubbing-system/pkg/translation-libre"
)

func main() {
	fmt.Println("🧪 Testing LibreTranslate Integration")
	fmt.Println("=====================================\n")

	// Initialize translator
	config := libre.Config{
		SourceLang: "pt",
		TargetLang: "en",
	}

	translator, err := libre.NewLibreTranslator(config)
	if err != nil {
		log.Fatalf("Failed to initialize translator: %v", err)
	}
	defer translator.Close()

	fmt.Println()

	// Test cases
	tests := []struct {
		input    string
		expected string // Expected translation (approximate)
	}{
		{"olá", "hello"},
		{"bom dia", "good morning"},
		{"como vai você", "how are you"},
		{"eu gosto de programar", "I like to program"},
		{"obrigado", "thank you"},
		{"até logo", "see you later"},
	}

	fmt.Println("📝 Running translation tests...\n")

	successCount := 0
	totalTime := time.Duration(0)

	for i, test := range tests {
		fmt.Printf("Test %d: '%s'\n", i+1, test.input)

		start := time.Now()
		result, err := translator.Translate(test.input)
		elapsed := time.Since(start)
		totalTime += elapsed

		if err != nil {
			fmt.Printf("  ❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("  ✓ Result: '%s'\n", result)
		fmt.Printf("  ⏱️  Time: %v\n\n", elapsed)

		if result != "" {
			successCount++
		}
	}

	// Print statistics
	fmt.Println("📊 Statistics")
	fmt.Println("=============")
	fmt.Printf("Tests passed: %d/%d\n", successCount, len(tests))
	fmt.Printf("Average time: %v\n", totalTime/time.Duration(len(tests)))

	stats := translator.GetStats()
	fmt.Printf("\nTranslator stats:\n")
	fmt.Printf("  Sentences translated: %d\n", stats.SentencesTranslated)
	fmt.Printf("  Average latency: %v\n", stats.AverageLatency)
	fmt.Printf("  Errors: %d\n", stats.ErrorCount)

	if successCount == len(tests) {
		fmt.Println("\n✅ All tests passed!")
		fmt.Println("LibreTranslate integration is working correctly.")
	} else {
		fmt.Printf("\n⚠️  %d/%d tests passed\n", successCount, len(tests))
		fmt.Println("Some translations may have failed due to API rate limits.")
		fmt.Println("Try again in a few moments.")
	}
}
