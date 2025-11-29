# Instalação do Argos Translate - 100% GRATUITO

## 🎯 O Que É Argos Translate

**Argos Translate** é um tradutor de código aberto que funciona **completamente offline** e é **100% gratuito**!

### ✅ Vantagens
- ✅ **100% GRATUITO** - Sem custos, sem API keys
- ✅ **Funciona OFFLINE** - Não precisa de internet
- ✅ **Código aberto** - MIT License
- ✅ **Boa qualidade** - Baseado em OpenNMT
- ✅ **Privacidade** - Seus dados não saem do computador

### ⚖️ Comparação

| Solução | Custo | Internet | Qualidade | Privacidade |
|---------|-------|----------|-----------|-------------|
| **Argos Translate** | ✅ Grátis | ❌ Não precisa | ⭐⭐⭐ Boa | ✅ Total |
| LibreTranslate API | ⚠️ Rate limited | ✅ Precisa | ⭐⭐⭐⭐ Muito boa | ⚠️ Parcial |
| Google Translate | 💰 Pago | ✅ Precisa | ⭐⭐⭐⭐⭐ Excelente | ❌ Nenhuma |

**Para MVP**: Argos Translate é PERFEITO! ✅

## 📦 Instalação

### Pré-requisitos

Você precisa do Python 3.8+ instalado:

```bash
# Verificar Python
python3 --version

# Se não tiver, instalar:
# Ubuntu/Debian
sudo apt-get install python3 python3-pip

# macOS
brew install python3

# Windows
# Download: https://www.python.org/downloads/
```

### Passo 1: Instalar Argos Translate

```bash
# Instalar via pip
pip install argostranslate

# Ou com pip3
pip3 install argostranslate
```

### Passo 2: Instalar Pacote de Tradução PT→EN

```bash
# Método 1: Via argospm (recomendado)
argospm install translate-pt_en

# Método 2: Via Python
python3 -c "import argostranslate.package; argostranslate.package.update_package_index(); [pkg.install() for pkg in argostranslate.package.get_available_packages() if pkg.from_code == 'pt' and pkg.to_code == 'en']"
```

### Passo 3: Verificar Instalação

```bash
# Testar tradução
argos-translate --from pt --to en "olá mundo"

# Resultado esperado: "hello world"
```

## 🚀 Instalação Rápida (Script Completo)

### Linux/macOS

```bash
#!/bin/bash

# Instalar Argos Translate
pip3 install argostranslate

# Baixar e instalar pacote PT→EN
python3 << 'EOF'
import argostranslate.package

# Atualizar índice de pacotes
argostranslate.package.update_package_index()

# Encontrar e instalar pacote PT→EN
available_packages = argostranslate.package.get_available_packages()
pt_en_package = next(
    (pkg for pkg in available_packages 
     if pkg.from_code == 'pt' and pkg.to_code == 'en'),
    None
)

if pt_en_package:
    argostranslate.package.install_from_path(pt_en_package.download())
    print("✓ Pacote PT→EN instalado com sucesso!")
else:
    print("❌ Pacote PT→EN não encontrado")
EOF

# Testar
argos-translate --from pt --to en "olá mundo"
```

### Windows (PowerShell)

```powershell
# Instalar Argos Translate
pip install argostranslate

# Instalar pacote PT→EN
python -c "import argostranslate.package; argostranslate.package.update_package_index(); [pkg.install() for pkg in argostranslate.package.get_available_packages() if pkg.from_code == 'pt' and pkg.to_code == 'en']"

# Testar
argos-translate --from pt --to en "olá mundo"
```

## ✅ Verificar Instalação

### Teste 1: Comando CLI

```bash
argos-translate --from pt --to en "olá"
# Esperado: hello

argos-translate --from pt --to en "bom dia"
# Esperado: good morning

argos-translate --from pt --to en "como vai você"
# Esperado: how are you
```

### Teste 2: Python

```python
import argostranslate.translate

# Traduzir
text = "olá mundo"
result = argostranslate.translate.translate(text, "pt", "en")
print(f"{text} → {result}")
# Esperado: olá mundo → hello world
```

### Teste 3: Com o MVP

```bash
go run cmd/test-translation/main.go
```

## 🎛️ Uso Avançado

### Listar Pacotes Instalados

```python
import argostranslate.package

installed_packages = argostranslate.package.get_installed_packages()
for pkg in installed_packages:
    print(f"{pkg.from_name} → {pkg.to_name}")
```

### Instalar Outros Idiomas

```bash
# Inglês → Espanhol
argospm install translate-en_es

# Francês → Inglês
argospm install translate-fr_en

# Ver todos disponíveis
argospm list
```

### Usar em Python

```python
import argostranslate.translate

# Traduzir texto
def translate(text, from_lang="pt", to_lang="en"):
    return argostranslate.translate.translate(text, from_lang, to_lang)

# Exemplo
print(translate("olá mundo"))  # hello world
print(translate("obrigado"))   # thank you
```

## 🐛 Troubleshooting

### Erro: "argos-translate: command not found"

**Solução 1**: Adicionar ao PATH

```bash
# Linux/macOS
export PATH="$PATH:$HOME/.local/bin"

# Adicionar ao ~/.bashrc ou ~/.zshrc
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
```

**Solução 2**: Usar Python diretamente

```bash
python3 -m argostranslate.cli --from pt --to en "olá"
```

### Erro: "No package found for pt→en"

```bash
# Atualizar índice de pacotes
python3 << 'EOF'
import argostranslate.package
argostranslate.package.update_package_index()
EOF

# Tentar instalar novamente
argospm install translate-pt_en
```

### Erro: "pip: command not found"

```bash
# Ubuntu/Debian
sudo apt-get install python3-pip

# macOS
brew install python3

# Windows
# Reinstalar Python com pip incluído
```

### Tradução retorna vazio

```bash
# Verificar pacotes instalados
python3 -c "import argostranslate.package; print([f'{p.from_code}→{p.to_code}' for p in argostranslate.package.get_installed_packages()])"

# Se não mostrar 'pt→en', reinstalar
argospm install translate-pt_en
```

## 📊 Qualidade Esperada

### Exemplos de Tradução

| Português | Argos Translate | Google Translate |
|-----------|-----------------|------------------|
| olá | hello | hello |
| bom dia | good morning | good morning |
| como vai você | how are you | how are you |
| eu gosto de programar | I like to program | I like to program |
| obrigado | thank you | thank you |

**Qualidade**: ⭐⭐⭐ Boa (suficiente para MVP!)

### Performance

- **Latência**: ~100-300ms por sentença
- **Throughput**: ~5-10 sentenças/segundo
- **Memória**: ~200MB (modelo carregado)
- **CPU**: Baixo uso

## 🎯 Integração com MVP

### Atualizar main.go

```go
import (
    argos "github.com/user/audio-dubbing-system/pkg/translation-argos"
)

func initTranslator() (*argos.ArgosTranslator, error) {
    config := argos.Config{
        SourceLang: "pt",
        TargetLang: "en",
    }
    
    return argos.NewArgosTranslator(config)
}
```

### Executar MVP

```bash
# Compilar
go build -o dubbing-mvp cmd/dubbing-mvp/main.go

# Executar com Argos Translate
./dubbing-mvp start --use-argos
```

## 🚀 Próximos Passos

Após instalar Argos Translate:

1. ✅ Instalar: `pip install argostranslate`
2. ✅ Instalar pacote: `argospm install translate-pt_en`
3. ✅ Testar: `argos-translate --from pt --to en "olá"`
4. ✅ Testar com MVP: `go run cmd/test-translation/main.go`
5. ✅ Usar no pipeline: `./dubbing-mvp start`

## 💡 Dicas

### Melhorar Qualidade

```bash
# Usar modelos maiores (se disponível)
argospm install translate-pt_en-large

# Pré-processar texto
# - Remover pontuação extra
# - Normalizar espaços
# - Dividir em sentenças curtas
```

### Otimizar Performance

```python
# Carregar modelo uma vez e reusar
import argostranslate.translate

# Modelo fica em memória
translator = argostranslate.translate.get_translator("pt", "en")

# Traduzir múltiplas vezes (rápido)
for text in texts:
    result = translator.translate(text)
```

### Cache de Traduções

O MVP já implementa cache automático para evitar traduzir o mesmo texto múltiplas vezes!

---

## ✅ Resumo

**Argos Translate é a solução PERFEITA para o MVP**:

- ✅ 100% gratuito
- ✅ Funciona offline
- ✅ Boa qualidade
- ✅ Fácil de instalar
- ✅ Privacidade total

**Tempo de instalação**: 5-10 minutos
**Custo**: R$ 0,00
**Qualidade**: Suficiente para MVP

---

**Próximo comando**:
```bash
pip install argostranslate
argospm install translate-pt_en
argos-translate --from pt --to en "olá mundo"
```
