package types

import (
	"fmt"
	"time"
)

// PCMFrame represents a chunk of audio data in PCM format
type PCMFrame struct {
	Data       []int16       // 16-bit PCM samples
	SampleRate int           // Samples per second (e.g., 16000, 48000)
	Channels   int           // Mono = 1, Stereo = 2
	Timestamp  time.Time     // Capture or generation timestamp
	Duration   time.Duration // Frame duration (typically 10-20ms)
	IsSpeech   bool          // VAD result (for future integration)
}

// AudioConfig holds configuration for audio capture/playback
type AudioConfig struct {
	DeviceID   string // Device identifier (empty for default)
	SampleRate int    // Samples per second
	Channels   int    // Number of audio channels
	FrameSize  int    // Samples per frame
	BufferSize int    // Buffer size in frames
}

// LatencyMetrics tracks performance metrics for the audio pipeline
type LatencyMetrics struct {
	CaptureLatency  time.Duration // Time from mic to frame delivery
	PlaybackLatency time.Duration // Time from frame to speaker
	BufferFillLevel float64       // 0.0 - 1.0, current buffer utilization
	DroppedFrames   int           // Count of frames lost
	Underruns       int           // Count of buffer underruns
	Overruns        int           // Count of buffer overruns
	Timestamp       time.Time     // When metrics were captured
}

// WASAPIMode represents the WASAPI operation mode
type WASAPIMode int

const (
	// Shared mode allows multiple applications to share the audio device
	Shared WASAPIMode = iota
	// Exclusive mode gives exclusive access for lower latency
	Exclusive
)

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Module    string    // Module where error occurred
	Operation string    // Operation that failed
	Err       error     // The actual error
	Timestamp time.Time // When error occurred
	Context   string    // Additional context
}

// Error implements the error interface
func (e *ErrorInfo) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s.%s: %v (context: %s)", e.Module, e.Operation, e.Err, e.Context)
	}
	return fmt.Sprintf("%s.%s failed (context: %s)", e.Module, e.Operation, e.Context)
}

// Unwrap returns the underlying error
func (e *ErrorInfo) Unwrap() error {
	return e.Err
}
