package translationlibre

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// LibreTranslator provides translation using LibreTranslate API
type LibreTranslator struct {
	apiURL     string
	apiKey     string
	sourceLang string
	targetLang string
	
	// Cache
	cache map[string]string
	
	// Statistics
	mu                  sync.RWMutex
	sentencesTranslated int64
	totalLatency        time.Duration
	errorCount          int64
}

// Config holds LibreTranslate configuration
type Config struct {
	APIURL     string // Default: https://libretranslate.com
	APIKey     string // Optional, for rate limit increase
	SourceLang string // "pt"
	TargetLang string // "en"
}

// NewLibreTranslator creates a new LibreTranslate translator
func NewLibreTranslator(config Config) (*LibreTranslator, error) {
	if config.APIURL == "" {
		config.APIURL = "https://libretranslate.com"
	}
	
	if config.SourceLang == "" {
		config.SourceLang = "pt"
	}
	
	if config.TargetLang == "" {
		config.TargetLang = "en"
	}

	translator := &LibreTranslator{
		apiURL:     config.APIURL,
		apiKey:     config.APIKey,
		sourceLang: config.SourceLang,
		targetLang: config.TargetLang,
		cache:      make(map[string]string),
	}

	log.Printf("✓ LibreTranslate initialized (%s → %s)", config.SourceLang, config.TargetLang)
	if config.APIKey != "" {
		log.Println("  Using API key for higher rate limits")
	} else {
		log.Println("  Using public API (rate limited)")
	}

	return translator, nil
}

// Translate converts Portuguese text to English
func (t *LibreTranslator) Translate(textPT string) (string, error) {
	if textPT == "" {
		return "", nil
	}
	
	textPT = strings.TrimSpace(textPT)
	if textPT == "" {
		return "", nil
	}

	// Check cache
	if cached, ok := t.getFromCache(textPT); ok {
		log.Printf("LibreTranslate (cached): '%s' → '%s'", textPT, cached)
		return cached, nil
	}

	start := time.Now()

	// Call LibreTranslate API
	textEN, err := t.callAPI(textPT)
	if err != nil {
		t.recordError()
		return "", fmt.Errorf("translation failed: %w", err)
	}

	elapsed := time.Since(start)
	
	// Cache result
	t.addToCache(textPT, textEN)
	
	// Record statistics
	t.recordLatency(elapsed)
	
	log.Printf("LibreTranslate: '%s' → '%s' (%v)", textPT, textEN, elapsed)

	return textEN, nil
}

// callAPI makes HTTP request to LibreTranslate
func (t *LibreTranslator) callAPI(text string) (string, error) {
	// Prepare request payload
	payload := map[string]string{
		"q":      text,
		"source": t.sourceLang,
		"target": t.targetLang,
	}
	
	if t.apiKey != "" {
		payload["api_key"] = t.apiKey
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Make HTTP request
	url := t.apiURL + "/translate"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var result struct {
		TranslatedText string `json:"translatedText"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	return result.TranslatedText, nil
}

// getFromCache retrieves translation from cache
func (t *LibreTranslator) getFromCache(text string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	translation, ok := t.cache[text]
	return translation, ok
}

// addToCache adds translation to cache
func (t *LibreTranslator) addToCache(source, target string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Limit cache size
	if len(t.cache) > 1000 {
		t.cache = make(map[string]string)
	}
	
	t.cache[source] = target
}

// Close releases translator resources
func (t *LibreTranslator) Close() error {
	log.Println("Closing LibreTranslate...")
	return nil
}

// GetStats returns translation statistics
func (t *LibreTranslator) GetStats() Stats {
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
func (t *LibreTranslator) recordLatency(latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.sentencesTranslated++
	t.totalLatency += latency
}

// recordError records an error
func (t *LibreTranslator) recordError() {
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
