# Design Document - Google Meets Integration

## Overview

This design document describes the architecture and implementation approach for integrating the real-time dubbing system with Google Meets and other video conferencing platforms. The key challenge is routing translated audio through a virtual audio cable while maintaining low latency and high quality.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        User's Computer                           │
│                                                                  │
│  ┌──────────────┐                                               │
│  │   Physical   │                                               │
│  │  Microphone  │                                               │
│  └──────┬───────┘                                               │
│         │ Portuguese Speech                                     │
│         ↓                                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Dubbing System Pipeline                      │  │
│  │                                                            │  │
│  │  Audio Capture → Silence Detection → ASR (Vosk)          │  │
│  │       ↓                                                    │  │
│  │  Translation Cache Check                                  │  │
│  │       ↓                                                    │  │
│  │  Translation (Argos) → TTS (Windows) → Audio Playback    │  │
│  └────────────────────────────────────────┬─────────────────┘  │
│                                            │ English Audio      │
│                                            ↓                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │         Virtual Audio Cable (VB-Audio)                   │  │
│  │                                                            │  │
│  │  CABLE Input ← [Audio Stream] → CABLE Output            │  │
│  └────────────────────────────────────────┬─────────────────┘  │
│                                            │                    │
│                                            ↓                    │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Google Meets (Browser)                       │  │
│  │                                                            │  │
│  │  Microphone: CABLE Output                                │  │
│  │  Transmits English audio to other participants           │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Audio Output Manager                      │
│                                                              │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────┐ │
│  │ Device Manager │  │ Playback Queue │  │ Audio Writer │ │
│  │                │  │                │  │              │ │
│  │ - Enumerate    │  │ - Order chunks │  │ - Write PCM  │ │
│  │ - Validate     │  │ - Buffer mgmt  │  │ - Handle     │ │
│  │ - Hot-plug     │  │ - Backpressure │  │   errors     │ │
│  └────────────────┘  └────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Audio Output Manager

**Purpose**: Manages audio playback to virtual audio cable or system output.

**Interface**:
```go
type AudioOutputManager interface {
    // Initialize output with device name
    Initialize(deviceName string) error
    
    // Play audio samples to output device
    PlayAudio(samples []float32) error
    
    // Queue audio for ordered playback
    QueueAudio(chunkID int, samples []float32) error
    
    // Get list of available output devices
    ListDevices() ([]AudioDevice, error)
    
    // Get current device info
    GetCurrentDevice() AudioDevice
    
    // Close and cleanup
    Close() error
}

type AudioDevice struct {
    Name        string
    ID          string
    IsDefault   bool
    IsVirtual   bool
    SampleRate  int
    Channels    int
}
```

**Implementation**:
- Use PortAudio or similar library for cross-platform audio output
- Maintain playback queue for ordered chunk delivery
- Handle device enumeration and hot-plugging
- Implement buffering strategy to prevent underruns

### 2. Device Manager

**Purpose**: Enumerate, validate, and manage audio devices.

**Interface**:
```go
type DeviceManager interface {
    // Enumerate all audio devices
    EnumerateDevices() ([]AudioDevice, error)
    
    // Find device by name or ID
    FindDevice(nameOrID string) (*AudioDevice, error)
    
    // Detect virtual audio cables
    DetectVirtualCables() ([]AudioDevice, error)
    
    // Validate device is accessible
    ValidateDevice(device AudioDevice) error
    
    // Watch for device changes
    WatchDevices(callback func(event DeviceEvent)) error
}

type DeviceEvent struct {
    Type   string // "added", "removed", "changed"
    Device AudioDevice
}
```

### 3. Playback Queue

**Purpose**: Ensure audio chunks play in correct order with proper buffering.

**Interface**:
```go
type PlaybackQueue interface {
    // Enqueue audio chunk with ID for ordering
    Enqueue(chunkID int, samples []float32) error
    
    // Dequeue next chunk in order
    Dequeue() ([]float32, error)
    
    // Get queue status
    GetStatus() QueueStatus
    
    // Clear queue
    Clear() error
}

type QueueStatus struct {
    Size          int
    NextExpected  int
    OldestChunkID int
    NewestChunkID int
    IsBlocked     bool
}
```

### 4. Audio Writer

**Purpose**: Low-level audio writing to output device.

**Interface**:
```go
type AudioWriter interface {
    // Open audio stream
    Open(device AudioDevice, sampleRate int, channels int) error
    
    // Write audio samples
    Write(samples []float32) error
    
    // Get buffer status
    GetBufferStatus() BufferStatus
    
    // Close stream
    Close() error
}

type BufferStatus struct {
    Available int
    Used      int
    Underruns int
    Overruns  int
}
```

## Data Models

### Audio Configuration

```go
type AudioConfig struct {
    InputDevice   string
    OutputDevice  string
    SampleRate    int
    Channels      int
    BufferSize    int
    LatencyMode   string // "low", "balanced", "high"
}
```

### Playback Metrics

```go
type PlaybackMetrics struct {
    ChunksPlayed    int64
    BytesWritten    int64
    Underruns       int64
    Overruns        int64
    AverageLatency  time.Duration
    CurrentDevice   string
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Audio Chunk Ordering
*For any* sequence of audio chunks with IDs, playback SHALL occur in ascending ID order regardless of arrival order.
**Validates: Requirements 7.1, 7.2**

### Property 2: Device Enumeration Consistency
*For any* system state, enumerating devices twice SHALL return the same devices (unless hardware changed).
**Validates: Requirements 1.1, 5.1**

### Property 3: Playback Continuity
*For any* continuous stream of audio chunks, gaps in playback SHALL be < 100ms.
**Validates: Requirements 2.2, 6.4**

### Property 4: Latency Budget Compliance
*For any* audio chunk in low-latency mode, end-to-end latency SHALL be < 3 seconds.
**Validates: Requirements 3.1**

### Property 5: Buffer Management
*For any* playback queue state, buffer usage SHALL not exceed configured maximum.
**Validates: Requirements 2.5, 8.4**

### Property 6: Device Validation
*For any* device name provided by user, validation SHALL correctly identify if device exists and is accessible.
**Validates: Requirements 1.3, 5.3**

### Property 7: Audio Quality Preservation
*For any* audio samples, playback SHALL maintain sample rate and bit depth without degradation.
**Validates: Requirements 4.1, 4.2, 12.1**

### Property 8: Error Recovery
*For any* recoverable error, the system SHALL resume normal operation within 5 seconds.
**Validates: Requirements 9.1, 9.5**

### Property 9: Resource Bounds
*For any* runtime duration, memory usage SHALL not exceed 500MB and CPU SHALL average < 50%.
**Validates: Requirements 8.1, 8.3**

### Property 10: Hot-Plug Handling
*For any* device removal event, the system SHALL detect it within 2 seconds and fallback gracefully.
**Validates: Requirements 11.1, 11.3**

## Error Handling

### Error Categories

1. **Device Errors**
   - Device not found
   - Device access denied
   - Device disconnected
   - Device format not supported

2. **Playback Errors**
   - Buffer underrun
   - Buffer overrun
   - Write failure
   - Stream interrupted

3. **Queue Errors**
   - Queue full
   - Out of order chunks
   - Missing chunks
   - Queue timeout

### Error Recovery Strategies

1. **Device Not Found**
   - Log error with available devices
   - Fallback to default device
   - Continue with fallback

2. **Buffer Underrun**
   - Log warning
   - Insert silence to maintain stream
   - Adjust buffer size if frequent

3. **Device Disconnected**
   - Detect disconnection
   - Attempt reconnection (3 retries)
   - Fallback to default device
   - Log event

4. **Queue Full**
   - Apply backpressure to pipeline
   - Drop oldest chunk if critical
   - Log warning

## Testing Strategy

### Unit Tests

1. **Device Manager Tests**
   - Test device enumeration
   - Test device validation
   - Test virtual cable detection
   - Test device not found handling

2. **Playback Queue Tests**
   - Test in-order playback
   - Test out-of-order arrival
   - Test queue full handling
   - Test queue clear

3. **Audio Writer Tests**
   - Test audio writing
   - Test buffer management
   - Test error handling
   - Test stream lifecycle

### Property-Based Tests

1. **Chunk Ordering Property**
   - Generate random chunk sequences
   - Verify playback order matches input order
   - Test with various arrival patterns

2. **Buffer Management Property**
   - Generate random audio streams
   - Verify buffer never exceeds limits
   - Test with various chunk sizes

3. **Latency Property**
   - Generate random audio chunks
   - Measure end-to-end latency
   - Verify < 3s in low-latency mode

4. **Device Validation Property**
   - Generate random device names
   - Verify validation correctness
   - Test with valid and invalid devices

### Integration Tests

1. **Virtual Cable Integration**
   - Test with VB-Audio Cable
   - Verify audio routing
   - Test with Google Meets

2. **End-to-End Latency**
   - Measure complete pipeline latency
   - Verify meets requirements
   - Test with different modes

3. **Device Hot-Plugging**
   - Simulate device removal
   - Verify graceful fallback
   - Test recovery

### Performance Tests

1. **Throughput Test**
   - Measure chunks per second
   - Verify no bottlenecks
   - Test with parallel processing

2. **Resource Usage Test**
   - Monitor CPU usage
   - Monitor memory usage
   - Verify within limits

3. **Stress Test**
   - Run for extended duration (1+ hour)
   - Verify no degradation
   - Check for memory leaks

## Implementation Notes

### Audio Library Selection

**Recommended**: PortAudio
- Cross-platform (Windows, macOS, Linux)
- Low-level control
- Good documentation
- Active maintenance

**Alternative**: miniaudio
- Header-only library
- Simpler API
- Good for basic use cases

### Virtual Cable Detection

Detect virtual cables by:
1. Device name patterns ("CABLE", "Virtual", "VB-Audio")
2. Device properties (loopback capability)
3. Known virtual cable drivers

### Buffer Size Tuning

- **Low-latency mode**: 512-1024 samples
- **Balanced mode**: 2048 samples
- **Quality mode**: 4096 samples

Adjust based on:
- CPU performance
- Underrun frequency
- Latency requirements

### Playback Queue Strategy

Use priority queue with:
- Chunk ID as priority
- Timeout for missing chunks (5s)
- Maximum queue size (100 chunks)
- Backpressure when full

## Performance Targets

| Metric | Low-Latency | Balanced | Quality |
|--------|-------------|----------|---------|
| End-to-end latency | < 3s | < 4s | < 5s |
| CPU usage (avg) | < 50% | < 40% | < 30% |
| Memory usage | < 500MB | < 400MB | < 300MB |
| Buffer underruns | < 1% | < 0.5% | < 0.1% |
| Chunk ordering errors | 0 | 0 | 0 |

## Security Considerations

1. **Device Access**
   - Validate device permissions
   - Handle access denied gracefully
   - Don't expose sensitive device info

2. **Audio Data**
   - Don't log audio samples
   - Clear buffers on close
   - Respect user privacy

3. **Configuration**
   - Validate all user inputs
   - Sanitize device names
   - Prevent injection attacks
