package metrics

import (
	"sort"
	"sync"
	"time"
)

const (
	// MaxHistorySize is the maximum number of chunks to keep in history
	MaxHistorySize = 100
)

// MetricsCollector collects and aggregates performance metrics
type MetricsCollector struct {
	mu sync.RWMutex
	
	// Current chunk metrics
	currentChunk ChunkMetrics
	
	// Historical data
	history []ChunkMetrics
	
	// Aggregated counters
	cacheHits   int64
	cacheMisses int64
	silenceSkips int64
	
	startTime time.Time
}

// ChunkMetrics holds metrics for a single processing chunk
type ChunkMetrics struct {
	Timestamp time.Time
	Latencies map[string]time.Duration
	TotalLatency time.Duration
}

// AggregatedStats holds aggregated statistics
type AggregatedStats struct {
	TotalChunks        int64
	AverageLatency     time.Duration
	P50Latency         time.Duration
	P95Latency         time.Duration
	P99Latency         time.Duration
	ComponentLatencies map[string]time.Duration
	CacheHits          int64
	CacheMisses        int64
	CacheHitRate       float64
	SilenceSkips       int64
	Uptime             time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		currentChunk: ChunkMetrics{
			Timestamp: time.Now(),
			Latencies: make(map[string]time.Duration),
		},
		history:   make([]ChunkMetrics, 0, MaxHistorySize),
		startTime: time.Now(),
	}
}

// RecordLatency records a latency measurement for a component
func (mc *MetricsCollector) RecordLatency(component string, latency time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.currentChunk.Latencies[component] = latency
	mc.currentChunk.TotalLatency += latency
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.cacheHits++
}

// RecordCacheMiss records a cache miss
func (mc *MetricsCollector) RecordCacheMiss() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.cacheMisses++
}

// RecordSilenceSkip records a silence skip
func (mc *MetricsCollector) RecordSilenceSkip() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.silenceSkips++
}

// NextChunk moves to the next chunk, saving current metrics to history
func (mc *MetricsCollector) NextChunk() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// Save current chunk to history
	mc.history = append(mc.history, mc.currentChunk)
	
	// Trim history if needed
	if len(mc.history) > MaxHistorySize {
		mc.history = mc.history[len(mc.history)-MaxHistorySize:]
	}
	
	// Reset current chunk
	mc.currentChunk = ChunkMetrics{
		Timestamp: time.Now(),
		Latencies: make(map[string]time.Duration),
	}
}

// GetAggregated returns aggregated statistics
func (mc *MetricsCollector) GetAggregated() AggregatedStats {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	stats := AggregatedStats{
		ComponentLatencies: make(map[string]time.Duration),
		CacheHits:          mc.cacheHits,
		CacheMisses:        mc.cacheMisses,
		SilenceSkips:       mc.silenceSkips,
		Uptime:             time.Since(mc.startTime),
	}
	
	// Include current chunk if it has data
	allChunks := mc.history
	if len(mc.currentChunk.Latencies) > 0 {
		allChunks = append([]ChunkMetrics{mc.currentChunk}, mc.history...)
	}
	
	stats.TotalChunks = int64(len(allChunks))
	
	// Calculate cache hit rate (independent of chunks)
	totalCacheOps := stats.CacheHits + stats.CacheMisses
	if totalCacheOps > 0 {
		stats.CacheHitRate = float64(stats.CacheHits) / float64(totalCacheOps)
	}
	
	if len(allChunks) == 0 {
		return stats
	}
	
	// Calculate component averages
	componentSums := make(map[string]time.Duration)
	componentCounts := make(map[string]int)
	
	var totalLatencySum time.Duration
	totalLatencies := make([]time.Duration, 0, len(allChunks))
	
	for _, chunk := range allChunks {
		totalLatencies = append(totalLatencies, chunk.TotalLatency)
		totalLatencySum += chunk.TotalLatency
		
		for component, latency := range chunk.Latencies {
			componentSums[component] += latency
			componentCounts[component]++
		}
	}
	
	// Calculate averages
	if len(allChunks) > 0 {
		stats.AverageLatency = totalLatencySum / time.Duration(len(allChunks))
		
		for component, sum := range componentSums {
			count := componentCounts[component]
			if count > 0 {
				stats.ComponentLatencies[component] = sum / time.Duration(count)
			}
		}
	}
	
	// Calculate percentiles
	if len(totalLatencies) > 0 {
		sort.Slice(totalLatencies, func(i, j int) bool {
			return totalLatencies[i] < totalLatencies[j]
		})
		
		stats.P50Latency = percentile(totalLatencies, 0.50)
		stats.P95Latency = percentile(totalLatencies, 0.95)
		stats.P99Latency = percentile(totalLatencies, 0.99)
	}
	
	return stats
}

// percentile calculates the percentile value from a sorted slice
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	
	index := int(float64(len(sorted)-1) * p)
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	
	return sorted[index]
}

// Reset clears all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.currentChunk = ChunkMetrics{
		Timestamp: time.Now(),
		Latencies: make(map[string]time.Duration),
	}
	mc.history = make([]ChunkMetrics, 0, MaxHistorySize)
	mc.cacheHits = 0
	mc.cacheMisses = 0
	mc.silenceSkips = 0
	mc.startTime = time.Now()
}

// GetCurrentChunk returns the current chunk metrics (for debugging)
func (mc *MetricsCollector) GetCurrentChunk() ChunkMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	return mc.currentChunk
}

// GetHistorySize returns the number of chunks in history
func (mc *MetricsCollector) GetHistorySize() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	return len(mc.history)
}
