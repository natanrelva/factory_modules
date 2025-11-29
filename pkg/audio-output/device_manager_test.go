package audiooutput

import (
	"strings"
	"testing"
	"testing/quick"
)

// Unit Tests

func TestNewDeviceManager(t *testing.T) {
	dm := NewDeviceManager()
	
	if dm == nil {
		t.Fatal("NewDeviceManager returned nil")
	}
}

func TestEnumerateDevices(t *testing.T) {
	dm := NewDeviceManager()
	
	devices, err := dm.EnumerateDevices()
	if err != nil {
		t.Fatalf("EnumerateDevices failed: %v", err)
	}
	
	// Should have at least one device (default)
	if len(devices) == 0 {
		t.Error("Expected at least one device")
	}
	
	// Verify device structure
	for _, device := range devices {
		if device.Name == "" {
			t.Error("Device name should not be empty")
		}
		if device.ID == "" {
			t.Error("Device ID should not be empty")
		}
		if device.SampleRate <= 0 {
			t.Errorf("Invalid sample rate: %d", device.SampleRate)
		}
		if device.Channels <= 0 {
			t.Errorf("Invalid channels: %d", device.Channels)
		}
	}
}

func TestFindDevice_ValidName(t *testing.T) {
	dm := NewDeviceManager()
	
	// Get available devices
	devices, err := dm.EnumerateDevices()
	if err != nil {
		t.Fatalf("EnumerateDevices failed: %v", err)
	}
	
	if len(devices) == 0 {
		t.Skip("No devices available for testing")
	}
	
	// Try to find the first device by name
	device, err := dm.FindDevice(devices[0].Name)
	if err != nil {
		t.Fatalf("FindDevice failed: %v", err)
	}
	
	if device == nil {
		t.Fatal("FindDevice returned nil")
	}
	
	if device.Name != devices[0].Name {
		t.Errorf("Expected device name %s, got %s", devices[0].Name, device.Name)
	}
}

func TestFindDevice_InvalidName(t *testing.T) {
	dm := NewDeviceManager()
	
	device, err := dm.FindDevice("NonExistentDevice12345")
	
	if err == nil {
		t.Error("Expected error for non-existent device")
	}
	
	if device != nil {
		t.Error("Expected nil device for non-existent device")
	}
}

func TestDetectVirtualCables(t *testing.T) {
	dm := NewDeviceManager()
	
	virtualDevices, err := dm.DetectVirtualCables()
	if err != nil {
		t.Fatalf("DetectVirtualCables failed: %v", err)
	}
	
	// Virtual devices may or may not be present
	// Just verify the function works
	t.Logf("Found %d virtual cable devices", len(virtualDevices))
	
	// Verify all returned devices are marked as virtual
	for _, device := range virtualDevices {
		if !device.IsVirtual {
			t.Errorf("Device %s should be marked as virtual", device.Name)
		}
	}
}

func TestDetectVirtualCables_KnownPatterns(t *testing.T) {
	dm := NewDeviceManager()
	
	// Test known virtual cable patterns
	testCases := []struct {
		name      string
		isVirtual bool
	}{
		{"CABLE Input (VB-Audio Virtual Cable)", true},
		{"CABLE Output (VB-Audio Virtual Cable)", true},
		{"VoiceMeeter Input", true},
		{"VoiceMeeter Output", true},
		{"Virtual Audio Cable", true},
		{"Realtek HD Audio", false},
		{"Microphone Array", false},
		{"Speakers", false},
	}
	
	for _, tc := range testCases {
		isVirtual := dm.isVirtualCable(tc.name)
		if isVirtual != tc.isVirtual {
			t.Errorf("Device %s: expected isVirtual=%v, got %v", tc.name, tc.isVirtual, isVirtual)
		}
	}
}

func TestValidateDevice_Valid(t *testing.T) {
	dm := NewDeviceManager()
	
	devices, err := dm.EnumerateDevices()
	if err != nil {
		t.Fatalf("EnumerateDevices failed: %v", err)
	}
	
	if len(devices) == 0 {
		t.Skip("No devices available for testing")
	}
	
	// Validate first device
	err = dm.ValidateDevice(devices[0])
	if err != nil {
		t.Errorf("ValidateDevice failed for valid device: %v", err)
	}
}

func TestValidateDevice_Invalid(t *testing.T) {
	dm := NewDeviceManager()
	
	invalidDevice := AudioDevice{
		Name:       "NonExistent",
		ID:         "invalid-id",
		SampleRate: 16000,
		Channels:   1,
	}
	
	err := dm.ValidateDevice(invalidDevice)
	if err == nil {
		t.Error("Expected error for invalid device")
	}
}

func TestGetDefaultDevice(t *testing.T) {
	dm := NewDeviceManager()
	
	device, err := dm.GetDefaultDevice()
	if err != nil {
		t.Fatalf("GetDefaultDevice failed: %v", err)
	}
	
	if device == nil {
		t.Fatal("GetDefaultDevice returned nil")
	}
	
	if !device.IsDefault {
		t.Error("Default device should be marked as default")
	}
}

// Property-Based Tests

// Property 2: Device Enumeration Consistency
// For any system state, enumerating devices twice SHALL return the same devices
func TestProperty_DeviceEnumerationConsistency(t *testing.T) {
	f := func() bool {
		dm := NewDeviceManager()
		
		// Enumerate devices twice
		devices1, err1 := dm.EnumerateDevices()
		if err1 != nil {
			return true // Skip on error
		}
		
		devices2, err2 := dm.EnumerateDevices()
		if err2 != nil {
			return true // Skip on error
		}
		
		// Should have same number of devices
		if len(devices1) != len(devices2) {
			return false
		}
		
		// Create maps for comparison
		map1 := make(map[string]AudioDevice)
		map2 := make(map[string]AudioDevice)
		
		for _, d := range devices1 {
			map1[d.ID] = d
		}
		
		for _, d := range devices2 {
			map2[d.ID] = d
		}
		
		// Verify all devices from first enumeration exist in second
		for id, d1 := range map1 {
			d2, exists := map2[id]
			if !exists {
				return false
			}
			
			// Verify key properties match
			if d1.Name != d2.Name {
				return false
			}
			if d1.IsDefault != d2.IsDefault {
				return false
			}
		}
		
		return true
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 10}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property: Device Validation Correctness
// For any device from enumeration, validation should succeed
func TestProperty_DeviceValidation(t *testing.T) {
	f := func() bool {
		dm := NewDeviceManager()
		
		devices, err := dm.EnumerateDevices()
		if err != nil || len(devices) == 0 {
			return true // Skip on error or no devices
		}
		
		// All enumerated devices should be valid
		for _, device := range devices {
			if err := dm.ValidateDevice(device); err != nil {
				return false
			}
		}
		
		return true
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 10}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Property: Virtual Cable Detection Accuracy
// For any device marked as virtual, name should contain virtual cable patterns
func TestProperty_VirtualCableDetection(t *testing.T) {
	f := func() bool {
		dm := NewDeviceManager()
		
		virtualDevices, err := dm.DetectVirtualCables()
		if err != nil {
			return true // Skip on error
		}
		
		// All virtual devices should have virtual patterns in name
		virtualPatterns := []string{"cable", "virtual", "voicemeeter", "vb-audio"}
		
		for _, device := range virtualDevices {
			if !device.IsVirtual {
				return false
			}
			
			// Name should contain at least one virtual pattern
			nameLower := strings.ToLower(device.Name)
			hasPattern := false
			for _, pattern := range virtualPatterns {
				if strings.Contains(nameLower, pattern) {
					hasPattern = true
					break
				}
			}
			
			if !hasPattern {
				return false
			}
		}
		
		return true
	}
	
	if err := quick.Check(f, &quick.Config{MaxCount: 10}); err != nil {
		t.Error("Property violated:", err)
	}
}

// Benchmark Tests

func BenchmarkEnumerateDevices(b *testing.B) {
	dm := NewDeviceManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dm.EnumerateDevices()
	}
}

func BenchmarkFindDevice(b *testing.B) {
	dm := NewDeviceManager()
	
	devices, err := dm.EnumerateDevices()
	if err != nil || len(devices) == 0 {
		b.Skip("No devices available")
	}
	
	deviceName := devices[0].Name
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dm.FindDevice(deviceName)
	}
}

func BenchmarkDetectVirtualCables(b *testing.B) {
	dm := NewDeviceManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dm.DetectVirtualCables()
	}
}
