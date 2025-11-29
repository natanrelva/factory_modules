package audiooutput

import (
	"fmt"
	"strings"
	"sync"
)

// DeviceManager manages audio output devices
type DeviceManager struct {
	mu      sync.RWMutex
	devices []AudioDevice
}

// NewDeviceManager creates a new device manager
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make([]AudioDevice, 0),
	}
}

// EnumerateDevices lists all available audio output devices
func (dm *DeviceManager) EnumerateDevices() ([]AudioDevice, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	// TODO: Replace with actual PortAudio enumeration
	// For now, return mock devices for testing
	devices := []AudioDevice{
		{
			Name:       "Default Output Device",
			ID:         "default",
			IsDefault:  true,
			IsVirtual:  false,
			SampleRate: 48000,
			Channels:   2,
		},
		{
			Name:       "Speakers (Realtek HD Audio)",
			ID:         "speakers-realtek",
			IsDefault:  false,
			IsVirtual:  false,
			SampleRate: 48000,
			Channels:   2,
		},
		{
			Name:       "CABLE Input (VB-Audio Virtual Cable)",
			ID:         "cable-input-vb",
			IsDefault:  false,
			IsVirtual:  true,
			SampleRate: 48000,
			Channels:   2,
		},
		{
			Name:       "VoiceMeeter Input (VB-Audio VoiceMeeter VAIO)",
			ID:         "voicemeeter-input",
			IsDefault:  false,
			IsVirtual:  true,
			SampleRate: 48000,
			Channels:   2,
		},
	}
	
	dm.devices = devices
	return devices, nil
}

// FindDevice finds a device by name or ID
func (dm *DeviceManager) FindDevice(nameOrID string) (*AudioDevice, error) {
	devices, err := dm.EnumerateDevices()
	if err != nil {
		return nil, err
	}
	
	// Search by name or ID
	for _, device := range devices {
		if device.Name == nameOrID || device.ID == nameOrID {
			return &device, nil
		}
	}
	
	return nil, fmt.Errorf("device not found: %s", nameOrID)
}

// DetectVirtualCables detects virtual audio cable devices
func (dm *DeviceManager) DetectVirtualCables() ([]AudioDevice, error) {
	devices, err := dm.EnumerateDevices()
	if err != nil {
		return nil, err
	}
	
	virtualDevices := make([]AudioDevice, 0)
	
	for _, device := range devices {
		if dm.isVirtualCable(device.Name) {
			device.IsVirtual = true
			virtualDevices = append(virtualDevices, device)
		}
	}
	
	return virtualDevices, nil
}

// isVirtualCable checks if a device name indicates a virtual audio cable
func (dm *DeviceManager) isVirtualCable(name string) bool {
	nameLower := strings.ToLower(name)
	
	// Known virtual cable patterns
	virtualPatterns := []string{
		"cable",
		"virtual",
		"voicemeeter",
		"vb-audio",
		"vac", // Virtual Audio Cable
		"loopback",
	}
	
	for _, pattern := range virtualPatterns {
		if strings.Contains(nameLower, pattern) {
			return true
		}
	}
	
	return false
}

// ValidateDevice validates that a device exists and is accessible
func (dm *DeviceManager) ValidateDevice(device AudioDevice) error {
	devices, err := dm.EnumerateDevices()
	if err != nil {
		return fmt.Errorf("failed to enumerate devices: %w", err)
	}
	
	// Check if device exists
	for _, d := range devices {
		if d.ID == device.ID || d.Name == device.Name {
			// Device exists
			return nil
		}
	}
	
	return fmt.Errorf("device not found or not accessible: %s", device.Name)
}

// GetDefaultDevice returns the default output device
func (dm *DeviceManager) GetDefaultDevice() (*AudioDevice, error) {
	devices, err := dm.EnumerateDevices()
	if err != nil {
		return nil, err
	}
	
	// Find default device
	for _, device := range devices {
		if device.IsDefault {
			return &device, nil
		}
	}
	
	// If no default found, return first device
	if len(devices) > 0 {
		return &devices[0], nil
	}
	
	return nil, fmt.Errorf("no output devices available")
}

// WatchDevices watches for device changes (hot-plugging)
// TODO: Implement actual device watching
func (dm *DeviceManager) WatchDevices(callback func(event DeviceEvent)) error {
	// TODO: Implement device watching
	// For now, this is a placeholder
	return nil
}
