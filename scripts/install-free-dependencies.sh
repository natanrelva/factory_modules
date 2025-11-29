#!/bin/bash

# Script de instalação de TODAS as dependências GRATUITAS
# Para o MVP de Dublagem PT→EN

set -e

echo "🚀 Instalando dependências GRATUITAS para MVP"
echo "=============================================="
echo ""
echo "Este script instalará:"
echo "  1. Argos Translate (tradução offline gratuita)"
echo "  2. eSpeak (síntese de voz gratuita)"
echo "  3. Vosk (reconhecimento de fala gratuito - opcional)"
echo ""
echo "Tempo estimado: 5-10 minutos"
echo ""
read -p "Continuar? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

echo ""
echo "📦 Passo 1/3: Instalando Argos Translate..."
echo "==========================================="

# Verificar se Python está instalado
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 não encontrado!"
    echo "Instale com: sudo apt-get install python3 python3-pip"
    exit 1
fi

# Instalar Argos Translate
pip3 install argostranslate

# Instalar pacote PT→EN
echo "Baixando pacote de tradução PT→EN..."
python3 << 'EOF'
import argostranslate.package

print("Atualizando índice de pacotes...")
argostranslate.package.update_package_index()

print("Procurando pacote PT→EN...")
available_packages = argostranslate.package.get_available_packages()
pt_en_package = next(
    (pkg for pkg in available_packages 
     if pkg.from_code == 'pt' and pkg.to_code == 'en'),
    None
)

if pt_en_package:
    print(f"Instalando {pt_en_package}...")
    argostranslate.package.install_from_path(pt_en_package.download())
    print("✓ Pacote PT→EN instalado!")
else:
    print("❌ Pacote PT→EN não encontrado")
    exit(1)
EOF

# Testar
echo ""
echo "Testando Argos Translate..."
RESULT=$(argos-translate --from pt --to en "olá mundo")
echo "  'olá mundo' → '$RESULT'"

if [ "$RESULT" != "" ]; then
    echo "✅ Argos Translate instalado com sucesso!"
else
    echo "⚠️  Argos Translate instalado mas teste falhou"
fi

echo ""
echo "📦 Passo 2/3: Instalando eSpeak..."
echo "=================================="

# Detectar sistema operacional
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    if command -v apt-get &> /dev/null; then
        sudo apt-get install -y espeak espeak-data
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y espeak
    elif command -v yum &> /dev/null; then
        sudo yum install -y espeak
    else
        echo "⚠️  Gerenciador de pacotes não suportado"
        echo "Instale eSpeak manualmente"
    fi
elif [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    if command -v brew &> /dev/null; then
        brew install espeak
    else
        echo "❌ Homebrew não encontrado!"
        echo "Instale Homebrew: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        exit 1
    fi
else
    echo "⚠️  Sistema operacional não detectado"
    echo "Instale eSpeak manualmente"
fi

# Testar eSpeak
echo ""
echo "Testando eSpeak..."
if command -v espeak &> /dev/null; then
    espeak "Hello world" 2>/dev/null
    echo "✅ eSpeak instalado com sucesso!"
else
    echo "⚠️  eSpeak não encontrado no PATH"
fi

echo ""
echo "📦 Passo 3/3: Vosk (Opcional)..."
echo "================================"
echo ""
echo "Vosk é opcional mas recomendado para melhor reconhecimento de fala."
echo "Você pode instalar agora ou pular e usar mock ASR."
echo ""
read -p "Instalar Vosk? (y/n) " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Baixando modelo Vosk PT (~50MB)..."
    
    mkdir -p models
    cd models
    
    if [ ! -d "vosk-model-small-pt-0.3" ]; then
        curl -LO https://alphacephei.com/vosk/models/vosk-model-small-pt-0.3.zip
        unzip vosk-model-small-pt-0.3.zip
        rm vosk-model-small-pt-0.3.zip
        echo "✅ Modelo Vosk baixado!"
    else
        echo "✓ Modelo Vosk já existe"
    fi
    
    cd ..
else
    echo "⏭️  Pulando Vosk (usará mock ASR)"
fi

echo ""
echo "🎉 Instalação Completa!"
echo "======================"
echo ""
echo "✅ Argos Translate (tradução offline gratuita)"
echo "✅ eSpeak (síntese de voz gratuita)"
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "✅ Vosk (reconhecimento de fala gratuito)"
fi
echo ""
echo "📊 Resumo:"
echo "  Custo total: R$ 0,00 💰"
echo "  Funciona offline: ✅"
echo "  Privacidade: 100% ✅"
echo ""
echo "🚀 Próximos passos:"
echo "  1. Compilar: go build -o dubbing-mvp cmd/dubbing-mvp/main.go"
echo "  2. Testar tradução: go run cmd/test-translation/main.go"
echo "  3. Testar TTS: go run cmd/test-tts/main.go"
echo "  4. Executar MVP: ./dubbing-mvp start"
echo ""
echo "✨ Tudo pronto para usar!"
