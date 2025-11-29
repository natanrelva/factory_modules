# Audio Interface Module (Windows)

Low-latency audio capture and playback interface for Windows using WASAPI.

## Project Structure

```
audio-interface/
├── pkg/
│   ├── types/          # Core data types
│   ├── interfaces/     # Interface definitions
│   ├── mocks/          # Mock implementations for testing
│   ├── buffer/         # Ring buffer implementation
│   ├── capture/        # Audio capture (M6.1.1-Win)
│   ├── playback/       # Audio playback (M6.1.2-Win)
│   ├── sync/           # Stream synchronization (M6.1.3-Win)
│   ├── latency/        # Latency management (M6.1.4-Win)
│   └── metrics/        # Metrics collection
├── cmd/
│   └── example/        # Example applications
└── go.mod
```

## Core Types

### PCMFrame
Represents a chunk of audio data in PCM format (10-20ms).

### AudioConfig
Configuration for audio capture/playback operations.

### LatencyMetrics
Performance metrics tracking for the audio pipeline.

## Interfaces

- **AudioCapture**: Audio capture operations
- **AudioPlayback**: Audio playback operations
- **StreamSynchronizer**: Stream synchronization
- **LatencyManager**: Latency optimization
- **MetricsCollector**: Metrics collection

## Requirements

- Go 1.21+
- Windows 10/11
- WASAPI support

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run property-based tests
go test -v ./pkg/buffer/...
```

## Latency Budget

- Capture: ≤30ms
- Playback: ≤50ms
- Total I/O: ≤80ms (contributes to 700ms end-to-end system budget)

## License

MIT
