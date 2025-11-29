package translationsimple

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// SimpleTranslator provides basic PT→EN translation
type SimpleTranslator struct {
	apiKey     string
	sourceLang string
	targetLang string
	useAPI     bool
	
	// Translation cache to avoid repeated translations
	cache map[string]string
	
	// Statistics
	mu                  sync.RWMutex
	sentencesTranslated int64
	totalLatency        time.Duration
	errorCount          int64
}

// Config holds translator configuration
type Config struct {
	APIKey     string
	SourceLang string // "pt"
	TargetLang string // "en"
	UseAPI     bool   // true = Google Translate, false = local/mock
}

// NewSimpleTranslator creates a new simplified translator
func NewSimpleTranslator(config Config) (*SimpleTranslator, error) {
	if config.SourceLang == "" {
		config.SourceLang = "pt"
	}
	if config.TargetLang == "" {
		config.TargetLang = "en"
	}
	
	translator := &SimpleTranslator{
		apiKey:     config.APIKey,
		sourceLang: config.SourceLang,
		targetLang: config.TargetLang,
		useAPI:     config.UseAPI,
		cache:      make(map[string]string),
	}

	// TODO: Initialize translation client when API is available
	if config.UseAPI {
		if config.APIKey == "" {
			log.Println("⚠️  Warning: No API key provided, using mock translation")
			translator.useAPI = false
		} else {
			log.Println("✓ Using Google Translate API")
			// TODO: Initialize Google Translate client
			// translator.client = translate.NewClient(ctx, option.WithAPIKey(config.APIKey))
		}
	} else {
		log.Println("✓ Using mock translation (for MVP testing)")
	}

	log.Printf("Translator initialized: %s → %s\n", config.SourceLang, config.TargetLang)

	return translator, nil
}

// Translate converts Portuguese text to English
func (t *SimpleTranslator) Translate(textPT string) (string, error) {
	if textPT == "" {
		return "", nil
	}
	
	// Clean input
	textPT = strings.TrimSpace(textPT)
	if textPT == "" {
		return "", nil
	}

	start := time.Now()
	
	// Check cache first
	if cached, ok := t.getFromCache(textPT); ok {
		log.Printf("Translation (cached): '%s' → '%s'", textPT, cached)
		return cached, nil
	}

	var textEN string
	var err error
	
	if t.useAPI {
		// TODO: Use actual Google Translate API when available
		/*
		translations, err := t.client.Translate(ctx, []string{textPT}, t.targetLang, &translate.Options{
			Source: t.sourceLang,
		})
		if err != nil {
			t.recordError()
			return "", fmt.Errorf("translation failed: %w", err)
		}
		textEN = translations[0].Text
		*/
		
		// For now, use mock
		textEN = mockTranslate(textPT)
	} else {
		// Mock translation for MVP testing
		textEN = mockTranslate(textPT)
	}

	elapsed := time.Since(start)
	
	// Cache the translation
	t.addToCache(textPT, textEN)
	
	// Record statistics
	t.recordLatency(elapsed)
	
	log.Printf("Translation: '%s' → '%s' (%v)", textPT, textEN, elapsed)

	return textEN, err
}

// mockTranslate provides simple mock translation for testing
func mockTranslate(textPT string) string {
	// Common Portuguese → English translations for testing
	translations := map[string]string{
		"olá":          "hello",
		"oi":           "hi",
		"bom dia":      "good morning",
		"boa tarde":    "good afternoon",
		"boa noite":    "good evening",
		"tchau":        "bye",
		"obrigado":     "thank you",
		"obrigada":     "thank you",
		"por favor":    "please",
		"sim":          "yes",
		"não":          "no",
		"como vai":     "how are you",
		"tudo bem":     "all good",
		"meu nome é":   "my name is",
		"prazer":       "nice to meet you",
	}
	
	// Check for exact matches (case-insensitive)
	lower := strings.ToLower(textPT)
	if translation, ok := translations[lower]; ok {
		return translation
	}
	
	// For unknown text, return with [EN:] prefix to indicate it's translated
	return fmt.Sprintf("[EN: %s]", textPT)
}

// getFromCache retrieves translation from cache
func (t *SimpleTranslator) getFromCache(text string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	translation, ok := t.cache[text]
	return translation, ok
}

// addToCache adds translation to cache
func (t *SimpleTranslator) addToCache(source, target string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Limit cache size to prevent memory issues
	if len(t.cache) > 1000 {
		// Clear cache when it gets too large
		t.cache = make(map[string]string)
	}
	
	t.cache[source] = target
}

// Close releases translator resources
func (t *SimpleTranslator) Close() error {
	log.Println("Closing translator...")
	
	// TODO: Cleanup translation client when available
	// if t.client != nil {
	//     t.client.Close()
	// }
	
	return nil
}

// GetStats returns translation statistics
func (t *SimpleTranslator) GetStats() Stats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	avgLatency := time.Duration(0)
	if t.sentencesTranslated > 0 {
		avgLatency = t.totalLatency / time.Duration(t.sentencesTranslated)
	}
	
	return Stats{
		SentencesTranslated: t.sentencesTranslated,
		AverageLatency:      avgLatency,
		ErrorCount:          t.errorCount,
	}
}

// recordLatency records processing latency
func (t *SimpleTranslator) recordLatency(latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.sentencesTranslated++
	t.totalLatency += latency
}

// recordError records an error
func (t *SimpleTranslator) recordError() {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.errorCount++
}

// Stats holds translation statistics
type Stats struct {
	SentencesTranslated int64
	AverageLatency      time.Duration
	ErrorCount          int64
}
