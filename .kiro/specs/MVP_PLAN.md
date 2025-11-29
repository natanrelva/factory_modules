# MVP Plan - Sistema de Dublagem PT→EN

## 🎯 Objetivo do MVP

Criar um sistema **minimamente funcional** que permita dublagem automática PT→EN em tempo real para uso em:
- Google Meets
- Zoom
- Discord
- Microsoft Teams
- Qualquer aplicação que use áudio do sistema

## 🚀 Escopo do MVP

### ✅ O Que ESTÁ no MVP

1. **Captura de áudio do microfone** (M6 - já implementado)
2. **Reconhecimento de fala PT** (M2 - versão simplificada)
3. **Tradução PT→EN** (M3 - versão simplificada)
4. **Síntese de voz EN** (M4 - versão simplificada)
5. **Saída para dispositivo virtual** (M6 - já implementado)
6. **CLI básico** para controle (M0 - versão mínima)

### ❌ O Que NÃO está no MVP

- ❌ Interface gráfica (System Tray, Overlay)
- ❌ Voice cloning (usa voz padrão)
- ❌ Prosody transfer avançado
- ❌ Perfis de uso
- ❌ Dashboard de métricas
- ❌ Language model integration
- ❌ Context window (tradução sentença por sentença)
- ❌ Quality assurance avançada
- ❌ Adaptive optimization

## 📊 Arquitetura Simplificada do MVP

```
┌─────────────┐
│ Microfone   │ Áudio PT
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│ M6: Audio Capture   │ ✅ JÁ IMPLEMENTADO
│ - WASAPI Capture    │
│ - Basic VAD         │
└──────┬──────────────┘
       │ PCM Frames
       ▼
┌─────────────────────┐
│ M2: ASR (Simple)    │ 📋 IMPLEMENTAR
│ - Whisper Tiny      │
│ - No streaming      │
│ - Chunk-based       │
└──────┬──────────────┘
       │ Text PT
       ▼
┌─────────────────────┐
│ M3: Translation     │ 📋 IMPLEMENTAR
│ - Google Translate  │
│ - API simples       │
│ - No context        │
└──────┬──────────────┘
       │ Text EN
       ▼
┌─────────────────────┐
│ M4: TTS (Simple)    │ 📋 IMPLEMENTAR
│ - gTTS / Piper      │
│ - Voz padrão        │
│ - No cloning        │
└──────┬──────────────┘
       │ PCM Frames EN
       ▼
┌─────────────────────┐
│ M6: Audio Playback  │ ✅ JÁ IMPLEMENTADO
│ - Virtual Device    │
│ - WASAPI Playback   │
└──────┬──────────────┘
       │
       ▼
┌─────────────┐
│ Google Meet │ Áudio EN
│ Zoom, etc   │
└─────────────┘
```

## 🎯 Requisitos do MVP

### Funcionais
1. ✅ Capturar áudio do microfone
2. ✅ Detectar fala (VAD básico)
3. 📋 Reconhecer fala em português
4. 📋 Traduzir para inglês
5. 📋 Sintetizar voz em inglês
6. ✅ Enviar para dispositivo virtual
7. 📋 Controlar via CLI (start/stop)

### Não-Funcionais
- Latência: < 2 segundos (relaxado para MVP)
- Qualidade: "Compreensível" (não perfeito)
- Estabilidade: Não crashar
- Uso de recursos: < 50% CPU, < 1GB RAM

## 📋 Plano de Implementação do MVP

### Fase 1: Setup e Infraestrutura (1 dia)

**1.1 Criar estrutura do projeto MVP**
```bash
audio-interface/
├── cmd/
│   └── dubbing-mvp/
│       └── main.go          # CLI principal
├── pkg/
│   ├── asr-simple/          # ASR simplificado
│   ├── translation-simple/  # Translation simplificado
│   └── tts-simple/          # TTS simplificado
└── go.mod
```

**1.2 Configurar dependências**
- Whisper.cpp (Go bindings)
- Google Translate API ou LibreTranslate
- Piper TTS ou gTTS

### Fase 2: M2 ASR Simplificado (2 dias)

**2.1 Implementar ASR básico**
```go
// pkg/asr-simple/asr.go
type SimpleASR struct {
    model *whisper.Model
}

func (a *SimpleASR) Transcribe(audio []float32) (string, error) {
    // Usar Whisper Tiny para velocidade
    // Processar chunk completo (não streaming)
    // Retornar texto simples
}
```

**Features**:
- ✅ Usar Whisper Tiny (mais rápido)
- ✅ Processar chunks de 3 segundos
- ✅ Sem streaming (espera chunk completo)
- ✅ Sem language model
- ✅ Sem timestamp alignment

**Testes mínimos**:
- Teste com áudio de exemplo
- Verificar que retorna texto PT

### Fase 3: M3 Translation Simplificado (1 dia)

**3.1 Implementar tradução básica**
```go
// pkg/translation-simple/translator.go
type SimpleTranslator struct {
    apiKey string
}

func (t *SimpleTranslator) Translate(textPT string) (string, error) {
    // Usar Google Translate API
    // Ou LibreTranslate (self-hosted, grátis)
    // Tradução direta, sem contexto
}
```

**Features**:
- ✅ Google Translate API (ou LibreTranslate)
- ✅ Tradução sentença por sentença
- ✅ Sem context window
- ✅ Sem quality assurance
- ✅ Sem prosody annotation

**Testes mínimos**:
- Teste com frases de exemplo
- Verificar tradução PT→EN

### Fase 4: M4 TTS Simplificado (2 dias)

**4.1 Implementar TTS básico**
```go
// pkg/tts-simple/tts.go
type SimpleTTS struct {
    voice string
}

func (t *SimpleTTS) Synthesize(textEN string) ([]float32, error) {
    // Usar Piper TTS (local, rápido)
    // Ou gTTS (Google TTS API)
    // Voz padrão, sem cloning
}
```

**Features**:
- ✅ Piper TTS (local) ou gTTS (API)
- ✅ Voz feminina/masculina padrão
- ✅ Sem voice cloning
- ✅ Sem prosody control
- ✅ Qualidade "boa o suficiente"

**Testes mínimos**:
- Teste com texto de exemplo
- Verificar áudio gerado

### Fase 5: Integração e CLI (2 dias)

**5.1 Implementar pipeline completo**
```go
// cmd/dubbing-mvp/main.go
func main() {
    // 1. Inicializar M6 (já existe)
    audioInterface := audio.NewCoordinator()
    
    // 2. Inicializar ASR
    asr := asrsimple.NewSimpleASR("models/whisper-tiny.bin")
    
    // 3. Inicializar Translator
    translator := translation.NewSimpleTranslator(apiKey)
    
    // 4. Inicializar TTS
    tts := tts.NewSimpleTTS("en-us-female")
    
    // 5. Pipeline loop
    for {
        // Capturar áudio (3s chunks)
        audioChunk := audioInterface.CaptureChunk(3 * time.Second)
        
        // ASR: Áudio → Texto PT
        textPT := asr.Transcribe(audioChunk)
        if textPT == "" { continue }
        
        // Translation: PT → EN
        textEN := translator.Translate(textPT)
        
        // TTS: Texto EN → Áudio EN
        audioEN := tts.Synthesize(textEN)
        
        // Playback: Enviar para dispositivo virtual
        audioInterface.PlayChunk(audioEN)
    }
}
```

**5.2 Implementar CLI básico**
```bash
# Iniciar dublagem
dubbing-mvp start

# Parar dublagem
dubbing-mvp stop

# Ver status
dubbing-mvp status

# Configurar dispositivos
dubbing-mvp config --input "Microfone" --output "Virtual Cable"
```

### Fase 6: Testes e Ajustes (1 dia)

**6.1 Testar com Google Meets**
- Abrir Google Meets
- Configurar microfone como "Virtual Cable"
- Iniciar dubbing-mvp
- Falar em português
- Verificar que outros ouvem em inglês

**6.2 Ajustar latência**
- Medir latência end-to-end
- Ajustar tamanho dos chunks
- Otimizar processamento

**6.3 Corrigir bugs**
- Tratar erros de rede (API)
- Tratar silêncios longos
- Prevenir crashes

## 🛠️ Stack Tecnológico do MVP

### Backend
- **Go 1.21+**
- **M6 Audio Interface** (já implementado)

### Modelos/APIs
- **ASR**: Whisper Tiny (via whisper.cpp)
- **Translation**: Google Translate API ou LibreTranslate
- **TTS**: Piper TTS (local) ou gTTS (API)

### Dependências Go
```go
// go.mod
module github.com/user/dubbing-mvp

require (
    github.com/ggerganov/whisper.cpp/bindings/go v0.0.0
    github.com/bregydoc/gtranslate v0.0.0
    // ou github.com/libretranslate/libretranslate-go
    github.com/rhasspy/piper-go v0.0.0
    // ou github.com/hegedustibor/htgo-tts
)
```

## 📦 Dispositivo de Áudio Virtual

Para que o Google Meets receba o áudio dublado, precisamos de um **dispositivo de áudio virtual**:

### Windows
- **VB-Audio Virtual Cable** (grátis)
- **VoiceMeeter** (grátis, mais features)

### Linux
- **PulseAudio Virtual Sink**
- **JACK Audio**

### macOS
- **BlackHole** (grátis)
- **Loopback** (pago)

### Configuração
```
1. Instalar Virtual Cable
2. Configurar dubbing-mvp:
   - Input: Microfone físico
   - Output: Virtual Cable Input
3. Configurar Google Meets:
   - Microfone: Virtual Cable Output
```

## 🎯 Critérios de Sucesso do MVP

### Funcional
- ✅ Captura áudio do microfone
- ✅ Reconhece fala em português
- ✅ Traduz para inglês
- ✅ Sintetiza voz em inglês
- ✅ Envia para dispositivo virtual
- ✅ Funciona em Google Meets

### Performance
- ✅ Latência < 2 segundos
- ✅ Não crashar durante 10 minutos
- ✅ CPU < 50%
- ✅ RAM < 1GB

### Qualidade
- ✅ Tradução compreensível (não perfeita)
- ✅ Voz sintética clara
- ✅ Sem cortes ou glitches graves

## 📅 Timeline do MVP

| Fase | Duração | Entregável |
|------|---------|------------|
| 1. Setup | 1 dia | Estrutura do projeto |
| 2. ASR | 2 dias | Reconhecimento PT funcionando |
| 3. Translation | 1 dia | Tradução PT→EN funcionando |
| 4. TTS | 2 dias | Síntese EN funcionando |
| 5. Integração | 2 dias | Pipeline completo + CLI |
| 6. Testes | 1 dia | MVP testado e ajustado |
| **TOTAL** | **9 dias** | **MVP funcional** |

## 🚀 Como Usar o MVP

### Instalação
```bash
# 1. Instalar Virtual Cable
# Windows: https://vb-audio.com/Cable/
# Linux: pulseaudio --load-module module-null-sink
# macOS: https://existential.audio/blackhole/

# 2. Clonar repositório
git clone https://github.com/user/dubbing-mvp
cd dubbing-mvp

# 3. Baixar modelos
./scripts/download-models.sh

# 4. Compilar
go build -o dubbing-mvp cmd/dubbing-mvp/main.go

# 5. Configurar
./dubbing-mvp config --input "Microfone" --output "Virtual Cable"
```

### Uso
```bash
# Iniciar dublagem
./dubbing-mvp start

# Em outra janela, abrir Google Meets
# Configurar microfone: "Virtual Cable Output"

# Falar em português → Outros ouvem em inglês!

# Parar dublagem
./dubbing-mvp stop
```

## 🔧 Troubleshooting

### Problema: Latência muito alta (> 3s)
**Solução**: Reduzir chunk size de 3s para 2s

### Problema: Qualidade ruim
**Solução**: Usar Whisper Small em vez de Tiny

### Problema: API rate limit
**Solução**: Usar LibreTranslate (self-hosted)

### Problema: Áudio cortado
**Solução**: Aumentar buffer size no M6

## 📈 Evolução Pós-MVP

Após o MVP funcionar, adicionar incrementalmente:

### Versão 1.1 (+ 1 semana)
- ✅ Interface gráfica básica (System Tray)
- ✅ Configuração via UI
- ✅ Indicador de status

### Versão 1.2 (+ 2 semanas)
- ✅ Voice cloning básico
- ✅ Context window (3 sentenças)
- ✅ Melhor qualidade de tradução

### Versão 2.0 (+ 1 mês)
- ✅ Prosody transfer
- ✅ Perfis de uso
- ✅ Dashboard de métricas
- ✅ Todas as features planejadas

## ✅ Checklist de Implementação

### Semana 1
- [ ] Dia 1: Setup do projeto
- [ ] Dia 2-3: Implementar ASR simplificado
- [ ] Dia 4: Implementar Translation simplificado
- [ ] Dia 5: Começar TTS simplificado

### Semana 2
- [ ] Dia 1: Finalizar TTS simplificado
- [ ] Dia 2-3: Integração completa + CLI
- [ ] Dia 4: Testes com Google Meets
- [ ] Dia 5: Ajustes finais e documentação

## 🎉 Resultado Esperado

Ao final de **9 dias úteis**, você terá:

1. ✅ Um executável `dubbing-mvp`
2. ✅ Que captura sua voz em português
3. ✅ Traduz para inglês em tempo real
4. ✅ Sintetiza voz em inglês
5. ✅ Envia para Google Meets/Zoom/Discord
6. ✅ Com latência aceitável (< 2s)
7. ✅ Qualidade "boa o suficiente"

**Você poderá participar de reuniões internacionais falando português e sendo ouvido em inglês!** 🚀

---

**Próximo Passo**: Começar implementação da Fase 1 (Setup)
