# Instalação do eSpeak

## 🎯 O Que É eSpeak

eSpeak é um sintetizador de voz de código aberto, compacto e rápido. Embora a voz seja robótica, é clara e funcional para o MVP.

## 📦 Instalação por Plataforma

### Linux (Ubuntu/Debian)

```bash
# Instalar eSpeak
sudo apt-get update
sudo apt-get install espeak espeak-data

# Verificar instalação
espeak --version

# Testar
espeak "Hello world"
```

### Linux (Fedora/RHEL)

```bash
# Instalar eSpeak
sudo dnf install espeak

# Verificar
espeak --version
```

### macOS

```bash
# Instalar via Homebrew
brew install espeak

# Verificar
espeak --version

# Testar
espeak "Hello world"
```

### Windows

#### Opção 1: Instalador (Recomendado)
1. Download: http://espeak.sourceforge.net/download.html
2. Baixar `setup_espeak-1.48.04.exe`
3. Instalar
4. Adicionar ao PATH:
   - Painel de Controle → Sistema → Variáveis de Ambiente
   - Adicionar `C:\Program Files (x86)\eSpeak\command_line` ao PATH

#### Opção 2: Chocolatey
```powershell
# Instalar via Chocolatey
choco install espeak

# Verificar
espeak --version
```

#### Opção 3: Scoop
```powershell
# Instalar via Scoop
scoop install espeak

# Verificar
espeak --version
```

## ✅ Verificar Instalação

### Teste 1: Versão
```bash
espeak --version
```

**Resultado esperado**:
```
eSpeak text-to-speech: 1.48.04  04.Mar.14
```

### Teste 2: Síntese Simples
```bash
espeak "Hello world"
```

**Resultado esperado**: Você deve ouvir "Hello world" sintetizado

### Teste 3: Gerar Arquivo WAV
```bash
espeak "Hello world" -w test.wav
```

**Resultado esperado**: Arquivo `test.wav` criado

### Teste 4: Reproduzir WAV
```bash
# Linux
aplay test.wav

# macOS
afplay test.wav

# Windows
# Abrir test.wav no Windows Media Player
```

## 🎛️ Opções do eSpeak

### Vozes Disponíveis
```bash
# Listar vozes
espeak --voices

# Usar voz específica
espeak -v en-us "Hello"      # Inglês americano
espeak -v en-gb "Hello"      # Inglês britânico
espeak -v en "Hello"         # Inglês padrão
```

### Velocidade
```bash
# Velocidade padrão: 175 palavras por minuto
espeak -s 175 "Normal speed"

# Mais rápido
espeak -s 250 "Fast speech"

# Mais devagar
espeak -s 100 "Slow speech"
```

### Pitch (Tom)
```bash
# Pitch padrão: 50 (0-99)
espeak -p 50 "Normal pitch"

# Mais agudo
espeak -p 80 "High pitch"

# Mais grave
espeak -p 20 "Low pitch"
```

### Amplitude (Volume)
```bash
# Amplitude padrão: 100 (0-200)
espeak -a 100 "Normal volume"

# Mais alto
espeak -a 150 "Loud"

# Mais baixo
espeak -a 50 "Quiet"
```

## 🧪 Testar com o MVP

### Teste 1: Teste Unitário
```bash
go run cmd/test-tts/main.go
```

**Resultado esperado**:
```
🧪 Testing eSpeak TTS Integration
==================================

✓ eSpeak TTS initialized (voice: en-us, speed: 175 wpm, pitch: 50)

📝 Running TTS tests...

Test 1: 'Hello world'
  ✓ Generated: 8000 samples
  ⏱️  Time: 234ms
  🎵 Duration: 0.50s

...

✅ All tests passed!
eSpeak TTS integration is working correctly.
```

### Teste 2: Pipeline Completo
```bash
# Compilar MVP
go build -o dubbing-mvp cmd/dubbing-mvp/main.go

# Executar com eSpeak
./dubbing-mvp start --use-espeak
```

## 🐛 Troubleshooting

### Erro: "espeak: command not found"

**Linux**:
```bash
# Verificar se está instalado
which espeak

# Se não estiver, instalar
sudo apt-get install espeak
```

**macOS**:
```bash
# Verificar Homebrew
brew --version

# Instalar eSpeak
brew install espeak
```

**Windows**:
- Verificar que eSpeak está no PATH
- Reiniciar terminal após instalação

### Erro: "No audio output"

**Linux**:
```bash
# Verificar ALSA
aplay -l

# Instalar se necessário
sudo apt-get install alsa-utils
```

**macOS**:
- Verificar volume do sistema
- Verificar permissões de áudio

### Erro: "Permission denied"

```bash
# Linux/macOS
sudo chmod +x /usr/bin/espeak
```

### Qualidade de Áudio Ruim

eSpeak tem voz robótica por design. Para melhor qualidade:

**Opção 1**: Ajustar parâmetros
```bash
espeak -s 150 -p 45 "Better quality"
```

**Opção 2**: Usar alternativa (futuro)
- Piper TTS (melhor qualidade)
- Google TTS API (qualidade profissional)

## 📊 Comparação de Qualidade

| TTS | Qualidade | Velocidade | Instalação | Custo |
|-----|-----------|------------|------------|-------|
| eSpeak | ⭐⭐ Robótica | ⚡⚡⚡ Muito rápida | ✅ Fácil | 💰 Grátis |
| Piper | ⭐⭐⭐⭐ Boa | ⚡⚡ Rápida | 🔧 Média | 💰 Grátis |
| Google TTS | ⭐⭐⭐⭐⭐ Excelente | ⚡ Média | ✅ Fácil | 💰💰 Pago |

**Para MVP**: eSpeak é suficiente! ✅

## 🎯 Próximos Passos

Após instalar eSpeak:

1. ✅ Verificar instalação: `espeak --version`
2. ✅ Testar síntese: `espeak "Hello world"`
3. ✅ Executar teste: `go run cmd/test-tts/main.go`
4. ✅ Integrar no pipeline: Atualizar `main.go`
5. ✅ Testar MVP completo: `./dubbing-mvp start`

---

**Tempo de instalação**: 5-10 minutos
**Dificuldade**: ⭐ Fácil
**Status**: Pronto para usar
