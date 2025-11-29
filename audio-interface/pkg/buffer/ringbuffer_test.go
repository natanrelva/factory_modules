package buffer

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"testing"
	"time"
)

func TestNewRingBuffer(t *testing.T) {
	rb := NewRingBuffer(10)
	if rb.Capacity() != 10 {
		t.Errorf("Expected capacity 10, got %d", rb.Capacity())
	}
	if rb.Size() != 0 {
		t.Errorf("Expected size 0, got %d", rb.Size())
	}
}

func TestNewRingBuffer_InvalidCapacity(t *testing.T) {
	rb := NewRingBuffer(0)
	if rb.Capacity() != 10 {
		t.Errorf("Expected default capacity 10, got %d", rb.Capacity())
	}
}

func TestRingBuffer_WriteRead(t *testing.T) {
	rb := NewRingBuffer(5)
	
	frame := types.PCMFrame{
		Data:       []int16{1, 2, 3},
		SampleRate: 16000,
		Channels:   1,
		Timestamp:  time.Now(),
		Duration:   10 * time.Millisecond,
	}

	err := rb.Write(frame)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if rb.Size() != 1 {
		t.Errorf("Expected size 1, got %d", rb.Size())
	}

	readFrame, err := rb.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(readFrame.Data) != len(frame.Data) {
		t.Errorf("Expected data length %d, got %d", len(frame.Data), len(readFrame.Data))
	}

	if rb.Size() != 0 {
		t.Errorf("Expected size 0 after read, got %d", rb.Size())
	}
}

func TestRingBuffer_Overflow(t *testing.T) {
	rb := NewRingBuffer(3)
	
	frame := types.PCMFrame{
		Data:       []int16{1, 2, 3},
		SampleRate: 16000,
		Channels:   1,
		Duration:   10 * time.Millisecond,
	}

	// Fill buffer
	for i := 0; i < 3; i++ {
		err := rb.Write(frame)
		if err != nil {
			t.Fatalf("Write %d failed: %v", i, err)
		}
	}

	// Try to overflow
	err := rb.Write(frame)
	if err == nil {
		t.Error("Expected overflow error, got nil")
	}

	overruns, _ := rb.GetStats()
	if overruns != 1 {
		t.Errorf("Expected 1 overrun, got %d", overruns)
	}
}

func TestRingBuffer_Underflow(t *testing.T) {
	rb := NewRingBuffer(5)

	// Try to read from empty buffer
	_, err := rb.Read()
	if err == nil {
		t.Error("Expected underflow error, got nil")
	}

	_, underruns := rb.GetStats()
	if underruns != 1 {
		t.Errorf("Expected 1 underrun, got %d", underruns)
	}
}

func TestRingBuffer_FillLevel(t *testing.T) {
	rb := NewRingBuffer(10)
	
	frame := types.PCMFrame{
		Data:     []int16{1, 2, 3},
		Duration: 10 * time.Millisecond,
	}

	// Empty buffer
	if rb.FillLevel() != 0.0 {
		t.Errorf("Expected fill level 0.0, got %f", rb.FillLevel())
	}

	// Half full
	for i := 0; i < 5; i++ {
		rb.Write(frame)
	}
	if rb.FillLevel() != 0.5 {
		t.Errorf("Expected fill level 0.5, got %f", rb.FillLevel())
	}

	// Full
	for i := 0; i < 5; i++ {
		rb.Write(frame)
	}
	if rb.FillLevel() != 1.0 {
		t.Errorf("Expected fill level 1.0, got %f", rb.FillLevel())
	}
}

func TestRingBuffer_Clear(t *testing.T) {
	rb := NewRingBuffer(5)
	
	frame := types.PCMFrame{
		Data:     []int16{1, 2, 3},
		Duration: 10 * time.Millisecond,
	}

	for i := 0; i < 3; i++ {
		rb.Write(frame)
	}

	rb.Clear()

	if rb.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", rb.Size())
	}

	if !rb.IsEmpty() {
		t.Error("Expected buffer to be empty after clear")
	}
}

func TestRingBuffer_ConcurrentAccess(t *testing.T) {
	rb := NewRingBuffer(100)
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				frame := types.PCMFrame{
					Data:     []int16{int16(id), int16(j)},
					Duration: 10 * time.Millisecond,
				}
				rb.Write(frame)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				rb.Read()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()

	// No crashes = success
}

func TestRingBuffer_TryWriteRead(t *testing.T) {
	rb := NewRingBuffer(2)
	
	frame := types.PCMFrame{
		Data:     []int16{1, 2, 3},
		Duration: 10 * time.Millisecond,
	}

	// TryWrite success
	if !rb.TryWrite(frame) {
		t.Error("TryWrite should succeed on empty buffer")
	}

	// TryRead success
	_, ok := rb.TryRead()
	if !ok {
		t.Error("TryRead should succeed when data available")
	}

	// TryRead failure
	_, ok = rb.TryRead()
	if ok {
		t.Error("TryRead should fail on empty buffer")
	}

	// Fill buffer
	rb.TryWrite(frame)
	rb.TryWrite(frame)

	// TryWrite failure
	if rb.TryWrite(frame) {
		t.Error("TryWrite should fail on full buffer")
	}
}

func TestRingBuffer_IsEmptyIsFull(t *testing.T) {
	rb := NewRingBuffer(2)
	
	frame := types.PCMFrame{
		Data:     []int16{1, 2, 3},
		Duration: 10 * time.Millisecond,
	}

	if !rb.IsEmpty() {
		t.Error("New buffer should be empty")
	}

	rb.Write(frame)
	if rb.IsEmpty() {
		t.Error("Buffer should not be empty after write")
	}

	rb.Write(frame)
	if !rb.IsFull() {
		t.Error("Buffer should be full")
	}
}

func TestRingBuffer_ResetStats(t *testing.T) {
	rb := NewRingBuffer(1)
	
	frame := types.PCMFrame{
		Data:     []int16{1, 2, 3},
		Duration: 10 * time.Millisecond,
	}

	// Cause overflow
	rb.Write(frame)
	rb.Write(frame) // overflow

	// Cause underflow
	rb.Read()
	rb.Read() // underflow

	overruns, underruns := rb.GetStats()
	if overruns != 1 || underruns != 1 {
		t.Errorf("Expected 1 overrun and 1 underrun, got %d and %d", overruns, underruns)
	}

	rb.ResetStats()
	overruns, underruns = rb.GetStats()
	if overruns != 0 || underruns != 0 {
		t.Errorf("Expected stats reset to 0, got %d and %d", overruns, underruns)
	}
}
