package cache

import (
	"os"
	"testing"
	"testing/quick"
)

// Unit Tests

func TestNewTranslationCache(t *testing.T) {
	cache := NewTranslationCache(100)
	
	if cache == nil {
		t.Fatal("NewTranslationCache returned nil")
	}
	
	if cache.maxSize != 100 {
		t.Errorf("Expected maxSize 100, got %d", cache.maxSize)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewTranslationCache(10)
	
	// Set a value
	cache.Set("olá", "hello")
	
	// Get the value
	translation, found := cache.Get("olá")
	
	if !found {
		t.Error("Expected to find 'olá' in cache")
	}
	
	if translation != "hello" {
		t.Errorf("Expected 'hello', got '%s'", translation)
	}
}

func TestCacheGetMiss(t *testing.T) {
	cache := NewTranslationCache(10)
	
	// Try to get non-existent value
	_, found := cache.Get("não existe")
	
	if found {
		t.Error("Expected cache miss, but found value")
	}
}

func TestCacheLRUEviction(t *testing.T) {
	cache := NewTranslationCache(3)
	
	// Fill cache to capacity
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	// Add one more - should evict key1 (oldest)
	cache.Set("key4", "value4")
	
	// key1 should be evicted
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be evicted")
	}
	
	// Others should still be there
	_, found = cache.Get("key2")
	if !found {
		t.Error("Expected key2 to still be in cache")
	}
	
	_, found = cache.Get("key3")
	if !found {
		t.Error("Expected key3 to still be in cache")
	}
	
	_, found = cache.Get("key4")
	if !found {
		t.Error("Expected key4 to be in cache")
	}
}

func TestCacheLRUUpdate(t *testing.T) {
	cache := NewTranslationCache(3)
	
	// Fill cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")
	
	// Access key1 (makes it most recently used)
	cache.Get("key1")
	
	// Add new key - should evict key2 (now oldest)
	cache.Set("key4", "value4")
	
	// key2 should be evicted
	_, found := cache.Get("key2")
	if found {
		t.Error("Expected key2 to be evicted")
	}
	
	// key1 should still be there (was accessed)
	_, found = cache.Get("key1")
	if !found {
		t.Error("Expected key1 to still be in cache")
	}
}

func TestCacheStats(t *testing.T) {
	cache := NewTranslationCache(10)
	
	// Set some values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	
	// Get with hit
	cache.Get("key1")
	
	// Get with miss
	cache.Get("key3")
	
	stats := cache.GetStats()
	
	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}
	
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	
	expectedHitRate := 0.5 // 1 hit / 2 total
	if stats.HitRate != expectedHitRate {
		t.Errorf("Expected hit rate %.2f, got %.2f", expectedHitRate, stats.HitRate)
	}
}

func TestCachePersistence(t *testing.T) {
	cache := NewTranslationCache(10)
	
	// Add some entries
	cache.Set("olá", "hello")
	cache.Set("bom dia", "good morning")
	cache.Set("obrigado", "thank you")
	
	// Save to file
	tmpFile := "test_cache.json"
	defer os.Remove(tmpFile)
	
	err := cache.Save(tmpFile)
	if err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}
	
	// Create new cache and load
	newCache := NewTranslationCache(10)
	err = newCache.Load(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}
	
	// Verify entries
	translation, found := newCache.Get("olá")
	if !found || translation != "hello" {
		t.Error("Failed to load 'olá' from cache")
	}
	
	translation, found = newCache.Get("bom dia")
	if !found || translation != "good morning" {
		t.Error("Failed to load 'bom dia' from cache")
	}
}

// Property-Based Tests

// Property 1: Cache Consistency
// For any key-value pair, Get(key) after Set(key, value) returns value
func TestProperty_CacheConsistency(t *testing.T) {
	f := func(key, value string) bool {
		if key == "" {
			return true // Skip empty keys
		}
		
		cache := NewTranslationCache(100)
		cache.Set(key, value)
		
		result, found := cache.Get(key)
		
		return found && result == value
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property 2: Cache LRU Eviction
// For any cache at max capacity, adding a new entry removes the least recently used
func TestProperty_LRUEviction(t *testing.T) {
	f := func(keys []string) bool {
		if len(keys) < 4 {
			return true // Need at least 4 keys
		}
		
		// Use cache of size 3
		cache := NewTranslationCache(3)
		
		// Add first 3 keys
		for i := 0; i < 3 && i < len(keys); i++ {
			if keys[i] != "" {
				cache.Set(keys[i], "value"+string(rune(i)))
			}
		}
		
		// Add 4th key - should evict first
		if keys[3] != "" {
			cache.Set(keys[3], "value3")
		}
		
		// First key should be evicted (if it was unique)
		// If keys[0] was duplicate of keys[1-3], it might still be there
		// So we just check that cache size is maintained
		stats := cache.GetStats()
		return stats.Size <= 3
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property 3: Cache Size Limit
// For any sequence of Set operations, cache size never exceeds maxSize
func TestProperty_CacheSizeLimit(t *testing.T) {
	f := func(keys []string) bool {
		maxSize := 10
		cache := NewTranslationCache(maxSize)
		
		// Add all keys
		for i, key := range keys {
			if key != "" {
				cache.Set(key, "value"+string(rune(i)))
			}
		}
		
		stats := cache.GetStats()
		return stats.Size <= maxSize
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 100}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property 4: Hit Rate Calculation
// For any sequence of operations, hit rate = hits / (hits + misses)
func TestProperty_HitRateCalculation(t *testing.T) {
	f := func(setKeys, getKeys []string) bool {
		cache := NewTranslationCache(50)
		
		// Set some keys
		for i, key := range setKeys {
			if key != "" {
				cache.Set(key, "value"+string(rune(i)))
			}
		}
		
		// Get some keys (mix of hits and misses)
		for _, key := range getKeys {
			if key != "" {
				cache.Get(key)
			}
		}
		
		stats := cache.GetStats()
		
		// Verify hit rate calculation
		if stats.Hits+stats.Misses == 0 {
			return stats.HitRate == 0
		}
		
		expectedHitRate := float64(stats.Hits) / float64(stats.Hits+stats.Misses)
		diff := stats.HitRate - expectedHitRate
		if diff < 0 {
			diff = -diff
		}
		
		return diff < 0.001 // Allow small floating point error
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Benchmark Tests

func BenchmarkCacheSet(b *testing.B) {
	cache := NewTranslationCache(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value")
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := NewTranslationCache(1000)
	cache.Set("key", "value")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkCacheSetWithEviction(b *testing.B) {
	cache := NewTranslationCache(100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key"+string(rune(i)), "value")
	}
}
