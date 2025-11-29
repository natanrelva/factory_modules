# Integração Rápida - MVP Funcional

## 🎯 Objetivo

Fazer o MVP funcionar **rapidamente** com bibliotecas Go simples, sem dependências complexas.

## 🚀 Abordagem Simplificada

Em vez de integrar Whisper.cpp, Piper TTS, etc. (que são complexos), vamos usar:

### 1. ASR - Usar Vosk (Go bindings prontos)
```bash
go get github.com/alphacep/vosk-api/go
```

**Vantagens**:
- Bindings Go nativos
- Modelos pequenos (~50MB)
- Funciona offline
- Fácil de instalar

### 2. Translation - Usar LibreTranslate (API gratuita)
```bash
# Sem instalação necessária
# Usa API pública: https://libretranslate.com
```

**Vantagens**:
- API gratuita
- Sem API key necessária
- Boa qualidade
- Simples HTTP request

### 3. TTS - Usar eSpeak (via exec)
```bash
# Linux
sudo apt-get install espeak

# macOS
brew install espeak

# Windows
# Download: http://espeak.sourceforge.net/download.html
```

**Vantagens**:
- Fácil de instalar
- Funciona via command line
- Sem dependências Go complexas

## 📦 Instalação Rápida

### Linux (Ubuntu/Debian)
```bash
# 1. Instalar eSpeak
sudo apt-get update
sudo apt-get install espeak espeak-data

# 2. Baixar modelo Vosk
mkdir -p models
cd models
wget https://alphacephei.com/vosk/models/vosk-model-small-pt-0.3.zip
unzip vosk-model-small-pt-0.3.zip
mv vosk-model-small-pt-0.3 vosk-model-pt
cd ..

# 3. Instalar dependências Go
go get github.com/alphacep/vosk-api/go
go mod tidy

# 4. Compilar
go build -o dubbing-mvp cmd/dubbing-mvp/main.go

# 5. Testar
./dubbing-mvp start
```

### macOS
```bash
# 1. Instalar eSpeak
brew install espeak

# 2. Baixar modelo Vosk
mkdir -p models
cd models
curl -LO https://alphacephei.com/vosk/models/vosk-model-small-pt-0.3.zip
unzip vosk-model-small-pt-0.3.zip
mv vosk-model-small-pt-0.3 vosk-model-pt
cd ..

# 3. Instalar dependências Go
go get github.com/alphacep/vosk-api/go
go mod tidy

# 4. Compilar
go build -o dubbing-mvp cmd/dubbing-mvp/main.go

# 5. Testar
./dubbing-mvp start
```

### Windows
```powershell
# 1. Instalar eSpeak
# Download: http://espeak.sourceforge.net/download.html
# Instalar e adicionar ao PATH

# 2. Baixar modelo Vosk
mkdir models
cd models
# Download manual: https://alphacephei.com/vosk/models/vosk-model-small-pt-0.3.zip
# Extrair para models/vosk-model-pt
cd ..

# 3. Instalar dependências Go
go get github.com/alphacep/vosk-api/go
go mod tidy

# 4. Compilar
go build -o dubbing-mvp.exe cmd/dubbing-mvp/main.go

# 5. Testar
.\dubbing-mvp.exe start
```

## 🔧 Implementação

### 1. ASR com Vosk

```go
// pkg/asr-vosk/asr.go
package asrvosk

import (
    "github.com/alphacep/vosk-api/go"
)

type VoskASR struct {
    model *vosk.VoskModel
    rec   *vosk.VoskRecognizer
}

func NewVoskASR(modelPath string) (*VoskASR, error) {
    model, err := vosk.NewModel(modelPath)
    if err != nil {
        return nil, err
    }
    
    rec, err := vosk.NewRecognizer(model, 16000.0)
    if err != nil {
        return nil, err
    }
    
    return &VoskASR{model: model, rec: rec}, nil
}

func (a *VoskASR) Transcribe(samples []float32) (string, error) {
    // Convert float32 to int16
    data := make([]int16, len(samples))
    for i, s := range samples {
        data[i] = int16(s * 32767)
    }
    
    // Feed to recognizer
    a.rec.AcceptWaveform(data)
    result := a.rec.Result()
    
    // Parse JSON result
    var res struct {
        Text string `json:"text"`
    }
    json.Unmarshal([]byte(result), &res)
    
    return res.Text, nil
}
```

### 2. Translation com LibreTranslate

```go
// pkg/translation-libre/translator.go
package translationlibre

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func Translate(text, source, target string) (string, error) {
    url := "https://libretranslate.com/translate"
    
    payload := map[string]string{
        "q":      text,
        "source": source,
        "target": target,
    }
    
    data, _ := json.Marshal(payload)
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var result struct {
        TranslatedText string `json:"translatedText"`
    }
    
    json.NewDecoder(resp.Body).Decode(&result)
    return result.TranslatedText, nil
}
```

### 3. TTS com eSpeak

```go
// pkg/tts-espeak/tts.go
package ttsespeak

import (
    "os/exec"
)

func Synthesize(text string, outputFile string) error {
    cmd := exec.Command("espeak", text, "-w", outputFile)
    return cmd.Run()
}

// Or get audio data directly
func SynthesizeToBytes(text string) ([]byte, error) {
    cmd := exec.Command("espeak", text, "--stdout")
    return cmd.Output()
}
```

## ✅ Vantagens desta Abordagem

1. **Rápido de implementar** - Poucas dependências
2. **Funciona offline** - Vosk e eSpeak são locais
3. **Gratuito** - Sem custos de API
4. **Simples** - Sem compilação complexa
5. **Testável** - Fácil de validar

## 📊 Qualidade Esperada

| Componente | Qualidade | Observação |
|------------|-----------|------------|
| Vosk ASR | ⭐⭐⭐ Boa | WER ~20% (aceitável) |
| LibreTranslate | ⭐⭐⭐⭐ Muito boa | BLEU ~35 |
| eSpeak TTS | ⭐⭐ Básica | Robótico mas claro |

**Para MVP**: Qualidade suficiente! ✅

## 🎯 Próximos Passos

1. Implementar `pkg/asr-vosk/asr.go`
2. Implementar `pkg/translation-libre/translator.go`
3. Implementar `pkg/tts-espeak/tts.go`
4. Atualizar `cmd/dubbing-mvp/main.go` para usar versões reais
5. Testar pipeline completo
6. Validar com Google Meets

**Tempo estimado**: 1-2 dias

## 🚀 Alternativa AINDA Mais Rápida

Se até isso for complexo, podemos usar **apenas APIs**:

```go
// Usar serviços online para tudo
// - Vosk Server (self-hosted ou cloud)
// - LibreTranslate API (público)
// - Google TTS API (com free tier)

// Vantagem: Zero instalação local
// Desvantagem: Requer internet
```

---

**Recomendação**: Usar Vosk + LibreTranslate + eSpeak para MVP rápido e funcional!
