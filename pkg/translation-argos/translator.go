package translationargos

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ArgosTranslator provides FREE offline translation using Argos Translate
type ArgosTranslator struct {
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

// Config holds Argos Translate configuration
type Config struct {
	SourceLang string // "pt"
	TargetLang string // "en"
}

// NewArgosTranslator creates a new Argos Translate translator
// 100% FREE, works OFFLINE, no API keys needed!
func NewArgosTranslator(config Config) (*ArgosTranslator, error) {
	if config.SourceLang == "" {
		config.SourceLang = "pt"
	}
	
	if config.TargetLang == "" {
		config.TargetLang = "en"
	}

	translator := &ArgosTranslator{
		sourceLang: config.SourceLang,
		targetLang: config.TargetLang,
		cache:      make(map[string]string),
	}

	// Check if Argos Translate is installed
	if err := checkArgosInstalled(); err != nil {
		return nil, fmt.Errorf("Argos Translate not found: %w\n\nInstall with:\n  pip install argostranslate\n  argospm install translate-pt_en", err)
	}

	log.Printf("✓ Argos Translate initialized (%s → %s)", config.SourceLang, config.TargetLang)
	log.Println("  100% FREE, works OFFLINE!")

	return translator, nil
}

// Translate converts Portuguese text to English
// Uses Argos Translate - completely FREE and OFFLINE
func (t *ArgosTranslator) Translate(textPT string) (string, error) {
	if textPT == "" {
		return "", nil
	}
	
	textPT = strings.TrimSpace(textPT)
	if textPT == "" {
		return "", nil
	}

	// Check cache
	if cached, ok := t.getFromCache(textPT); ok {
		log.Printf("Argos (cached): '%s' → '%s'", textPT, cached)
		return cached, nil
	}

	start := time.Now()

	// Call Argos Translate CLI
	textEN, err := t.callArgos(textPT)
	if err != nil {
		t.recordError()
		return "", fmt.Errorf("translation failed: %w", err)
	}

	elapsed := time.Since(start)
	
	// Cache result
	t.addToCache(textPT, textEN)
	
	// Record statistics
	t.recordLatency(elapsed)
	
	log.Printf("Argos: '%s' → '%s' (%v)", textPT, textEN, elapsed)

	return textEN, nil
}

// callArgos executes Argos Translate command
func (t *ArgosTranslator) callArgos(text string) (string, error) {
	// Build command: argos-translate --from pt --to en "text"
	args := []string{
		"--from", t.sourceLang,
		"--to", t.targetLang,
		text,
	}

	cmd := exec.Command("argos-translate", args...)
	
	// Execute with timeout
	done := make(chan error, 1)
	var output []byte
	var err error
	
	go func() {
		output, err = cmd.Output()
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("argos-translate command failed: %w", err)
		}
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		return "", fmt.Errorf("argos-translate command timed out")
	}

	// Parse output
	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("argos-translate returned empty result")
	}

	return result, nil
}

// checkArgosInstalled verifies Argos Translate is available
func checkArgosInstalled() error {
	cmd := exec.Command("argos-translate", "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("argos-translate not found in PATH")
	}
	return nil
}

// getFromCache retrieves translation from cache
func (t *ArgosTranslator) getFromCache(text string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	translation, ok := t.cache[text]
	return translation, ok
}

// addToCache adds translation to cache
func (t *ArgosTranslator) addToCache(source, target string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Limit cache size
	if len(t.cache) > 1000 {
		t.cache = make(map[string]string)
	}
	
	t.cache[source] = target
}

// Close releases translator resources
func (t *ArgosTranslator) Close() error {
	log.Println("Closing Argos Translate...")
	return nil
}

// GetStats returns translation statistics
func (t *ArgosTranslator) GetStats() Stats {
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
func (t *ArgosTranslator) recordLatency(latency time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	t.sentencesTranslated++
	t.totalLatency += latency
}

// recordError records an error
func (t *ArgosTranslator) recordError() {
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
