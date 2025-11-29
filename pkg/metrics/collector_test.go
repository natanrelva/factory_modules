package metrics

import (
	"math"
	"testing"
	"testing/quick"
	"time"
)

// Unit Tests

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	
	if collector == nil {
		t.Fatal("NewMetricsCollector returned nil")
	}
	
	stats := collector.GetAggregated()
	if stats.TotalChunks != 0 {
		t.Errorf("Expected 0 total chunks, got %d", stats.TotalChunks)
	}
}

func TestRecordLatency(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Record some latencies
	collector.RecordLatency("ASR", 100*time.Millisecond)
	collector.RecordLatency("Translation", 200*time.Millisecond)
	collector.RecordLatency("TTS", 50*time.Millisecond)
	
	stats := collector.GetAggregated()
	
	if stats.TotalChunks != 1 {
		t.Errorf("Expected 1 total chunk, got %d", stats.TotalChunks)
	}
	
	// Check component latencies
	if stats.ComponentLatencies["ASR"] != 100*time.Millisecond {
		t.Errorf("Expected ASR latency 100ms, got %v", stats.ComponentLatencies["ASR"])
	}
	
	if stats.ComponentLatencies["Translation"] != 200*time.Millisecond {
		t.Errorf("Expected Translation latency 200ms, got %v", stats.ComponentLatencies["Translation"])
	}
	
	// Check total latency
	expectedTotal := 350 * time.Millisecond
	if stats.AverageLatency != expectedTotal {
		t.Errorf("Expected total latency %v, got %v", expectedTotal, stats.AverageLatency)
	}
}

func TestRecordCacheHitMiss(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Record cache operations
	collector.RecordCacheHit()
	collector.RecordCacheHit()
	collector.RecordCacheMiss()
	
	stats := collector.GetAggregated()
	
	if stats.CacheHits != 2 {
		t.Errorf("Expected 2 cache hits, got %d", stats.CacheHits)
	}
	
	if stats.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", stats.CacheMisses)
	}
	
	expectedHitRate := 2.0 / 3.0
	if math.Abs(stats.CacheHitRate-expectedHitRate) > 0.001 {
		t.Errorf("Expected cache hit rate %f, got %f", expectedHitRate, stats.CacheHitRate)
	}
	
	// Test with more operations
	collector.RecordCacheHit()
	collector.RecordCacheMiss()
	
	stats = collector.GetAggregated()
	expectedHitRate = 3.0 / 5.0
	if math.Abs(stats.CacheHitRate-expectedHitRate) > 0.001 {
		t.Errorf("Expected cache hit rate %f after more ops, got %f", expectedHitRate, stats.CacheHitRate)
	}
}

func TestRecordSilenceSkip(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Record silence skips
	collector.RecordSilenceSkip()
	collector.RecordSilenceSkip()
	
	stats := collector.GetAggregated()
	
	if stats.SilenceSkips != 2 {
		t.Errorf("Expected 2 silence skips, got %d", stats.SilenceSkips)
	}
}

func TestMultipleChunks(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Process multiple chunks
	for i := 0; i < 5; i++ {
		collector.RecordLatency("ASR", time.Duration(100+i*10)*time.Millisecond)
		collector.RecordLatency("Translation", time.Duration(200+i*20)*time.Millisecond)
		collector.RecordLatency("TTS", time.Duration(50+i*5)*time.Millisecond)
		collector.NextChunk()
	}
	
	stats := collector.GetAggregated()
	
	if stats.TotalChunks != 5 {
		t.Errorf("Expected 5 total chunks, got %d", stats.TotalChunks)
	}
	
	// Check average latencies
	expectedASR := 120 * time.Millisecond // (100+110+120+130+140)/5
	if stats.ComponentLatencies["ASR"] != expectedASR {
		t.Errorf("Expected average ASR latency %v, got %v", expectedASR, stats.ComponentLatencies["ASR"])
	}
}

func TestPercentiles(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Add latencies: 100, 200, 300, 400, 500 ms
	for i := 1; i <= 5; i++ {
		collector.RecordLatency("Total", time.Duration(i*100)*time.Millisecond)
		collector.NextChunk()
	}
	
	stats := collector.GetAggregated()
	
	// P50 should be 300ms (middle value, index 2)
	if stats.P50Latency != 300*time.Millisecond {
		t.Errorf("Expected P50 300ms, got %v", stats.P50Latency)
	}
	
	// P95 of 5 values: 0.95 * 4 = 3.8 -> index 3 = 400ms
	if stats.P95Latency != 400*time.Millisecond {
		t.Errorf("Expected P95 400ms, got %v", stats.P95Latency)
	}
	
	// P99 of 5 values: 0.99 * 4 = 3.96 -> index 3 = 400ms
	if stats.P99Latency != 400*time.Millisecond {
		t.Errorf("Expected P99 400ms, got %v", stats.P99Latency)
	}
}

func TestReset(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Add some data
	collector.RecordLatency("ASR", 100*time.Millisecond)
	collector.RecordCacheHit()
	collector.RecordSilenceSkip()
	
	// Reset
	collector.Reset()
	
	stats := collector.GetAggregated()
	if stats.TotalChunks != 0 {
		t.Errorf("Expected 0 total chunks after reset, got %d", stats.TotalChunks)
	}
	
	if stats.CacheHits != 0 {
		t.Errorf("Expected 0 cache hits after reset, got %d", stats.CacheHits)
	}
	
	if stats.SilenceSkips != 0 {
		t.Errorf("Expected 0 silence skips after reset, got %d", stats.SilenceSkips)
	}
}

func TestMaxHistorySize(t *testing.T) {
	collector := NewMetricsCollector()
	
	// Add more than max history (default 100)
	for i := 0; i < 150; i++ {
		collector.RecordLatency("ASR", time.Duration(i)*time.Millisecond)
		collector.NextChunk()
	}
	
	stats := collector.GetAggregated()
	
	// Should only keep last 100
	if stats.TotalChunks != 100 {
		t.Errorf("Expected 100 total chunks (max history), got %d", stats.TotalChunks)
	}
}

// Property-Based Tests

// Property 8: Metrics Accuracy
// For any sequence of recorded latencies, all should be preserved (up to max history)
func TestProperty_MetricsAccuracy(t *testing.T) {
	f := func(latencies []int) bool {
		if len(latencies) == 0 {
			return true
		}
		
		collector := NewMetricsCollector()
		
		// Record latencies
		for _, lat := range latencies {
			if lat < 0 {
				lat = -lat
			}
			if lat > 10000 {
				lat = 10000 // Cap at 10 seconds
			}
			
			collector.RecordLatency("Test", time.Duration(lat)*time.Millisecond)
			collector.NextChunk()
		}
		
		stats := collector.GetAggregated()
		
		// Should have recorded all latencies (up to max history)
		expectedChunks := len(latencies)
		if expectedChunks > 100 {
			expectedChunks = 100 // Max history
		}
		
		return int(stats.TotalChunks) == expectedChunks
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property: Cache Hit Rate Calculation
// For any sequence of cache operations, hit rate should be hits / (hits + misses)
func TestProperty_CacheHitRateCalculation(t *testing.T) {
	f := func(operations []bool) bool {
		if len(operations) == 0 {
			return true
		}
		
		collector := NewMetricsCollector()
		
		hits := 0
		misses := 0
		
		// Record operations (true = hit, false = miss)
		for _, isHit := range operations {
			if isHit {
				collector.RecordCacheHit()
				hits++
			} else {
				collector.RecordCacheMiss()
				misses++
			}
		}
		
		stats := collector.GetAggregated()
		
		// Verify counts
		if int(stats.CacheHits) != hits || int(stats.CacheMisses) != misses {
			return false
		}
		
		// Verify hit rate
		if hits+misses == 0 {
			return stats.CacheHitRate == 0
		}
		
		expectedHitRate := float64(hits) / float64(hits+misses)
		diff := math.Abs(stats.CacheHitRate - expectedHitRate)
		return diff < 0.001
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property: Average Latency Calculation
// For any sequence of latencies, average should be sum / count
func TestProperty_AverageLatencyCalculation(t *testing.T) {
	f := func(latencies []int) bool {
		if len(latencies) == 0 {
			return true
		}
		
		collector := NewMetricsCollector()
		
		sum := 0
		for _, lat := range latencies {
			if lat < 0 {
				lat = -lat
			}
			if lat > 10000 {
				lat = 10000
			}
			
			collector.RecordLatency("Test", time.Duration(lat)*time.Millisecond)
			collector.NextChunk()
			sum += lat
		}
		
		stats := collector.GetAggregated()
		
		// Calculate expected average (considering max history)
		count := len(latencies)
		if count > 100 {
			// Only last 100 are kept
			sum = 0
			for i := len(latencies) - 100; i < len(latencies); i++ {
				lat := latencies[i]
				if lat < 0 {
					lat = -lat
				}
				if lat > 10000 {
					lat = 10000
				}
				sum += lat
			}
			count = 100
		}
		
		expectedAvg := time.Duration(sum/count) * time.Millisecond
		
		// Allow some tolerance for rounding
		diff := int64(stats.ComponentLatencies["Test"]) - int64(expectedAvg)
		if diff < 0 {
			diff = -diff
		}
		
		return diff < int64(time.Millisecond)
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 30}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Benchmark Tests

func BenchmarkRecordLatency(b *testing.B) {
	collector := NewMetricsCollector()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordLatency("ASR", 100*time.Millisecond)
	}
}

func BenchmarkGetAggregated(b *testing.B) {
	collector := NewMetricsCollector()
	
	// Add some data
	for i := 0; i < 100; i++ {
		collector.RecordLatency("ASR", time.Duration(i)*time.Millisecond)
		collector.RecordLatency("Translation", time.Duration(i*2)*time.Millisecond)
		collector.NextChunk()
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.GetAggregated()
	}
}

func BenchmarkConcurrentRecording(b *testing.B) {
	collector := NewMetricsCollector()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.RecordLatency("ASR", 100*time.Millisecond)
			collector.RecordCacheHit()
		}
	})
}
