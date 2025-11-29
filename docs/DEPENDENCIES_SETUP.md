# Configuração de Dependências - MVP

## 📦 Dependências Necessárias

Para o MVP funcionar com reconhecimento, tradução e síntese reais, precisamos instalar:

1. **Whisper.cpp** - Para ASR (reconhecimento de fala)
2. **Google Translate** ou **LibreTranslate** - Para tradução
3. **Piper TTS** ou **eSpeak** - Para síntese de voz

## 🔧 Instalação por Plataforma

### Windows

#### 1. Whisper.cpp
```powershell
# Instalar dependências
# Requer: Visual Studio Build Tools ou MinGW

# Clonar whisper.cpp
git clone https://github.com/ggerganov/whisper.cpp.git third_party/whisper.cpp
cd third_party/whisper.cpp

# Compilar
mkdir build
cd build
cmake ..
cmake --build . --config Release

# Baixar modelo
cd ../models
./download-ggml-model.sh tiny
```

#### 2. Piper TTS
```powershell
# Baixar release pré-compilado
# https://github.com/rhasspy/piper/releases

# Ou instalar via pip
pip install piper-tts

# Baixar modelo de voz
piper --download-voice en_US-lessac-medium
```

#### 3. Google Translate (Opcional)
```powershell
# Criar conta Google Cloud
# Ativar Translation API
# Obter API key
# Configurar: set GOOGLE_TRANSLATE_API_KEY=your_key_here
```

### Linux (Ubuntu/Debian)

#### 1. Whisper.cpp
```bash
# Instalar dependências
sudo apt-get update
sudo apt-get install build-essential cmake git

# Clonar e compilar
git clone https://github.com/ggerganov/whisper.cpp.git third_party/whisper.cpp
cd third_party/whisper.cpp
make

# Baixar modelo
bash ./models/download-ggml-model.sh tiny
```

#### 2. Piper TTS
```bash
# Instalar via pip
pip3 install piper-tts

# Ou baixar release
wget https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz
tar -xzf piper_amd64.tar.gz

# Baixar modelo
./piper --download-voice en_US-lessac-medium
```

#### 3. eSpeak (Alternativa mais simples)
```bash
# Instalar eSpeak
sudo apt-get install espeak espeak-data

# Testar
espeak "Hello world"
```

### macOS

#### 1. Whisper.cpp
```bash
# Instalar Homebrew se necessário
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Instalar dependências
brew install cmake

# Clonar e compilar
git clone https://github.com/ggerganov/whisper.cpp.git third_party/whisper.cpp
cd third_party/whisper.cpp
make

# Baixar modelo
bash ./models/download-ggml-model.sh tiny
```

#### 2. Piper TTS
```bash
# Instalar via pip
pip3 install piper-tts

# Baixar modelo
piper --download-voice en_US-lessac-medium
```

## 🚀 Alternativa Rápida: Usar APIs Online

Se a instalação local for complexa, podemos usar APIs:

### 1. Whisper API (OpenAI)
```bash
# Requer API key da OpenAI
export OPENAI_API_KEY=your_key_here
```

### 2. Google Translate API
```bash
# Requer API key do Google Cloud
export GOOGLE_TRANSLATE_API_KEY=your_key_here
```

### 3. Google TTS API
```bash
# Usa mesma API key do Google Cloud
export GOOGLE_TTS_API_KEY=your_key_here
```

## 📋 Verificar Instalação

### Whisper.cpp
```bash
cd third_party/whisper.cpp
./main -m models/ggml-tiny.bin -f samples/jfk.wav
```

### Piper TTS
```bash
echo "Hello world" | piper --model en_US-lessac-medium --output_file test.wav
```

### eSpeak
```bash
espeak "Hello world" -w test.wav
```

## 🎯 Configuração Recomendada para MVP

Para começar rapidamente, recomendo:

### Opção 1: Tudo Local (Melhor para produção)
- ✅ Whisper.cpp (local)
- ✅ LibreTranslate (self-hosted ou API gratuita)
- ✅ Piper TTS (local)

**Vantagens**: Sem custos, funciona offline, baixa latência
**Desvantagens**: Setup mais complexo

### Opção 2: Híbrido (Recomendado para MVP)
- ✅ Whisper.cpp (local) - Rápido e gratuito
- ✅ Google Translate API - Melhor qualidade
- ✅ Piper TTS (local) - Rápido e gratuito

**Vantagens**: Boa qualidade, setup médio
**Desvantagens**: Requer API key (mas tem free tier)

### Opção 3: Tudo API (Mais rápido para testar)
- ✅ Whisper API (OpenAI)
- ✅ Google Translate API
- ✅ Google TTS API

**Vantagens**: Setup instantâneo, melhor qualidade
**Desvantagens**: Custos por uso, requer internet

## 🔑 Configurar API Keys

Criar arquivo `.env` na raiz do projeto:

```bash
# .env
OPENAI_API_KEY=sk-...
GOOGLE_TRANSLATE_API_KEY=AIza...
GOOGLE_TTS_API_KEY=AIza...
```

Ou configurar variáveis de ambiente:

```bash
# Linux/macOS
export OPENAI_API_KEY=sk-...
export GOOGLE_TRANSLATE_API_KEY=AIza...

# Windows
set OPENAI_API_KEY=sk-...
set GOOGLE_TRANSLATE_API_KEY=AIza...
```

## 📝 Próximos Passos

Após instalar as dependências:

1. Atualizar `go.mod` com bibliotecas Go necessárias
2. Implementar integrações reais em cada módulo
3. Testar cada módulo individualmente
4. Testar pipeline completo
5. Validar com Google Meets

## 🐛 Troubleshooting

### Whisper.cpp não compila
**Solução**: Verificar que tem compilador C++ instalado
```bash
# Linux
sudo apt-get install g++

# macOS
xcode-select --install

# Windows
# Instalar Visual Studio Build Tools
```

### Piper TTS não encontrado
**Solução**: Adicionar ao PATH
```bash
export PATH=$PATH:/path/to/piper
```

### API key inválida
**Solução**: Verificar que a API está ativada no console do Google Cloud

---

**Próximo**: Implementar integrações reais nos módulos
