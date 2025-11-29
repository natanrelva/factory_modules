# Instalação do PortAudio - Captura de Áudio Real

## 🎯 O Que É PortAudio

**PortAudio** é uma biblioteca de áudio multiplataforma que permite capturar e reproduzir áudio em tempo real.

### ✅ Vantagens
- ✅ **Multiplataforma** - Windows, Linux, macOS
- ✅ **Baixa latência** - Ideal para tempo real
- ✅ **Gratuito** - MIT License
- ✅ **Maduro** - Usado em muitos projetos

### ⚠️ Desvantagens
- ⚠️ Requer instalação nativa
- ⚠️ Compilação mais complexa

## 📦 Instalação

### Windows

#### Opção 1: Chocolatey (Recomendado)
```powershell
# Instalar Chocolatey (se não tiver)
# Ver: https://chocolatey.org/install

# Instalar PortAudio
choco install portaudio
```

#### Opção 2: Manual
1. Baixar PortAudio: http://www.portaudio.com/download.html
2. Extrair para `C:\portaudio`
3. Adicionar ao PATH: `C:\portaudio\bin`
4. Copiar DLLs para `C:\Windows\System32`

#### Opção 3: MSYS2 (Para desenvolvedores)
```bash
# Instalar MSYS2: https://www.msys2.org/
# Abrir MSYS2 terminal

pacman -S mingw-w64-x86_64-portaudio
```

### Linux

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install portaudio19-dev
```

#### Fedora/RHEL
```bash
sudo dnf install portaudio-devel
```

#### Arch Linux
```bash
sudo pacman -S portaudio
```

### macOS

```bash
# Com Homebrew
brew install portaudio
```

## 🔧 Instalação do Binding Go

Após instalar PortAudio nativo:

```bash
# Instalar binding Go
go get github.com/gordonklaus/portaudio

# Atualizar dependências
go mod tidy
```

## 🚀 Compilar com PortAudio

### Compilação Normal (Sem PortAudio)
```bash
go build -o dubbing-mvp cmd/dubbing-mvp/main.go
```
**Resultado**: Usa áudio simulado

### Compilação com PortAudio
```bash
go build -tags portaudio -o dubbing-mvp cmd/dubbing-mvp/main.go
```
**Resultado**: Usa captura real do microfone

## ✅ Verificar Instalação

### Teste 1: PortAudio Nativo

**Windows**:
```powershell
# Verificar se DLL existe
Test-Path "C:\Windows\System32\portaudio_x64.dll"
```

**Linux**:
```bash
# Verificar se biblioteca existe
ldconfig -p | grep portaudio
```

**macOS**:
```bash
# Verificar se biblioteca existe
brew list portaudio
```

### Teste 2: Binding Go

```bash
# Tentar compilar com portaudio
go build -tags portaudio -o test-audio cmd/dubbing-mvp/main.go
```

**Se funcionar**: ✅ PortAudio instalado corretamente
**Se falhar**: ⚠️ PortAudio não instalado ou não encontrado

## 🎯 Uso no MVP

### Sem PortAudio (Padrão)
```bash
# Compila normalmente
go build -o dubbing-mvp cmd/dubbing-mvp/main.go

# Usa áudio simulado
./dubbing-mvp start --use-argos --use-real-audio
```

**Resultado**: Áudio simulado (para testes)

### Com PortAudio (Captura Real)
```bash
# Compila com tag portaudio
go build -tags portaudio -o dubbing-mvp cmd/dubbing-mvp/main.go

# Usa captura real do microfone
./dubbing-mvp start --use-argos --use-real-audio
```

**Resultado**: Captura real do microfone! 🎙️

## 📊 Comparação

| Modo | Instalação | Captura | Uso |
|------|------------|---------|-----|
| **Simulado** | ✅ Zero | ⚠️ Fake | Testes |
| **PortAudio** | 🔧 Requer | ✅ Real | Produção |

## 🐛 Troubleshooting

### Erro: "portaudio.h: No such file or directory"

**Problema**: PortAudio não instalado

**Solução**:
```bash
# Linux
sudo apt-get install portaudio19-dev

# macOS
brew install portaudio

# Windows
choco install portaudio
```

### Erro: "undefined: portaudio"

**Problema**: Não compilou com tag portaudio

**Solução**:
```bash
go build -tags portaudio -o dubbing-mvp cmd/dubbing-mvp/main.go
```

### Erro: "cannot find -lportaudio"

**Problema**: Biblioteca não encontrada pelo linker

**Solução Linux**:
```bash
# Atualizar cache de bibliotecas
sudo ldconfig

# Verificar se está instalada
ldconfig -p | grep portaudio
```

**Solução Windows**:
```powershell
# Adicionar ao PATH
$env:PATH = "C:\portaudio\bin;$env:PATH"

# Ou copiar DLL para System32
Copy-Item "C:\portaudio\bin\portaudio_x64.dll" "C:\Windows\System32\"
```

### Erro: "No audio devices found"

**Problema**: Sem dispositivos de áudio

**Solução**:
1. Verificar se microfone está conectado
2. Verificar permissões de áudio
3. Testar com outro aplicativo (ex: gravador do Windows)

### Áudio com ruído/distorção

**Problema**: Buffer size ou sample rate incorreto

**Solução**:
```go
// Ajustar configuração
config := audiocapture.Config{
    SampleRate: 44100,  // Tentar sample rate maior
    Channels:   1,
    BufferSize: 2048,   // Tentar buffer maior
}
```

## 📈 Performance

### Sem PortAudio (Simulado)
- **Latência**: ~0ms (instantâneo)
- **CPU**: Muito baixo
- **Qualidade**: N/A (fake)

### Com PortAudio (Real)
- **Latência**: ~10-50ms (baixa)
- **CPU**: Baixo
- **Qualidade**: Alta (áudio real)

## 🎯 Recomendação

### Para Desenvolvimento/Testes
👉 **Sem PortAudio** - Mais rápido, sem instalação

### Para Produção/Demo
👉 **Com PortAudio** - Captura real do microfone

## 📚 Recursos

### Documentação
- **PortAudio**: http://www.portaudio.com/docs/
- **Go Binding**: https://github.com/gordonklaus/portaudio

### Exemplos
- Ver: `pkg/audio-capture/capture_portaudio.go`
- Ver: `cmd/dubbing-mvp/main.go`

## 🚀 Próximos Passos

### Opção A: Usar Simulado (Agora)
```bash
# Compila normalmente
go build -o dubbing-mvp cmd/dubbing-mvp/main.go

# Executa
./dubbing-mvp start --use-argos --use-real-audio
```

### Opção B: Instalar PortAudio (30 min)
```bash
# 1. Instalar PortAudio nativo
choco install portaudio  # Windows
# ou
sudo apt-get install portaudio19-dev  # Linux
# ou
brew install portaudio  # macOS

# 2. Instalar binding Go
go get github.com/gordonklaus/portaudio
go mod tidy

# 3. Compilar com PortAudio
go build -tags portaudio -o dubbing-mvp cmd/dubbing-mvp/main.go

# 4. Executar
./dubbing-mvp start --use-argos --use-real-audio
```

## ✅ Checklist

- [ ] PortAudio nativo instalado
- [ ] Binding Go instalado
- [ ] Compila com `-tags portaudio`
- [ ] Microfone detectado
- [ ] Áudio capturado
- [ ] Pipeline funcionando

---

**Recomendação**: Para MVP rápido, use simulado. Para produção, instale PortAudio.

**Tempo de instalação**: 30 minutos
**Complexidade**: Média
**Benefício**: Captura real do microfone! 🎙️
