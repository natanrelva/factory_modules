package pipeline

import (
	"errors"
	"sync"
	"testing"
	"testing/quick"
	"time"
)

// Unit Tests

func TestNewPipelineManager(t *testing.T) {
	manager := NewPipelineManager(3, 10)
	
	if manager == nil {
		t.Fatal("NewPipelineManager returned nil")
	}
	
	if manager.numWorkers != 3 {
		t.Errorf("Expected 3 workers, got %d", manager.numWorkers)
	}
	
	if manager.queueSize != 10 {
		t.Errorf("Expected queue size 10, got %d", manager.queueSize)
	}
}

func TestStartStop(t *testing.T) {
	manager := NewPipelineManager(2, 5)
	
	// Start
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	
	// Stop
	err = manager.Stop()
	if err != nil {
		t.Fatalf("Failed to stop manager: %v", err)
	}
}

func TestProcessChunk_Simple(t *testing.T) {
	manager := NewPipelineManager(2, 5)
	defer manager.Stop()
	
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	
	// Simple processor that doubles the input
	processor := func(input interface{}) (interface{}, error) {
		n := input.(int)
		return n * 2, nil
	}
	
	// Process a chunk
	result, err := manager.ProcessChunk(1, 42, processor)
	if err != nil {
		t.Fatalf("ProcessChunk failed: %v", err)
	}
	
	if result.(int) != 84 {
		t.Errorf("Expected 84, got %v", result)
	}
}

func TestProcessChunk_OrderPreservation(t *testing.T) {
	manager := NewPipelineManager(3, 10)
	defer manager.Stop()
	
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	
	// Processor that adds a delay to simulate work
	processor := func(input interface{}) (interface{}, error) {
		n := input.(int)
		// Add variable delay to test ordering
		time.Sleep(time.Duration(10-n) * time.Millisecond)
		return n * 2, nil
	}
	
	// Process multiple chunks in parallel
	numChunks := 10
	results := make([]interface{}, numChunks)
	var wg sync.WaitGroup
	
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(chunkID int) {
			defer wg.Done()
			result, err := manager.ProcessChunk(chunkID, chunkID, processor)
			if err != nil {
				t.Errorf("ProcessChunk %d failed: %v", chunkID, err)
				return
			}
			results[chunkID] = result
		}(i)
	}
	
	wg.Wait()
	
	// Verify order is preserved
	for i := 0; i < numChunks; i++ {
		expected := i * 2
		if results[i].(int) != expected {
			t.Errorf("Chunk %d: expected %d, got %v", i, expected, results[i])
		}
	}
}

func TestProcessChunk_Error(t *testing.T) {
	manager := NewPipelineManager(2, 5)
	defer manager.Stop()
	
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	
	// Processor that returns an error
	processor := func(input interface{}) (interface{}, error) {
		return nil, errors.New("processing error")
	}
	
	// Process a chunk
	_, err = manager.ProcessChunk(1, 42, processor)
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestProcessChunk_Panic(t *testing.T) {
	manager := NewPipelineManager(2, 5)
	defer manager.Stop()
	
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	
	// Processor that panics
	processor := func(input interface{}) (interface{}, error) {
		panic("test panic")
	}
	
	// Process a chunk - should recover from panic
	_, err = manager.ProcessChunk(1, 42, processor)
	if err == nil {
		t.Error("Expected error from panic, got nil")
	}
}

func TestGetStats(t *testing.T) {
	manager := NewPipelineManager(2, 5)
	defer manager.Stop()
	
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	
	// Process some chunks
	processor := func(input interface{}) (interface{}, error) {
		return input, nil
	}
	
	for i := 0; i < 5; i++ {
		_, err := manager.ProcessChunk(i, i, processor)
		if err != nil {
			t.Fatalf("ProcessChunk failed: %v", err)
		}
	}
	
	stats := manager.GetStats()
	
	if stats.ChunksProcessed != 5 {
		t.Errorf("Expected 5 chunks processed, got %d", stats.ChunksProcessed)
	}
	
	if stats.NumWorkers != 2 {
		t.Errorf("Expected 2 workers, got %d", stats.NumWorkers)
	}
}

func TestConcurrentProcessing(t *testing.T) {
	manager := NewPipelineManager(4, 20)
	defer manager.Stop()
	
	err := manager.Start()
	if err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	
	// Processor that simulates work
	processor := func(input interface{}) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return input, nil
	}
	
	// Process many chunks concurrently
	numChunks := 20
	var wg sync.WaitGroup
	errors := make(chan error, numChunks)
	
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(chunkID int) {
			defer wg.Done()
			_, err := manager.ProcessChunk(chunkID, chunkID, processor)
			if err != nil {
				errors <- err
			}
		}(i)
	}
	
	wg.Wait()
	close(errors)
	
	// Check for errors
	for err := range errors {
		t.Errorf("ProcessChunk failed: %v", err)
	}
	
	stats := manager.GetStats()
	if stats.ChunksProcessed != int64(numChunks) {
		t.Errorf("Expected %d chunks processed, got %d", numChunks, stats.ChunksProcessed)
	}
}

// Property-Based Tests

// Property 2: Parallel Processing Order
// For any sequence of chunks, output order SHALL match input order
func TestProperty_OrderPreservation(t *testing.T) {
	f := func(inputs []int) bool {
		if len(inputs) == 0 || len(inputs) > 50 {
			return true // Skip empty or very large inputs
		}
		
		manager := NewPipelineManager(3, len(inputs)+10)
		defer manager.Stop()
		
		if err := manager.Start(); err != nil {
			return false
		}
		
		// Processor with variable delay
		processor := func(input interface{}) (interface{}, error) {
			n := input.(int)
			// Add small random delay
			delay := (n % 10) + 1
			time.Sleep(time.Duration(delay) * time.Millisecond)
			return n, nil
		}
		
		// Process all chunks
		results := make([]interface{}, len(inputs))
		var wg sync.WaitGroup
		
		for i, input := range inputs {
			wg.Add(1)
			go func(chunkID, value int) {
				defer wg.Done()
				result, err := manager.ProcessChunk(chunkID, value, processor)
				if err != nil {
					return
				}
				results[chunkID] = result
			}(i, input)
		}
		
		wg.Wait()
		
		// Verify order
		for i, input := range inputs {
			if results[i] == nil {
				return false
			}
			if results[i].(int) != input {
				return false
			}
		}
		
		return true
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 20}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property: Parallel Throughput Gain
// For any workload, processing with N workers should be faster than 1 worker
func TestProperty_ParallelThroughput(t *testing.T) {
	// This test is more of a benchmark, but we can verify basic throughput
	numChunks := 20
	workDelay := 10 * time.Millisecond
	
	processor := func(input interface{}) (interface{}, error) {
		time.Sleep(workDelay)
		return input, nil
	}
	
	// Test with 1 worker
	manager1 := NewPipelineManager(1, numChunks+5)
	defer manager1.Stop()
	manager1.Start()
	
	start1 := time.Now()
	for i := 0; i < numChunks; i++ {
		manager1.ProcessChunk(i, i, processor)
	}
	duration1 := time.Since(start1)
	
	// Test with 4 workers
	manager4 := NewPipelineManager(4, numChunks+5)
	defer manager4.Stop()
	manager4.Start()
	
	start4 := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		go func(chunkID int) {
			defer wg.Done()
			manager4.ProcessChunk(chunkID, chunkID, processor)
		}(i)
	}
	wg.Wait()
	duration4 := time.Since(start4)
	
	// 4 workers should be significantly faster (at least 2x)
	if duration4 >= duration1/2 {
		t.Logf("1 worker: %v, 4 workers: %v", duration1, duration4)
		t.Error("Expected significant speedup with 4 workers")
	}
}

// Benchmark Tests

func BenchmarkProcessChunk_Sequential(b *testing.B) {
	manager := NewPipelineManager(1, 100)
	defer manager.Stop()
	manager.Start()
	
	processor := func(input interface{}) (interface{}, error) {
		return input, nil
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.ProcessChunk(i, i, processor)
	}
}

func BenchmarkProcessChunk_Parallel(b *testing.B) {
	manager := NewPipelineManager(4, 100)
	defer manager.Stop()
	manager.Start()
	
	processor := func(input interface{}) (interface{}, error) {
		return input, nil
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			manager.ProcessChunk(i, i, processor)
			i++
		}
	})
}
