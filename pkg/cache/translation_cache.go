package cache

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// TranslationCache provides LRU caching for translations
type TranslationCache struct {
	cache   map[string]*cacheNode
	maxSize int
	head    *cacheNode // Most recently used
	tail    *cacheNode // Least recently used
	mu      sync.RWMutex
	
	// Statistics
	hits   int64
	misses int64
}

// cacheNode represents a node in the doubly-linked list
type cacheNode struct {
	key         string
	translation string
	timestamp   time.Time
	hitCount    int
	prev        *cacheNode
	next        *cacheNode
}

// CacheStats holds cache statistics
type CacheStats struct {
	Size    int
	MaxSize int
	Hits    int64
	Misses  int64
	HitRate float64
}

// cacheData is used for JSON serialization
type cacheData struct {
	Entries []cacheEntry `json:"entries"`
}

type cacheEntry struct {
	Key         string    `json:"key"`
	Translation string    `json:"translation"`
	Timestamp   time.Time `json:"timestamp"`
	HitCount    int       `json:"hit_count"`
}

// NewTranslationCache creates a new translation cache with specified max size
func NewTranslationCache(maxSize int) *TranslationCache {
	return &TranslationCache{
		cache:   make(map[string]*cacheNode),
		maxSize: maxSize,
	}
}

// Get retrieves a translation from cache
func (tc *TranslationCache) Get(text string) (string, bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	node, found := tc.cache[text]
	if !found {
		tc.misses++
		return "", false
	}
	
	// Update statistics
	tc.hits++
	node.hitCount++
	node.timestamp = time.Now()
	
	// Move to front (most recently used)
	tc.moveToFront(node)
	
	return node.translation, true
}

// Set stores a translation in cache
func (tc *TranslationCache) Set(text, translation string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	// Check if already exists
	if node, found := tc.cache[text]; found {
		// Update existing node
		node.translation = translation
		node.timestamp = time.Now()
		tc.moveToFront(node)
		return
	}
	
	// Create new node
	node := &cacheNode{
		key:         text,
		translation: translation,
		timestamp:   time.Now(),
		hitCount:    0,
	}
	
	// Add to cache
	tc.cache[text] = node
	tc.addToFront(node)
	
	// Check if we need to evict
	if len(tc.cache) > tc.maxSize {
		tc.evictLRU()
	}
}

// GetStats returns cache statistics
func (tc *TranslationCache) GetStats() CacheStats {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	
	total := tc.hits + tc.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(tc.hits) / float64(total)
	}
	
	return CacheStats{
		Size:    len(tc.cache),
		MaxSize: tc.maxSize,
		Hits:    tc.hits,
		Misses:  tc.misses,
		HitRate: hitRate,
	}
}

// Save persists cache to disk
func (tc *TranslationCache) Save(path string) error {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	
	// Collect all entries
	entries := make([]cacheEntry, 0, len(tc.cache))
	for _, node := range tc.cache {
		entries = append(entries, cacheEntry{
			Key:         node.key,
			Translation: node.translation,
			Timestamp:   node.timestamp,
			HitCount:    node.hitCount,
		})
	}
	
	data := cacheData{Entries: entries}
	
	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(path, jsonData, 0644)
}

// Load restores cache from disk
func (tc *TranslationCache) Load(path string) error {
	// Read file
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	// Unmarshal JSON
	var data cacheData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}
	
	// Load entries
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	for _, entry := range data.Entries {
		node := &cacheNode{
			key:         entry.Key,
			translation: entry.Translation,
			timestamp:   entry.Timestamp,
			hitCount:    entry.HitCount,
		}
		
		tc.cache[entry.Key] = node
		tc.addToFront(node)
		
		// Respect max size
		if len(tc.cache) > tc.maxSize {
			tc.evictLRU()
		}
	}
	
	return nil
}

// Clear removes all entries from cache
func (tc *TranslationCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	tc.cache = make(map[string]*cacheNode)
	tc.head = nil
	tc.tail = nil
	tc.hits = 0
	tc.misses = 0
}

// Internal methods for LRU list management

func (tc *TranslationCache) moveToFront(node *cacheNode) {
	if node == tc.head {
		return // Already at front
	}
	
	// Remove from current position
	tc.removeNode(node)
	
	// Add to front
	tc.addToFront(node)
}

func (tc *TranslationCache) addToFront(node *cacheNode) {
	node.next = tc.head
	node.prev = nil
	
	if tc.head != nil {
		tc.head.prev = node
	}
	
	tc.head = node
	
	if tc.tail == nil {
		tc.tail = node
	}
}

func (tc *TranslationCache) removeNode(node *cacheNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		tc.head = node.next
	}
	
	if node.next != nil {
		node.next.prev = node.prev
	} else {
		tc.tail = node.prev
	}
}

func (tc *TranslationCache) evictLRU() {
	if tc.tail == nil {
		return
	}
	
	// Remove from map
	delete(tc.cache, tc.tail.key)
	
	// Remove from list
	if tc.tail.prev != nil {
		tc.tail.prev.next = nil
		tc.tail = tc.tail.prev
	} else {
		tc.head = nil
		tc.tail = nil
	}
}
