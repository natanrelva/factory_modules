package mocks

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// MockAudioCapture is a mock implementation of AudioCapture for testing
type MockAudioCapture struct {
	mu            sync.Mutex
	initialized   bool
	running       bool
	frameChannel  chan types.PCMFrame
	latency       time.Duration
	shouldError   bool
	errorOnMethod string
}

func NewMockAudioCapture() *MockAudioCapture {
	return &MockAudioCapture{
		frameChannel: make(chan types.PCMFrame, 10),
		latency:      20 * time.Millisecond,
	}
}

func (m *MockAudioCapture) Initialize(config types.AudioConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shouldError && m.errorOnMethod == "Initialize" {
		return &types.ErrorInfo{Module: "MockCapture", Operation: "Initialize"}
	}
	m.initialized = true
	return nil
}

func (m *MockAudioCapture) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.initialized {
		return &types.ErrorInfo{Module: "MockCapture", Operation: "Start"}
	}
	m.running = true
	return nil
}

func (m *MockAudioCapture) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

func (m *MockAudioCapture) GetFrameChannel() <-chan types.PCMFrame {
	return m.frameChannel
}

func (m *MockAudioCapture) GetCaptureLatency() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.latency
}

func (m *MockAudioCapture) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	close(m.frameChannel)
	m.initialized = false
	m.running = false
	return nil
}

// Helper methods for testing
func (m *MockAudioCapture) SendFrame(frame types.PCMFrame) {
	m.frameChannel <- frame
}

func (m *MockAudioCapture) SetError(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = true
	m.errorOnMethod = method
}

// MockAudioPlayback is a mock implementation of AudioPlayback for testing
type MockAudioPlayback struct {
	mu              sync.Mutex
	initialized     bool
	running         bool
	receivedFrames  []types.PCMFrame
	latency         time.Duration
	bufferFillLevel float64
	shouldError     bool
	errorOnMethod   string
}

func NewMockAudioPlayback() *MockAudioPlayback {
	return &MockAudioPlayback{
		receivedFrames:  make([]types.PCMFrame, 0),
		latency:         30 * time.Millisecond,
		bufferFillLevel: 0.5,
	}
}

func (m *MockAudioPlayback) Initialize(config types.AudioConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.shouldError && m.errorOnMethod == "Initialize" {
		return &types.ErrorInfo{Module: "MockPlayback", Operation: "Initialize"}
	}
	m.initialized = true
	return nil
}

func (m *MockAudioPlayback) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.initialized {
		return &types.ErrorInfo{Module: "MockPlayback", Operation: "Start"}
	}
	m.running = true
	return nil
}

func (m *MockAudioPlayback) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
	return nil
}

func (m *MockAudioPlayback) WriteFrame(frame types.PCMFrame) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running {
		return &types.ErrorInfo{Module: "MockPlayback", Operation: "WriteFrame"}
	}
	m.receivedFrames = append(m.receivedFrames, frame)
	return nil
}

func (m *MockAudioPlayback) GetPlaybackLatency() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.latency
}

func (m *MockAudioPlayback) GetBufferFillLevel() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.bufferFillLevel
}

func (m *MockAudioPlayback) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initialized = false
	m.running = false
	return nil
}

// Helper methods for testing
func (m *MockAudioPlayback) GetReceivedFrames() []types.PCMFrame {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]types.PCMFrame{}, m.receivedFrames...)
}

func (m *MockAudioPlayback) SetError(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = true
	m.errorOnMethod = method
}

func (m *MockAudioPlayback) SetBufferFillLevel(level float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bufferFillLevel = level
}
