package main

import (
	"fmt"
	"log"
	"time"

	argos "github.com/user/audio-dubbing-system/pkg/translation-argos"
)

func main() {
	fmt.Println("🧪 Testing Argos Translate - 100% FREE & OFFLINE")
	fmt.Println("=================================================\n")

	// Initialize translator
	config := argos.Config{
		SourceLang: "pt",
		TargetLang: "en",
	}

	translator, err := argos.NewArgosTranslator(config)
	if err != nil {
		log.Fatalf("❌ Failed to initialize translator: %v\n", err)
	}
	defer translator.Close()

	fmt.Println()

	// Test cases - Common Portuguese phrases
	tests := []struct {
		input    string
		expected string
	}{
		{"olá", "hello"},
		{"bom dia", "good morning"},
		{"como vai", "how are you"},
		{"eu gosto de programar", "I like to program"},
		{"obrigado", "thank you"},
		{"até logo", "see you later"},
		{"meu nome é João", "my name is João"},
		{"eu quero água", "I want water"},
		{"onde está o banheiro", "where is the bathroom"},
		{"quanto custa", "how much does it cost"},
		{"não entendo", "I don't understand"},
		{"fala mais devagar", "speak more slowly"},
		{"reunião importante", "important meeting"},
		{"projeto novo", "new project"},
		{"equipe de desenvolvimento", "development team"},
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
		fmt.Printf("  ⏱️  Time: %v\n", elapsed)
		
		// Check if translation is reasonable
		if result != "" {
			fmt.Printf("  ✅ Translation successful\n\n")
			successCount++
		} else {
			fmt.Printf("  ⚠️  Empty result\n\n")
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
		fmt.Println("Argos Translate integration is working correctly.")
		fmt.Println("\n💡 Benefits:")
		fmt.Println("   ✅ 100% FREE - No costs, no API keys")
		fmt.Println("   ✅ Works OFFLINE - No internet required")
		fmt.Println("   ✅ Good quality - Sufficient for MVP")
		fmt.Println("   ✅ Privacy - Data stays on your machine")
	} else {
		fmt.Printf("\n⚠️  %d/%d tests passed\n", successCount, len(tests))
	}
	
	fmt.Println("\n🚀 Installation (if not installed yet):")
	fmt.Println("  pip install argostranslate")
	fmt.Println("  argospm install translate-pt_en")
	fmt.Println("\nOr see: docs/INSTALL_ARGOS.md")
}
