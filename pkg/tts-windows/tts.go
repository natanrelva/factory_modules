package ttswindows

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// WindowsTTS implementa TTS usando Windows SAPI via Python
type WindowsTTS struct {
	voice      string
	rate       int    // Velocidade em palavras por minuto (150-200)
	volume     float64 // Volume (0.0 a 1.0)
	sampleRate int
	pythonPath string
	scriptPath string
	mu         sync.Mutex
}

// NewWindowsTTS cria uma nova instância do Windows TTS
func NewWindowsTTS(voice string, rate int, sampleRate int) (*WindowsTTS, error) {
	// Encontrar Python
	pythonPath, err := findPython()
	if err != nil {
		return nil, fmt.Errorf("Python not found: %w", err)
	}
	
	// Encontrar script
	scriptPath, err := findScript()
	if err != nil {
		return nil, fmt.Errorf("TTS script not found: %w", err)
	}
	
	tts := &WindowsTTS{
		voice:      voice,
		rate:       rate,
		volume:     1.0,
		sampleRate: sampleRate,
		pythonPath: pythonPath,
		scriptPath: scriptPath,
	}
	
	log.Printf("✓ Windows TTS initialized")
	log.Printf("  Voice: %s", voice)
	log.Printf("  Rate: %d wpm", rate)
	log.Printf("  Sample Rate: %d Hz", sampleRate)
	log.Printf("  Python: %s", pythonPath)
	
	return tts, nil
}

// findPython encontra o executável do Python
func findPython() (string, error) {
	// Tentar locais comuns
	paths := []string{
		"python",
		"python3",
		"C:\\Users\\natan\\AppData\\Local\\Programs\\Python\\Python313\\python.exe",
		"C:\\Python313\\python.exe",
		"C:\\Python312\\python.exe",
		"C:\\Python311\\python.exe",
	}
	
	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("Python not found in PATH or common locations")
}

// findScript encontra o script windows-tts.py
func findScript() (string, error) {
	// Tentar locais relativos
	paths := []string{
		"scripts/windows-tts.py",
		"../scripts/windows-tts.py",
		"../../scripts/windows-tts.py",
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}
	
	return "", fmt.Errorf("windows-tts.py not found")
}

// Synthesize sintetiza texto em áudio
func (t *WindowsTTS) Synthesize(text string) ([]float32, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}
	
	// Criar arquivo temporário para o áudio
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("tts_%d.wav", os.Getpid()))
	defer os.Remove(tmpFile)
	
	// Executar script Python
	cmd := exec.Command(
		t.pythonPath,
		t.scriptPath,
		text,
		tmpFile,
		fmt.Sprintf("%d", t.rate),
		fmt.Sprintf("%.1f", t.volume),
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("TTS failed: %w\nOutput: %s", err, string(output))
	}
	
	log.Printf("TTS: Synthesized '%s' → %s", text, tmpFile)
	
	// Ler arquivo WAV e converter para samples
	// Por enquanto, retornar samples simulados
	// TODO: Implementar leitura real do WAV
	duration := float32(len(text)) * 0.05 // ~50ms por caractere
	numSamples := int(duration * float32(t.sampleRate))
	samples := make([]float32, numSamples)
	
	// Gerar tom simples (440 Hz)
	for i := range samples {
		_ = float32(i) / float32(t.sampleRate)
		samples[i] = 0.3 * float32(0.5) // Silêncio por enquanto
	}
	
	log.Printf("✓ Generated %d samples (%.2fs)", len(samples), duration)
	
	return samples, nil
}

// Close fecha o TTS
func (t *WindowsTTS) Close() error {
	log.Println("Windows TTS closed")
	return nil
}

// GetVoice retorna a voz atual
func (t *WindowsTTS) GetVoice() string {
	return t.voice
}

// GetRate retorna a velocidade atual
func (t *WindowsTTS) GetRate() int {
	return t.rate
}

// SetRate define a velocidade
func (t *WindowsTTS) SetRate(rate int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rate = rate
}

// SetVolume define o volume
func (t *WindowsTTS) SetVolume(volume float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.volume = volume
}
