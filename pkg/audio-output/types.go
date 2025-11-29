package audiooutput

import "time"

// AudioDevice represents an audio output device
type AudioDevice struct {
	Name       string
	ID         string
	IsDefault  bool
	IsVirtual  bool
	SampleRate int
	Channels   int
}

// DeviceEvent represents a device change event
type DeviceEvent struct {
	Type   string // "added", "removed", "changed"
	Device AudioDevice
}

// QueueStatus represents the status of the playback queue
type QueueStatus struct {
	Size          int
	NextExpected  int
	OldestChunkID int
	NewestChunkID int
	IsBlocked     bool
}

// BufferStatus represents the status of audio buffers
type BufferStatus struct {
	Available int
	Used      int
	Underruns int
	Overruns  int
}

// PlaybackMetrics holds playback statistics
type PlaybackMetrics struct {
	ChunksPlayed   int64
	BytesWritten   int64
	Underruns      int64
	Overruns       int64
	AverageLatency time.Duration
	CurrentDevice  string
}

// AudioConfig holds audio configuration
type AudioConfig struct {
	InputDevice  string
	OutputDevice string
	SampleRate   int
	Channels     int
	BufferSize   int
	LatencyMode  string // "low", "balanced", "high"
}
