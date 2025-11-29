#!/bin/bash

# Script para baixar modelos necessários para o MVP
# Whisper Tiny (~75MB) + Piper Voice (~10MB)

set -e

echo "🚀 Downloading models for Dubbing MVP..."
echo ""

# Criar diretório de modelos
mkdir -p models

# 1. Download Whisper Tiny model
echo "📥 Downloading Whisper Tiny model (~75MB)..."
if [ ! -f "models/ggml-tiny.bin" ]; then
    curl -L "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-tiny.bin" \
        -o "models/ggml-tiny.bin"
    echo "✓ Whisper Tiny downloaded"
else
    echo "✓ Whisper Tiny already exists"
fi

# 2. Download Piper TTS voice
echo ""
echo "📥 Downloading Piper TTS voice (~10MB)..."
if [ ! -f "models/en_US-lessac-medium.onnx" ]; then
    curl -L "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx" \
        -o "models/en_US-lessac-medium.onnx"
    
    curl -L "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx.json" \
        -o "models/en_US-lessac-medium.onnx.json"
    echo "✓ Piper voice downloaded"
else
    echo "✓ Piper voice already exists"
fi

echo ""
echo "✅ All models downloaded successfully!"
echo ""
echo "Models location: $(pwd)/models/"
echo "  - Whisper Tiny: models/ggml-tiny.bin"
echo "  - Piper Voice:  models/en_US-lessac-medium.onnx"
echo ""
echo "Next steps:"
echo "  1. Build: go build -o dubbing-mvp cmd/dubbing-mvp/main.go"
echo "  2. Run:   ./dubbing-mvp start"
