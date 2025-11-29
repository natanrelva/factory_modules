# Getting Started - Dubbing MVP

## 🚀 Quick Start Guide

Este guia mostra como compilar e testar o MVP do sistema de dublagem PT→EN.

## 📋 Pré-requisitos

### Software Necessário
- Go 1.21 ou superior
- Git
- (Opcional) Virtual Audio Cable para teste com Google Meets

### Verificar Instalação
```bash
go version  # Deve mostrar go1.21 ou superior
```

## 🔧 Instalação

### 1. Clonar o Repositório
```bash
git clone https://github.com/user/audio-dubbing-system
cd audio-dubbing-system
```

### 2. Baixar Dependências
```bash
go mod download
go mod tidy
```

### 3. Compilar o MVP
```bash
go build -o dubbing-mvp cmd/dubbing-mvp/main.go
```

Se a compilação for bem-sucedida, você verá o executável `dubbing-mvp` (ou `dubbing-mvp.exe` no Windows).

## 🧪 Testar o MVP

### Teste 1: Verificar CLI
```bash
./dubbing-mvp --version
./dubbing-mvp --help
```

**Resultado esperado**: Deve mostrar a versão e opções de ajuda.

### Teste 2: Listar Dispositivos
```bash
./dubbing-mvp devices
```

**Resultado esperado**: Lista de dispositivos de áudio (mock por enquanto).

### Teste 3: Executar Testes Unitários
```bash
go test ./pkg/asr-simple/...
go test ./pkg/translation-simple/...
go test ./pkg/tts-simple/...
go test ./cmd/dubbing-mvp/...
```

**Resultado esperado**: Todos os testes devem passar.

### Teste 4: Executar Pipeline Completo
```bash
./dubbing-mvp start --chunk-size 3
```

**Resultado esperado**:
```
🎙️  Dubbing MVP - Starting...
Version: 0.1.0-mvp

📦 Initializing components...
  ✓ Audio Interface (M6)
  ✓ ASR Module (Whisper Tiny)
  ✓ Translation Module (Google Translate)
  ✓ TTS Module (Piper TTS)

🚀 Dubbing started!
📊 Status:
  Input:  Default Microphone
  Output: Virtual Cable Input
  Chunk:  3s

💡 Speak in Portuguese → Others hear in English
⏹️  Press Ctrl+C to stop

--- Processing chunk #1 ---
✓ Captured 48000 audio samples
✓ ASR: '[PT: Texto transcrito apareceria aqui]'
✓ Translation: '[EN: Texto transcrito apareceria aqui]'
✓ TTS: Generated 8000 audio samples
✓ Audio played
📊 Statistics:
  ASR:         1 chunks, avg latency: 5ms
  Translation: 1 sentences, avg latency: 2ms
  TTS:         1 sentences, avg latency: 10ms
```

## 📊 Status Atual do MVP

### ✅ Funcionando
- CLI básico com cobra
- Estrutura de todos os módulos
- Pipeline de processamento
- Estatísticas em tempo real
- Mock de ASR, Translation e TTS

### 🔄 Em Desenvolvimento
- Integração com Whisper.cpp real
- Integração com Google Translate API
- Integração com Piper TTS
- Integração com M6 Audio Interface

### 📋 Próximos Passos
1. Integrar Whisper.cpp para ASR real
2. Integrar Google Translate API
3. Integrar Piper TTS
4. Conectar com M6 para captura/reprodução real
5. Testar com Google Meets

## 🐛 Troubleshooting

### Erro: "command not found: go"
**Solução**: Instalar Go de https://golang.org/dl/

### Erro: "cannot find package"
**Solução**: 
```bash
go mod download
go mod tidy
```

### Erro: "permission denied"
**Solução** (Linux/macOS):
```bash
chmod +x dubbing-mvp
```

### Testes falhando
**Solução**: Verificar que está no diretório raiz do projeto:
```bash
pwd  # Deve mostrar .../audio-dubbing-system
```

## 📈 Progresso do MVP

| Componente | Status | Próximo |
|------------|--------|---------|
| CLI | ✅ 100% | - |
| ASR Mock | ✅ 100% | Integrar Whisper |
| Translation Mock | ✅ 100% | Integrar API |
| TTS Mock | ✅ 100% | Integrar Piper |
| Pipeline | ✅ 100% | Integrar M6 |
| **TOTAL** | **60%** | **Integrações reais** |

## 🎯 Próxima Sessão de Desenvolvimento

### Objetivo
Integrar Whisper.cpp para reconhecimento de fala real

### Tarefas
1. Adicionar whisper.cpp como submódulo
2. Compilar bindings Go
3. Atualizar pkg/asr-simple/asr.go
4. Testar com áudio real

### Tempo Estimado
2 dias

## 📚 Documentação Adicional

- `MVP_README.md` - Documentação completa do MVP
- `.kiro/specs/MVP_PLAN.md` - Plano detalhado
- `.kiro/specs/MVP_NEXT_STEPS.md` - Próximos passos
- `.kiro/specs/MVP_SUMMARY.md` - Resumo executivo

## 🤝 Contribuindo

Para contribuir com o desenvolvimento:

1. Escolha uma tarefa do MVP_NEXT_STEPS.md
2. Crie uma branch: `git checkout -b feature/nome-da-feature`
3. Implemente e teste
4. Commit: `git commit -m "feat: descrição"`
5. Push: `git push origin feature/nome-da-feature`
6. Abra um Pull Request

## ✅ Checklist de Validação

Antes de considerar o MVP completo, verificar:

- [ ] CLI funciona em Windows/Linux/macOS
- [ ] Todos os testes passam
- [ ] ASR reconhece português real
- [ ] Translation traduz corretamente
- [ ] TTS sintetiza voz clara
- [ ] Pipeline completo funciona
- [ ] Latência < 2 segundos
- [ ] Funciona com Google Meets
- [ ] Documentação atualizada
- [ ] Sem crashes em 10 minutos

---

**Status**: 🚀 MVP 60% completo
**Próximo**: Integrar Whisper.cpp
**Data**: 2025-11-29
