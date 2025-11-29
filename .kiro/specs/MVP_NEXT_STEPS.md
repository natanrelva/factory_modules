# MVP - Próximos Passos de Implementação

## 📍 Status Atual

### ✅ Completo
- Estrutura do projeto criada
- CLI básico implementado (cobra)
- Interfaces dos módulos simplificados
- Scripts de download de modelos
- Documentação do MVP
- M6 Audio Interface (já existia)

### 🔄 Em Progresso
- Integração com bibliotecas reais

## 🎯 Próximos Passos (Ordem de Implementação)

### Passo 1: Integrar Whisper.cpp para ASR (2 dias)

**Objetivo**: Fazer o reconhecimento de fala funcionar

**Tarefas**:
1. Adicionar dependência do Whisper.cpp Go bindings
2. Implementar carregamento do modelo
3. Implementar transcrição de áudio
4. Testar com áudio de exemplo

**Código a modificar**:
- `pkg/asr-simple/asr.go`
- `go.mod`

**Dependência**:
```go
// go.mod
require (
    github.com/ggerganov/whisper.cpp/bindings/go v0.0.0-latest
)
```

**Implementação**:
```go
// pkg/asr-simple/asr.go
import "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"

type SimpleASR struct {
    model   whisper.Model
    context whisper.Context
}

func (a *SimpleASR) Transcribe(audioSamples []float32) (string, error) {
    // Processar com Whisper
    result, err := a.context.Process(audioSamples, nil, nil)
    if err != nil {
        return "", err
    }
    
    // Extrair texto
    text := result.Text()
    return text, nil
}
```

**Teste**:
```bash
# Testar com arquivo de áudio
go run cmd/dubbing-mvp/main.go test-asr --file test.wav
```

---

### Passo 2: Integrar Google Translate API (1 dia)

**Objetivo**: Fazer a tradução PT→EN funcionar

**Tarefas**:
1. Adicionar cliente do Google Translate
2. Implementar tradução
3. Tratar rate limits e erros
4. Testar com frases de exemplo

**Código a modificar**:
- `pkg/translation-simple/translator.go`
- `go.mod`

**Dependência**:
```go
// go.mod
require (
    cloud.google.com/go/translate v1.10.0
    // OU para alternativa gratuita:
    github.com/bregydoc/gtranslate v0.0.0-latest
)
```

**Implementação**:
```go
// pkg/translation-simple/translator.go
import "github.com/bregydoc/gtranslate"

func (t *SimpleTranslator) Translate(textPT string) (string, error) {
    textEN, err := gtranslate.TranslateWithParams(
        textPT,
        gtranslate.TranslationParams{
            From: "pt",
            To:   "en",
        },
    )
    if err != nil {
        return "", err
    }
    return textEN, nil
}
```

**Teste**:
```bash
# Testar tradução
go run cmd/dubbing-mvp/main.go test-translation --text "Olá mundo"
# Esperado: "Hello world"
```

---

### Passo 3: Integrar Piper TTS (2 dias)

**Objetivo**: Fazer a síntese de voz funcionar

**Tarefas**:
1. Adicionar bindings do Piper TTS
2. Implementar carregamento do modelo de voz
3. Implementar síntese de áudio
4. Testar com texto de exemplo

**Código a modificar**:
- `pkg/tts-simple/tts.go`
- `go.mod`

**Dependência**:
```go
// go.mod
require (
    github.com/rhasspy/piper-go v0.0.0-latest
    // OU alternativa:
    github.com/hegedustibor/htgo-tts v0.0.0-latest
)
```

**Implementação**:
```go
// pkg/tts-simple/tts.go
import "github.com/rhasspy/piper-go/pkg/piper"

type SimpleTTS struct {
    engine *piper.Engine
    voice  *piper.Voice
}

func (t *SimpleTTS) Synthesize(textEN string) ([]float32, error) {
    // Sintetizar com Piper
    audio, err := t.engine.Synthesize(textEN, t.voice)
    if err != nil {
        return nil, err
    }
    return audio, nil
}
```

**Teste**:
```bash
# Testar síntese
go run cmd/dubbing-mvp/main.go test-tts --text "Hello world" --output test.wav
```

---

### Passo 4: Conectar com M6 Audio Interface (1 dia)

**Objetivo**: Integrar captura e reprodução de áudio

**Tarefas**:
1. Importar M6 Audio Interface
2. Configurar captura de áudio
3. Configurar reprodução de áudio
4. Testar loopback (captura → reprodução)

**Código a modificar**:
- `cmd/dubbing-mvp/main.go`

**Implementação**:
```go
// cmd/dubbing-mvp/main.go
import (
    "github.com/user/audio-dubbing-system/audio-interface/pkg/coordinator"
    "github.com/user/audio-dubbing-system/audio-interface/pkg/types"
)

func processingLoop() {
    // Inicializar M6
    audioCoord := coordinator.NewCoordinator(coordinator.Config{
        SampleRate:   16000,
        ChannelCount: 1,
        BufferSize:   4096,
    })
    
    audioCoord.Initialize()
    audioCoord.Start()
    
    // Loop de processamento
    for {
        // 1. Capturar chunk de áudio
        frames := captureAudioChunk(audioCoord, 3*time.Second)
        
        // 2. Converter para float32
        audioSamples := framesToFloat32(frames)
        
        // 3-5. Processar (ASR → Translation → TTS)
        // ...
        
        // 6. Reproduzir áudio traduzido
        playAudioChunk(audioCoord, synthesizedAudio)
    }
}
```

**Teste**:
```bash
# Testar loopback
go run cmd/dubbing-mvp/main.go test-loopback
# Deve capturar do mic e reproduzir no speaker
```

---

### Passo 5: Implementar Pipeline Completo (2 dias)

**Objetivo**: Conectar todos os módulos em um pipeline funcional

**Tarefas**:
1. Implementar loop de processamento completo
2. Adicionar tratamento de erros
3. Adicionar logging
4. Otimizar latência

**Código a modificar**:
- `cmd/dubbing-mvp/main.go`

**Implementação**:
```go
func processingLoop() {
    // Inicializar todos os módulos
    audioCoord := initAudioInterface()
    asr := initASR()
    translator := initTranslator()
    tts := initTTS()
    
    log.Println("Pipeline started")
    
    for {
        // 1. Capturar áudio (3s chunks)
        audioChunk := captureAudioChunk(audioCoord, 3*time.Second)
        if len(audioChunk) == 0 {
            continue // Silêncio
        }
        
        // 2. ASR: Áudio PT → Texto PT
        textPT, err := asr.Transcribe(audioChunk)
        if err != nil {
            log.Printf("ASR error: %v", err)
            continue
        }
        if textPT == "" {
            continue // Nada reconhecido
        }
        log.Printf("PT: %s", textPT)
        
        // 3. Translation: PT → EN
        textEN, err := translator.Translate(textPT)
        if err != nil {
            log.Printf("Translation error: %v", err)
            continue
        }
        log.Printf("EN: %s", textEN)
        
        // 4. TTS: Texto EN → Áudio EN
        audioEN, err := tts.Synthesize(textEN)
        if err != nil {
            log.Printf("TTS error: %v", err)
            continue
        }
        
        // 5. Reproduzir áudio EN
        playAudioChunk(audioCoord, audioEN)
        
        log.Println("Chunk processed successfully")
    }
}
```

**Teste**:
```bash
# Testar pipeline completo
./dubbing-mvp start
# Falar em português → Deve ouvir em inglês
```

---

### Passo 6: Testar com Google Meets (1 dia)

**Objetivo**: Validar funcionamento em aplicação real

**Tarefas**:
1. Configurar Virtual Cable
2. Configurar Google Meets
3. Testar dublagem em reunião
4. Ajustar latência e qualidade
5. Corrigir bugs encontrados

**Setup**:
```bash
# 1. Instalar Virtual Cable
# Windows: https://vb-audio.com/Cable/

# 2. Configurar dubbing-mvp
./dubbing-mvp config \
  --input "Microfone" \
  --output "Virtual Cable Input"

# 3. Iniciar dublagem
./dubbing-mvp start

# 4. Abrir Google Meets
# Configurar microfone: "Virtual Cable Output"

# 5. Entrar em reunião e testar
```

**Checklist de Testes**:
- [ ] Áudio é capturado corretamente
- [ ] Reconhecimento PT funciona
- [ ] Tradução PT→EN funciona
- [ ] Síntese EN funciona
- [ ] Outros participantes ouvem em inglês
- [ ] Latência é aceitável (< 3s)
- [ ] Não há crashes durante 10 minutos
- [ ] Qualidade é compreensível

---

## 📊 Cronograma Detalhado

| Dia | Tarefa | Entregável |
|-----|--------|------------|
| 1 | Integrar Whisper.cpp | ASR funcionando |
| 2 | Finalizar ASR + testes | ASR testado |
| 3 | Integrar Google Translate | Translation funcionando |
| 4 | Integrar Piper TTS | TTS funcionando |
| 5 | Finalizar TTS + testes | TTS testado |
| 6 | Conectar com M6 | Audio I/O funcionando |
| 7 | Pipeline completo | MVP funcionando |
| 8 | Testes e ajustes | MVP estável |
| 9 | Teste com Google Meets | MVP validado |

**Total: 9 dias úteis**

---

## 🐛 Problemas Esperados e Soluções

### Problema 1: Whisper.cpp não compila
**Solução**: Usar bindings pré-compilados ou Docker

### Problema 2: Google Translate rate limit
**Solução**: Implementar cache de traduções ou usar LibreTranslate

### Problema 3: Latência muito alta
**Solução**: 
- Reduzir chunk size (3s → 2s)
- Usar Whisper Tiny em vez de Small
- Processar em paralelo quando possível

### Problema 4: Qualidade de áudio ruim
**Solução**:
- Aumentar sample rate (16kHz → 22kHz)
- Usar modelo TTS melhor
- Adicionar filtros de áudio

### Problema 5: Virtual Cable não funciona
**Solução**:
- Reinstalar driver
- Usar alternativa (VoiceMeeter)
- Verificar permissões de áudio

---

## 🎯 Critérios de Sucesso

### Mínimo Viável
- ✅ Captura áudio do microfone
- ✅ Reconhece português
- ✅ Traduz para inglês
- ✅ Sintetiza voz inglesa
- ✅ Funciona no Google Meets

### Desejável
- ✅ Latência < 2 segundos
- ✅ Qualidade compreensível
- ✅ Estável por 10+ minutos
- ✅ Uso de recursos razoável

### Opcional (pós-MVP)
- Interface gráfica
- Voice cloning
- Múltiplos idiomas
- Configuração avançada

---

## 📝 Comandos Úteis

```bash
# Compilar
go build -o dubbing-mvp cmd/dubbing-mvp/main.go

# Executar
./dubbing-mvp start

# Testar módulos individualmente
go test ./pkg/asr-simple/...
go test ./pkg/translation-simple/...
go test ./pkg/tts-simple/...

# Debug
go run cmd/dubbing-mvp/main.go start --verbose

# Profiling
go run cmd/dubbing-mvp/main.go start --profile cpu.prof
```

---

## 🚀 Começar Agora

**Próximo comando a executar**:
```bash
# 1. Tornar script executável
chmod +x scripts/download-models.sh

# 2. Baixar modelos
./scripts/download-models.sh

# 3. Começar implementação do Passo 1 (Whisper integration)
# Editar: pkg/asr-simple/asr.go
```

---

**Status**: 📋 Pronto para implementação
**Próximo Passo**: Integrar Whisper.cpp (Passo 1)
**Tempo Estimado**: 9 dias úteis para MVP completo
