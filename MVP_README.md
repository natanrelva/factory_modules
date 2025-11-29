# Dubbing MVP - Real-time PT→EN Translation

## 🎯 O Que É

Um sistema **mínimo viável** que permite falar em **Português** e ser ouvido em **Inglês** em tempo real no Google Meets, Zoom, Discord e outras aplicações.

## ✨ Features do MVP

- ✅ Captura áudio do microfone
- ✅ Reconhece fala em português (Whisper)
- ✅ Traduz para inglês (Google Translate)
- ✅ Sintetiza voz em inglês (Piper TTS)
- ✅ Envia para dispositivo virtual
- ✅ CLI simples para controle

## 🚀 Quick Start

### 1. Pré-requisitos

#### Windows
```bash
# Instalar Virtual Cable
# Download: https://vb-audio.com/Cable/
```

#### Linux
```bash
# Criar dispositivo virtual com PulseAudio
pactl load-module module-null-sink sink_name=virtual_cable
```

#### macOS
```bash
# Instalar BlackHole
brew install blackhole-2ch
```

### 2. Instalação

```bash
# Clonar repositório
git clone https://github.com/user/audio-dubbing-system
cd audio-dubbing-system

# Baixar modelos
./scripts/download-models.sh

# Compilar
go build -o dubbing-mvp cmd/dubbing-mvp/main.go
```

### 3. Configuração

```bash
# Listar dispositivos disponíveis
./dubbing-mvp devices

# Configurar dispositivos
./dubbing-mvp config \
  --input "Microfone" \
  --output "Virtual Cable Input"
```

### 4. Uso

```bash
# Iniciar dublagem
./dubbing-mvp start

# Em outra janela, abrir Google Meets
# Configurar microfone: "Virtual Cable Output"

# Falar em português → Outros ouvem em inglês! 🎉

# Parar dublagem (Ctrl+C)
```

## 📊 Status Atual

### Implementado
- ✅ Estrutura do projeto
- ✅ CLI básico (cobra)
- ✅ Interfaces dos módulos simplificados
- ✅ M6 Audio Interface (já existia)

### Em Desenvolvimento
- 🔄 M2 ASR (Whisper integration)
- 🔄 M3 Translation (Google Translate API)
- 🔄 M4 TTS (Piper TTS integration)
- 🔄 Pipeline completo

### Próximos Passos
1. Integrar Whisper.cpp para ASR
2. Integrar Google Translate API
3. Integrar Piper TTS
4. Conectar com M6 Audio Interface
5. Testar com Google Meets

## 🛠️ Arquitetura do MVP

```
Microfone → M6 Capture → M2 ASR → M3 Translation → M4 TTS → M6 Playback → Virtual Device → Google Meets
   (PT)        PCM       Text PT      Text EN        PCM EN      Audio EN         (EN)
```

## 📦 Dependências

### Go Packages
- `github.com/spf13/cobra` - CLI framework
- TODO: Whisper.cpp bindings
- TODO: Google Translate client
- TODO: Piper TTS bindings

### Modelos ML
- Whisper Tiny (~75MB) - ASR
- Piper voice (~10MB) - TTS

### Sistema
- Virtual Audio Cable (Windows/macOS)
- PulseAudio (Linux)

## ⚙️ Configuração

### Arquivo de Configuração
```yaml
# ~/.dubbing-mvp/config.yaml
input_device: "Microfone"
output_device: "Virtual Cable Input"
chunk_size: 3  # segundos
asr:
  model: "models/whisper-tiny.bin"
  language: "pt"
translation:
  api_key: "YOUR_GOOGLE_TRANSLATE_API_KEY"
  source: "pt"
  target: "en"
tts:
  voice: "en-us-female"
  engine: "piper"
```

## 🎮 Comandos

```bash
# Iniciar dublagem
dubbing-mvp start

# Iniciar com configurações customizadas
dubbing-mvp start --chunk-size 2 --api-key "YOUR_KEY"

# Ver status
dubbing-mvp status

# Listar dispositivos
dubbing-mvp devices

# Configurar
dubbing-mvp config --input "Mic" --output "Virtual"

# Ver versão
dubbing-mvp --version

# Ajuda
dubbing-mvp --help
```

## 📈 Performance Esperada

| Métrica | Target MVP | Observação |
|---------|------------|------------|
| Latência | < 2s | Aceitável para MVP |
| CPU | < 50% | Em máquina moderna |
| RAM | < 1GB | Com modelos pequenos |
| Qualidade | "Compreensível" | Não perfeito, mas funcional |

## 🐛 Troubleshooting

### Problema: "No audio devices found"
**Solução**: Instalar Virtual Cable e reiniciar

### Problema: Latência muito alta
**Solução**: Reduzir chunk-size para 2 segundos

### Problema: Qualidade ruim
**Solução**: Usar Whisper Small em vez de Tiny (mais lento)

### Problema: API rate limit
**Solução**: Usar LibreTranslate (self-hosted, grátis)

## 🔧 Desenvolvimento

### Estrutura do Código
```
audio-dubbing-system/
├── cmd/
│   └── dubbing-mvp/
│       └── main.go              # CLI principal
├── pkg/
│   ├── asr-simple/
│   │   └── asr.go               # ASR simplificado
│   ├── translation-simple/
│   │   └── translator.go        # Translation simplificado
│   └── tts-simple/
│       └── tts.go               # TTS simplificado
├── audio-interface/             # M6 (já implementado)
├── scripts/
│   └── download-models.sh       # Download de modelos
├── go.mod
└── MVP_README.md
```

### Compilar
```bash
go build -o dubbing-mvp cmd/dubbing-mvp/main.go
```

### Testar
```bash
go test ./...
```

### Executar em modo debug
```bash
go run cmd/dubbing-mvp/main.go start --verbose
```

## 📝 Roadmap

### MVP (9 dias) - ATUAL
- [x] Estrutura do projeto
- [x] CLI básico
- [ ] ASR integration (Whisper)
- [ ] Translation integration
- [ ] TTS integration
- [ ] Pipeline completo
- [ ] Testes com Google Meets

### v1.1 (+ 1 semana)
- [ ] Interface gráfica (System Tray)
- [ ] Configuração via UI
- [ ] Indicador de status

### v1.2 (+ 2 semanas)
- [ ] Voice cloning básico
- [ ] Context window
- [ ] Melhor qualidade

### v2.0 (+ 1 mês)
- [ ] Todas as features planejadas
- [ ] Prosody transfer
- [ ] Perfis de uso
- [ ] Dashboard de métricas

## 🤝 Contribuindo

Este é um MVP em desenvolvimento ativo. Contribuições são bem-vindas!

1. Fork o projeto
2. Crie uma branch (`git checkout -b feature/amazing`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing`)
5. Abra um Pull Request

## 📄 Licença

[A definir]

## 🙏 Agradecimentos

- OpenAI Whisper - ASR
- Google Translate - Translation
- Piper TTS - Text-to-Speech
- VB-Audio - Virtual Cable

---

**Status**: 🚧 Em Desenvolvimento Ativo
**Versão**: 0.1.0-mvp
**Última Atualização**: 2025-11-29
