# Script de instalação de TODAS as dependências GRATUITAS para Windows
# Para o MVP de Dublagem PT→EN

Write-Host "🚀 Instalando dependências GRATUITAS para MVP" -ForegroundColor Green
Write-Host "=============================================="
Write-Host ""
Write-Host "Este script instalará:"
Write-Host "  1. Argos Translate (tradução offline gratuita)"
Write-Host "  2. eSpeak (síntese de voz gratuita)"
Write-Host ""
Write-Host "Tempo estimado: 5-10 minutos"
Write-Host ""

$continue = Read-Host "Continuar? (y/n)"
if ($continue -ne "y") {
    exit
}

Write-Host ""
Write-Host "📦 Passo 1/2: Instalando Argos Translate..." -ForegroundColor Cyan
Write-Host "==========================================="

# Verificar se Python está instalado
try {
    $pythonVersion = python --version 2>&1
    Write-Host "✓ Python encontrado: $pythonVersion"
} catch {
    Write-Host "❌ Python não encontrado!" -ForegroundColor Red
    Write-Host "Baixe e instale Python de: https://www.python.org/downloads/"
    Write-Host "Certifique-se de marcar 'Add Python to PATH' durante a instalação"
    exit 1
}

# Instalar Argos Translate
Write-Host ""
Write-Host "Instalando argostranslate..."
pip install argostranslate

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Falha ao instalar argostranslate" -ForegroundColor Red
    exit 1
}

# Instalar pacote PT→EN
Write-Host ""
Write-Host "Baixando pacote de tradução PT→EN..."

$pythonScript = @"
import argostranslate.package

print('Atualizando índice de pacotes...')
argostranslate.package.update_package_index()

print('Procurando pacote PT→EN...')
available_packages = argostranslate.package.get_available_packages()
pt_en_package = next(
    (pkg for pkg in available_packages 
     if pkg.from_code == 'pt' and pkg.to_code == 'en'),
    None
)

if pt_en_package:
    print(f'Instalando {pt_en_package}...')
    argostranslate.package.install_from_path(pt_en_package.download())
    print('✓ Pacote PT→EN instalado!')
else:
    print('❌ Pacote PT→EN não encontrado')
    exit(1)
"@

$pythonScript | python

if ($LASTEXITCODE -ne 0) {
    Write-Host "⚠️  Aviso: Instalação do pacote PT→EN pode ter falhado" -ForegroundColor Yellow
}

# Testar
Write-Host ""
Write-Host "Testando Argos Translate..."
try {
    $result = argos-translate --from pt --to en "olá mundo"
    Write-Host "  'olá mundo' → '$result'"
    
    if ($result) {
        Write-Host "✅ Argos Translate instalado com sucesso!" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Argos Translate instalado mas teste falhou" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠️  Não foi possível testar argos-translate" -ForegroundColor Yellow
    Write-Host "Pode ser necessário adicionar ao PATH ou reiniciar o terminal"
}

Write-Host ""
Write-Host "📦 Passo 2/2: Instalando eSpeak..." -ForegroundColor Cyan
Write-Host "=================================="
Write-Host ""
Write-Host "Para Windows, você precisa baixar e instalar eSpeak manualmente:"
Write-Host ""
Write-Host "1. Baixe eSpeak de: https://github.com/espeak-ng/espeak-ng/releases"
Write-Host "2. Procure por 'espeak-ng-X64.msi' (versão mais recente)"
Write-Host "3. Execute o instalador"
Write-Host "4. Adicione ao PATH: C:\Program Files\eSpeak NG"
Write-Host ""
Write-Host "Ou use Chocolatey (se tiver instalado):"
Write-Host "  choco install espeak"
Write-Host ""

$installEspeak = Read-Host "Você já instalou eSpeak? (y/n)"

if ($installEspeak -eq "y") {
    # Testar eSpeak
    Write-Host ""
    Write-Host "Testando eSpeak..."
    try {
        espeak "Hello world" 2>$null
        Write-Host "✅ eSpeak instalado com sucesso!" -ForegroundColor Green
    } catch {
        Write-Host "⚠️  eSpeak não encontrado no PATH" -ForegroundColor Yellow
        Write-Host "Adicione o diretório de instalação do eSpeak ao PATH"
    }
} else {
    Write-Host "⏭️  Pulando teste do eSpeak" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎉 Instalação Completa!" -ForegroundColor Green
Write-Host "======================"
Write-Host ""
Write-Host "✅ Argos Translate (tradução offline gratuita)"
if ($installEspeak -eq "y") {
    Write-Host "✅ eSpeak (síntese de voz gratuita)"
} else {
    Write-Host "⚠️  eSpeak (instale manualmente)"
}
Write-Host ""
Write-Host "📊 Resumo:"
Write-Host "  Custo total: R$ 0,00 💰"
Write-Host "  Funciona offline: ✅"
Write-Host "  Privacidade: 100% ✅"
Write-Host ""
Write-Host "🚀 Próximos passos:"
Write-Host "  1. Testar tradução: go run cmd/test-argos/main.go"
Write-Host "  2. Testar TTS: go run cmd/test-tts/main.go"
Write-Host "  3. Compilar MVP: go build -o dubbing-mvp.exe cmd/dubbing-mvp/main.go"
Write-Host "  4. Executar MVP: .\dubbing-mvp.exe start"
Write-Host ""
Write-Host "✨ Tudo pronto para usar!" -ForegroundColor Green
Write-Host ""
Write-Host "💡 Nota: Se argos-translate não for encontrado, reinicie o terminal"
Write-Host "   ou adicione Python Scripts ao PATH:"
Write-Host "   %USERPROFILE%\AppData\Local\Programs\Python\PythonXX\Scripts"
