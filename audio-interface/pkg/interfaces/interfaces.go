package interfaces

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"time"
)

// AudioCapture defines the interface for audio capture operations
type AudioCapture interface {
	Initialize(config types.AudioConfig) error
	Start() error
	Stop() error
	GetFrameChannel() <-chan types.PCMFrame
	GetCaptureLatency() time.Duration
	Close() error
}

// AudioPlayback defines the interface for audio playback operations
type AudioPlayback interface {
	Initialize(config types.AudioConfig) error
	Start() error
	Stop() error
	WriteFrame(frame types.PCMFrame) error
	GetPlaybackLatency() time.Duration
	GetBufferFillLevel() float64
	Close() error
}

// StreamSynchronizer defines the interface for stream synchronization
type StreamSynchronizer interface {
	SyncCapturePlayback(captureTime, playbackTime time.Time) error
	GetDriftCompensation() time.Duration
	AdjustBufferSize(targetLatency time.Duration) error
}

// LatencyManager defines the interface for latency management
type LatencyManager interface {
	MonitorLatency() types.LatencyMetrics
	OptimizeBuffers(cpuLoad float64) error
	SelectOperationMode() (types.WASAPIMode, error)
}

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	RecordLatency(module string, latency time.Duration)
	RecordError(errInfo types.ErrorInfo)
	GetMetrics() types.LatencyMetrics
	Reset()
}
