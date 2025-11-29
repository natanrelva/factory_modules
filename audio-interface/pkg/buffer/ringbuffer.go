package buffer

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
)

// RingBuffer is a thread-safe circular buffer for PCM frames
type RingBuffer struct {
	mu       sync.RWMutex
	buffer   []types.PCMFrame
	capacity int
	head     int // Write position
	tail     int // Read position
	size     int // Current number of elements
	overruns int // Count of write overflows
	underruns int // Count of read underflows
}

// NewRingBuffer creates a new ring buffer with the specified capacity
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 10 // Default capacity
	}
	return &RingBuffer{
		buffer:   make([]types.PCMFrame, capacity),
		capacity: capacity,
		head:     0,
		tail:     0,
		size:     0,
	}
}

// Write adds a frame to the buffer
// Returns error if buffer is full (overflow protection)
func (rb *RingBuffer) Write(frame types.PCMFrame) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.size >= rb.capacity {
		rb.overruns++
		return fmt.Errorf("buffer overflow: capacity %d reached", rb.capacity)
	}

	rb.buffer[rb.head] = frame
	rb.head = (rb.head + 1) % rb.capacity
	rb.size++

	return nil
}

// Read retrieves a frame from the buffer
// Returns error if buffer is empty (underflow protection)
func (rb *RingBuffer) Read() (types.PCMFrame, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.size == 0 {
		rb.underruns++
		return types.PCMFrame{}, fmt.Errorf("buffer underflow: no data available")
	}

	frame := rb.buffer[rb.tail]
	rb.tail = (rb.tail + 1) % rb.capacity
	rb.size--

	return frame, nil
}

// TryWrite attempts to write without blocking
// Returns false if buffer is full
func (rb *RingBuffer) TryWrite(frame types.PCMFrame) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.size >= rb.capacity {
		rb.overruns++
		return false
	}

	rb.buffer[rb.head] = frame
	rb.head = (rb.head + 1) % rb.capacity
	rb.size++

	return true
}

// TryRead attempts to read without blocking
// Returns false if buffer is empty
func (rb *RingBuffer) TryRead() (types.PCMFrame, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.size == 0 {
		rb.underruns++
		return types.PCMFrame{}, false
	}

	frame := rb.buffer[rb.tail]
	rb.tail = (rb.tail + 1) % rb.capacity
	rb.size--

	return frame, true
}

// Size returns the current number of frames in the buffer
func (rb *RingBuffer) Size() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size
}

// Capacity returns the maximum capacity of the buffer
func (rb *RingBuffer) Capacity() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.capacity
}

// FillLevel returns the buffer fill level as a percentage (0.0 - 1.0)
func (rb *RingBuffer) FillLevel() float64 {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	if rb.capacity == 0 {
		return 0.0
	}
	return float64(rb.size) / float64(rb.capacity)
}

// IsEmpty returns true if the buffer is empty
func (rb *RingBuffer) IsEmpty() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size == 0
}

// IsFull returns true if the buffer is full
func (rb *RingBuffer) IsFull() bool {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.size >= rb.capacity
}

// Clear removes all frames from the buffer
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.head = 0
	rb.tail = 0
	rb.size = 0
}

// GetStats returns buffer statistics
func (rb *RingBuffer) GetStats() (overruns, underruns int) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.overruns, rb.underruns
}

// ResetStats resets the overflow and underflow counters
func (rb *RingBuffer) ResetStats() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.overruns = 0
	rb.underruns = 0
}
