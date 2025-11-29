package asrvoskpython

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// VoskASR implementa ASR usando Vosk via Python
type VoskASR struct {
	language   string
	sampleRate int
	modelPath  string
	pythonPath string
	scriptPath string
	mu         sync.Mutex
}

// NewVoskASR cria uma nova instância do Vosk ASR
func NewVoskASR(language string, sampleRate int) (*VoskASR, error) {
	// Encontrar Python
	pythonPath, err := findPython()
	if err != nil {
		return nil, fmt.Errorf("Python not found: %w", err)
	}
	
	// Encontrar script
	scriptPath, err := findScript()
	if err != nil {
		return nil, fmt.Errorf("Vosk script not found: %w", err)
	}
	
	// Encontrar modelo
	modelPath, err := findModel(language)
	if err != nil {
		return nil, fmt.Errorf("Vosk model not found: %w", err)
	}
	
	asr := &VoskASR{
		language:   language,
		sampleRate: sampleRate,
		modelPath:  modelPath,
		pythonPath: pythonPath,
		scriptPath: scriptPath,
	}
	
	log.Printf("✓ Vosk ASR initialized")
	log.Printf("  Language: %s", language)
	log.Printf("  Sample Rate: %d Hz", sampleRate)
	log.Printf("  Model: %s", modelPath)
	log.Printf("  Python: %s", pythonPath)
	
	return asr, nil
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

// findScript encontra o script vosk-asr.py
func findScript() (string, error) {
	paths := []string{
		"scripts/vosk-asr.py",
		"../scripts/vosk-asr.py",
		"../../scripts/vosk-asr.py",
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath, nil
		}
	}
	
	return "", fmt.Errorf("vosk-asr.py not found")
}

// findModel encontra o modelo Vosk
func findModel(language string) (string, error) {
	// Mapear idioma para modelo
	modelNames := map[string][]string{
		"pt": {
			"models/vosk-model-small-pt-0.3",
			"models/vosk-model-pt-fb-v0.1.1-20220516_2113",
			"../models/vosk-model-small-pt-0.3",
		},
		"en": {
			"models/vosk-model-small-en-us-0.15",
			"models/vosk-model-en-us-0.22",
			"../models/vosk-model-small-en-us-0.15",
		},
	}
	
	models, ok := modelNames[language]
	if !ok {
		return "", fmt.Errorf("unsupported language: %s", language)
	}
	
	for _, modelPath := range models {
		if _, err := os.Stat(modelPath); err == nil {
			absPath, _ := filepath.Abs(modelPath)
			return absPath, nil
		}
	}
	
	return "", fmt.Errorf("Vosk model not found for language: %s", language)
}

// Transcribe transcreve áudio em texto
func (a *VoskASR) Transcribe(audioSamples []float32) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if len(audioSamples) == 0 {
		return "", fmt.Errorf("empty audio")
	}
	
	// Criar arquivo WAV temporário
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("asr_%d.wav", os.Getpid()))
	defer os.Remove(tmpFile)
	
	// Salvar samples como WAV
	if err := saveWAV(tmpFile, audioSamples, a.sampleRate); err != nil {
		return "", fmt.Errorf("failed to save WAV: %w", err)
	}
	
	// Executar script Python
	cmd := exec.Command(
		a.pythonPath,
		a.scriptPath,
		tmpFile,
		a.modelPath,
		fmt.Sprintf("%d", a.sampleRate),
	)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Vosk failed: %w\nStderr: %s", err, stderr.String())
	}
	
	// Extrair texto do output
	output := stdout.String()
	lines := strings.Split(output, "\n")
	
	// Procurar pela última linha não vazia (que é o texto transcrito)
	text := ""
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && !strings.HasPrefix(line, "[") {
			text = line
			break
		}
	}
	
	if text == "" {
		log.Println("Vosk: No speech detected")
		return "", nil
	}
	
	log.Printf("Vosk: Transcribed '%s'", text)
	return text, nil
}

// saveWAV salva samples como arquivo WAV
func saveWAV(filename string, samples []float32, sampleRate int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// WAV header
	numSamples := len(samples)
	dataSize := numSamples * 2 // 16-bit = 2 bytes
	
	// RIFF header
	file.WriteString("RIFF")
	binary.Write(file, binary.LittleEndian, uint32(36+dataSize))
	file.WriteString("WAVE")
	
	// fmt chunk
	file.WriteString("fmt ")
	binary.Write(file, binary.LittleEndian, uint32(16))        // chunk size
	binary.Write(file, binary.LittleEndian, uint16(1))         // PCM
	binary.Write(file, binary.LittleEndian, uint16(1))         // mono
	binary.Write(file, binary.LittleEndian, uint32(sampleRate))
	binary.Write(file, binary.LittleEndian, uint32(sampleRate*2)) // byte rate
	binary.Write(file, binary.LittleEndian, uint16(2))         // block align
	binary.Write(file, binary.LittleEndian, uint16(16))        // bits per sample
	
	// data chunk
	file.WriteString("data")
	binary.Write(file, binary.LittleEndian, uint32(dataSize))
	
	// Write samples
	for _, sample := range samples {
		// Clampar entre -1.0 e 1.0
		if sample > 1.0 {
			sample = 1.0
		} else if sample < -1.0 {
			sample = -1.0
		}
		
		// Converter para int16
		intSample := int16(sample * 32767)
		binary.Write(file, binary.LittleEndian, intSample)
	}
	
	return nil
}

// Close fecha o ASR
func (a *VoskASR) Close() error {
	log.Println("Vosk ASR closed")
	return nil
}
