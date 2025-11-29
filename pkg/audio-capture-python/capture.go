package audiocapturepython

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// PythonAudioCapture implementa captura de áudio usando PyAudio via Python
type PythonAudioCapture struct {
	deviceName string
	sampleRate int
	channels   int
	pythonPath string
	scriptPath string
	mu         sync.Mutex
}

// NewPythonAudioCapture cria uma nova instância de captura de áudio
func NewPythonAudioCapture(deviceName string, sampleRate int) (*PythonAudioCapture, error) {
	// Encontrar Python
	pythonPath, err := findPython()
	if err != nil {
		return nil, fmt.Errorf("Python not found: %w", err)
	}
	
	// Encontrar script
	scriptPath, err := findScript()
	if err != nil {
		return nil, fmt.Errorf("Audio capture script not found: %w", err)
	}
	
	capture := &PythonAudioCapture{
		deviceName: deviceName,
		sampleRate: sampleRate,
		channels:   1, // Mono
		pythonPath: pythonPath,
		scriptPath: scriptPath,
	}
	
	log.Printf("✓ Python Audio Capture initialized")
	log.Printf("  Device: %s", deviceName)
	log.Printf("  Sample Rate: %d Hz", sampleRate)
	log.Printf("  Channels: %d", 1)
	log.Printf("  Mode: Real (PyAudio)")
	log.Printf("  Python: %s", pythonPath)
	
	return capture, nil
}

// findPython encontra o executável do Python
func findPython() (string, error) {
	paths := []string{
		"python",
		"python3",
		"C:\\Users\\natan\\AppData\\Local\\Programs\\Python\\Python313\\python.exe",
	}
	
	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("Python not found")
}

// findScript encontra o script audio-capture.py
func findScript() (string, error) {
	paths := []string{
		"scripts/audio-capture.py",
		"../scripts/audio-capture.py",
		"../../scripts/audio-capture.py",
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}
	
	return "", fmt.Errorf("audio-capture.py not found")
}

// Capture captura áudio do microfone
func (c *PythonAudioCapture) Capture(duration float64) ([]float32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Criar arquivo WAV temporário
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("capture_%d.wav", os.Getpid()))
	defer os.Remove(tmpFile)
	
	// Executar script Python para capturar
	cmd := exec.Command(
		c.pythonPath,
		c.scriptPath,
		"capture",
		tmpFile,
		fmt.Sprintf("%.1f", duration),
		fmt.Sprintf("%d", c.sampleRate),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("capture failed: %w\nOutput: %s", err, string(output))
	}
	
	// Ler arquivo WAV e converter para samples
	samples, err := readWAVSamples(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read WAV: %w", err)
	}
	
	log.Printf("✓ Captured %d samples from microphone", len(samples))
	
	return samples, nil
}

// readWAVSamples lê samples de um arquivo WAV
func readWAVSamples(filename string) ([]float32, error) {
	// Abrir arquivo
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	// Pular header WAV (44 bytes)
	header := make([]byte, 44)
	if _, err := file.Read(header); err != nil {
		return nil, err
	}
	
	// Ler dados de áudio
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	// Converter int16 para float32
	audioData := data[44:] // Pular header
	samples := make([]float32, len(audioData)/2)
	
	for i := 0; i < len(samples); i++ {
		// Ler int16 (little-endian)
		idx := i * 2
		if idx+1 < len(audioData) {
			intSample := int16(audioData[idx]) | int16(audioData[idx+1])<<8
			// Converter para float32 (-1.0 a 1.0)
			samples[i] = float32(intSample) / 32768.0
		}
	}
	
	return samples, nil
}

// Close fecha a captura
func (c *PythonAudioCapture) Close() error {
	log.Println("Python Audio Capture closed")
	return nil
}

// GetDeviceName retorna o nome do dispositivo
func (c *PythonAudioCapture) GetDeviceName() string {
	return c.deviceName
}

// GetSampleRate retorna a taxa de amostragem
func (c *PythonAudioCapture) GetSampleRate() int {
	return c.sampleRate
}
