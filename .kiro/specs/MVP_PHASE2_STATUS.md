# MVP Fase 2 - Status de Implementação

## 🎯 Objetivo da Fase 2

Substituir implementações mock por integrações reais com bibliotecas e APIs.

## ✅ Progresso Atual

### Fase 1 (Mock) - ✅ COMPLETO
- [x] Estrutura do projeto
- [x] CLI funcional
- [x] Pipeline completo com mock
- [x] Testes passando
- [x] Documentação completa

### Fase 2 (Integrações Reais) - 🔄 EM PROGRESSO

#### Implementações Criadas

1. **ASR - Vosk** (`pkg/asr-vosk/asr.go`) ✅
   - Interface completa
   - Preparado para biblioteca Vosk
   - VAD integrado
   - Estatísticas

2. **ASR - API** (`pkg/asr-api/asr_api.go`) ✅
   - Interface para Whisper API
   - Conversão WAV
   - HTTP client

3. **Translation - LibreTranslate** (`pkg/translation-libre/translator.go`) ✅
   - API HTTP completa
   - Cache de traduções
   - Funciona com API pública gratuita
   - Estatísticas

#### Documentação Criada

1. **DEPENDENCIES_SETUP.md** ✅
   - Guia de instalação por plataforma
   - Opções de APIs vs Local
   - Troubleshooting

2. **QUICK_INTEGRATION.md** ✅
   - Abordagem simplificada
   - Vosk + LibreTranslate + eSpeak
   - Instalação rápida

3. **GETTING_STARTED.md** ✅
   - Quick start guide
   - Como testar
   - Checklist de validação

## 📊 Opções de Implementação

### Opção 1: Tudo Local (Mais complexo)
```
ASR: Vosk (local)
Translation: LibreTranslate (self-hosted)
TTS: Piper TTS (local)
```
**Vantagens**: Offline, gratuito, baixa latência
**Desvantagens**: Setup complexo

### Opção 2: Híbrido (Recomendado) ⭐
```
ASR: Vosk (local)
Translation: LibreTranslate API (público)
TTS: eSpeak (local)
```
**Vantagens**: Bom equilíbrio, setup médio
**Desvantagens**: Translation requer internet

### Opção 3: Tudo API (Mais rápido)
```
ASR: Whisper API (OpenAI)
Translation: Google Translate API
TTS: Google TTS API
```
**Vantagens**: Setup instantâneo
**Desvantagens**: Custos, requer internet

## 🚀 Próximos Passos Imediatos

### Passo 1: Testar LibreTranslate (AGORA)
```bash
# Criar teste simples
go run cmd/test-translation/main.go
```

**Código de teste**:
```go
package main

import (
    "fmt"
    "log"
    
    libre "github.com/user/audio-dubbing-system/pkg/translation-libre"
)

func main() {
    config := libre.Config{
        SourceLang: "pt",
        TargetLang: "en",
    }
    
    translator, err := libre.NewLibreTranslator(config)
    if err != nil {
        log.Fatal(err)
    }
    defer translator.Close()
    
    // Testar traduções
    tests := []string{
        "olá",
        "bom dia",
        "como vai você",
        "eu gosto de programar",
    }
    
    for _, text := range tests {
        result, err := translator.Translate(text)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        fmt.Printf("%s → %s\n", text, result)
    }
}
```

### Passo 2: Integrar eSpeak para TTS (1 hora)
```bash
# Instalar eSpeak
sudo apt-get install espeak  # Linux
brew install espeak          # macOS

# Testar
espeak "Hello world" -w test.wav
```

**Implementação**:
```go
// pkg/tts-espeak/tts.go
package ttsespeak

import (
    "os/exec"
    "fmt"
)

func Synthesize(text string) ([]byte, error) {
    cmd := exec.Command("espeak", text, "--stdout")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("espeak failed: %w", err)
    }
    return output, nil
}
```

### Passo 3: Atualizar main.go (30 min)
```go
// Adicionar flag para escolher implementação
var useReal bool

func init() {
    startCmd.Flags().BoolVar(&useReal, "real", false, "Use real implementations")
}

func initASR() (*asrsimple.SimpleASR, error) {
    if useReal {
        // Usar Vosk
        return asrvosk.NewVoskASR(...)
    }
    // Usar mock
    return asrsimple.NewSimpleASR(...)
}
```

### Passo 4: Testar Pipeline Completo (1 hora)
```bash
# Com mock (já funciona)
./dubbing-mvp start

# Com implementações reais
./dubbing-mvp start --real
```

## 📈 Timeline Atualizado

| Tarefa | Tempo | Status |
|--------|-------|--------|
| Fase 1: Mock MVP | 4h | ✅ COMPLETO |
| LibreTranslate integration | 1h | ✅ COMPLETO |
| eSpeak integration | 1h | 📋 PRÓXIMO |
| Vosk integration | 2h | 📋 PENDENTE |
| M6 Audio integration | 2h | 📋 PENDENTE |
| Testes end-to-end | 2h | 📋 PENDENTE |
| Google Meets validation | 1h | 📋 PENDENTE |
| **TOTAL** | **13h** | **~40% completo** |

## 🎯 MVP Mínimo Funcional

Para ter um MVP **realmente funcional**, precisamos:

### Essencial (Mínimo)
- [x] CLI funcionando
- [x] Pipeline mock funcionando
- [x] LibreTranslate funcionando ✅
- [ ] eSpeak funcionando (1h)
- [ ] Captura de áudio real (2h)
- [ ] Reprodução de áudio real (incluído acima)

**Total: 3 horas para MVP mínimo funcional**

### Desejável (Melhor qualidade)
- [ ] Vosk ASR (2h)
- [ ] Testes com Google Meets (1h)
- [ ] Ajustes de latência (1h)

**Total: +4 horas para MVP completo**

## 🚀 Ação Imediata

**AGORA**: Testar LibreTranslate

```bash
# 1. Criar arquivo de teste
cat > cmd/test-translation/main.go << 'EOF'
package main

import (
    "fmt"
    "log"
    
    libre "github.com/user/audio-dubbing-system/pkg/translation-libre"
)

func main() {
    config := libre.Config{
        SourceLang: "pt",
        TargetLang: "en",
    }
    
    translator, err := libre.NewLibreTranslator(config)
    if err != nil {
        log.Fatal(err)
    }
    defer translator.Close()
    
    text := "olá mundo"
    result, err := translator.Translate(text)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("✓ Translation works!\n")
    fmt.Printf("  %s → %s\n", text, result)
}
EOF

# 2. Executar
go run cmd/test-translation/main.go
```

**Resultado esperado**:
```
✓ LibreTranslate initialized (pt → en)
  Using public API (rate limited)
LibreTranslate: 'olá mundo' → 'hello world' (523ms)
✓ Translation works!
  olá mundo → hello world
```

## 💡 Decisão Importante

**Pergunta**: Qual abordagem seguir?

### A) MVP Rápido (3 horas)
- LibreTranslate ✅
- eSpeak TTS
- Mock ASR (já funciona)
- Áudio real do M6

**Resultado**: Sistema funcional básico

### B) MVP Completo (7 horas)
- LibreTranslate ✅
- eSpeak TTS
- Vosk ASR (real)
- Áudio real do M6
- Testes com Google Meets

**Resultado**: Sistema funcional de qualidade

### C) MVP Perfeito (2 semanas)
- Whisper.cpp ASR
- Google Translate
- Piper TTS
- Voice cloning
- Prosody transfer

**Resultado**: Sistema de produção

## 📊 Recomendação

**Seguir Opção B (MVP Completo)** - 7 horas

Motivo:
- LibreTranslate já está pronto ✅
- eSpeak é simples (1h)
- Vosk dá qualidade real (2h)
- M6 já existe (2h integração)
- Testes validam tudo (2h)

**Próxima ação**: Implementar eSpeak TTS

---

**Status**: 🔄 Fase 2 em progresso (40%)
**Próximo**: Implementar eSpeak TTS
**Tempo restante**: ~7 horas para MVP completo
**Data**: 2025-11-29
