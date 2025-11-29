# Plano Completo de Integração do Sistema de Dublagem PT→EN

## Visão Geral

Este documento consolida o planejamento completo para implementação de todos os módulos restantes e o módulo principal de integração com foco em UX de sistema operacional.

## Status Atual

✅ **M6: Audio Interface Module** - COMPLETO (V2.0)
✅ **M2: ASR Module** - SPEC COMPLETA (pronto para implementação)
🔄 **M3: Translation Module** - SPEC EM ANDAMENTO
📋 **M4: TTS Module** - A CRIAR SPEC
📋 **M0: Main Integration Module** - A CRIAR SPEC

---

## Módulos Restantes

### M3: Translation Module (Tradução PT→EN)

**Status**: Requirements criado, Design em andamento

**Resumo Executivo**:
- Traduz texto PT para EN com BLEU > 30
- Mantém contexto de 3 sentenças anteriores
- Latência < 150ms por sentença
- Preserva prosódia para TTS natural

**Próximos Passos**:
1. Completar design.md com propriedades de correção
2. Criar tasks.md com plano de implementação
3. Implementar seguindo padrão do M6

---

### M4: TTS Module (Text-to-Speech)

**Responsabilidade**: Sintetizar texto EN em áudio EN com voz clonada

**Requisitos Principais**:
1. Síntese de alta qualidade (MOS > 4.0)
2. Clonagem de voz do usuário (similaridade > 70%)
3. Transferência de prosódia PT→EN
4. Latência < 200ms por sentença
5. Streaming incremental de áudio

**Sub-módulos Planejados**:
```
M4: TTS Module
├── M4.1: Text Processing
│   ├── TextNormalizer (expandir abreviações, números)
│   ├── Phonemizer (texto → fonemas)
│   └── ProsodyParser (interpretar marcadores)
│
├── M4.2: Voice Cloning
│   ├── SpeakerEncoder (extrair embedding de voz)
│   ├── VoiceProfileManager (gerenciar perfis)
│   └── AdaptationEngine (adaptar modelo)
│
├── M4.3: Synthesis Engine
│   ├── AcousticModel (gerar mel-spectrogram)
│   ├── DurationPredictor (prever duração fonemas)
│   └── PitchPredictor (prever contorno pitch)
│
├── M4.4: Vocoder
│   ├── VocoderModel (mel → waveform)
│   ├── StreamingVocoder (geração incremental)
│   └── PostFilter (melhorar qualidade)
│
├── M4.5: Prosody Control
│   ├── ProsodyTransfer (transferir PT→EN)
│   ├── EmotionController (ajustar emoção)
│   └── RhythmAdjuster (velocidade/pausas)
│
└── M4.6: Orchestration
    └── TTSCoordinator (coordenar pipeline)
```

**Tecnologias Recomendadas**:
- **Coqui TTS** ou **Piper TTS** para síntese
- **XTTS** para clonagem de voz
- **HiFi-GAN** ou **UnivNet** para vocoder
- **FastSpeech 2** para controle de prosódia

---

## M0: Main Integration Module (Módulo Principal)

### Visão de UX para Sistema Operacional

O módulo principal deve funcionar como uma **aplicação de sistema** com as seguintes características:

#### 1. Modos de Operação

**Modo 1: System Tray Application (Recomendado)**
```
┌─────────────────────────────────────┐
│  🎙️ DubbingPT→EN                    │
├─────────────────────────────────────┤
│  ● Ativo (Dublando)                 │
│  ⏸️  Pausado                         │
│  ⏹️  Parado                          │
├─────────────────────────────────────┤
│  📊 Status:                          │
│    Latência: 465ms                  │
│    Qualidade: 92%                   │
│    CPU: 28%                         │
├─────────────────────────────────────┤
│  ⚙️  Configurações                   │
│  📈 Métricas Detalhadas             │
│  🔊 Dispositivos de Áudio           │
│  ❌ Sair                             │
└─────────────────────────────────────┘
```

**Modo 2: Overlay Transparente**
```
┌──────────────────┐
│ 🎙️ PT→EN         │
│ ● 465ms | 92%   │
└──────────────────┘
```

**Modo 3: CLI para Automação**
```bash
dubbing-pten start --profile gaming
dubbing-pten status
dubbing-pten stop
```



#### 2. Arquitetura do Módulo Principal

```
M0: Main Integration Module
├── M0.1: Application Layer (UI/UX)
│   ├── SystemTrayUI (ícone na bandeja)
│   ├── OverlayRenderer (overlay transparente)
│   ├── SettingsPanel (painel de configurações)
│   └── MetricsDashboard (dashboard de métricas)
│
├── M0.2: Pipeline Orchestrator
│   ├── ModuleLifecycleManager (gerenciar M2, M3, M4, M6)
│   ├── DataFlowCoordinator (coordenar fluxo de dados)
│   ├── BackpressureManager (gerenciar backpressure global)
│   └── ErrorRecoveryManager (recuperação de erros)
│
├── M0.3: Configuration Management
│   ├── ProfileManager (perfis: gaming, meeting, movie)
│   ├── DeviceManager (gerenciar dispositivos áudio)
│   ├── ModelManager (gerenciar modelos ML)
│   └── PreferencesStore (salvar preferências)
│
├── M0.4: Monitoring & Telemetry
│   ├── PerformanceMonitor (CPU, memória, latência)
│   ├── QualityMonitor (WER, BLEU, MOS)
│   ├── HealthChecker (verificar saúde dos módulos)
│   └── TelemetryCollector (coletar telemetria)
│
├── M0.5: Audio Routing
│   ├── VirtualAudioDevice (dispositivo virtual)
│   ├── AudioRouter (rotear entrada/saída)
│   └── MixerController (controlar volumes)
│
└── M0.6: System Integration
    ├── HotkeyManager (atalhos globais)
    ├── NotificationManager (notificações do SO)
    ├── AutostartManager (iniciar com SO)
    └── UpdateManager (atualizações automáticas)
```

#### 3. Fluxo de Dados End-to-End

```
┌─────────────┐
│ Microfone   │ (Áudio PT)
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│ M6: Audio Interface                 │
│  - Captura áudio                    │
│  - Aplica VAD                       │
│  - Gera PCM frames                  │
└──────┬──────────────────────────────┘
       │ PCM Frames (PT)
       ▼
┌─────────────────────────────────────┐
│ M2: ASR Module                      │
│  - Extrai features (MFCC)           │
│  - Reconhece fala (Whisper/Vosk)    │
│  - Emite tokens PT                  │
└──────┬──────────────────────────────┘
       │ ASR Tokens (PT)
       ▼
┌─────────────────────────────────────┐
│ M3: Translation Module              │
│  - Traduz PT→EN (NLLB/DeepL)        │
│  - Mantém contexto                  │
│  - Adiciona prosódia                │
└──────┬──────────────────────────────┘
       │ Translated Tokens (EN)
       ▼
┌─────────────────────────────────────┐
│ M4: TTS Module                      │
│  - Sintetiza voz EN                 │
│  - Clona voz do usuário             │
│  - Transfere prosódia               │
└──────┬──────────────────────────────┘
       │ PCM Frames (EN)
       ▼
┌─────────────────────────────────────┐
│ M6: Audio Interface                 │
│  - Reproduz áudio EN                │
│  - Sincroniza com entrada           │
└──────┬──────────────────────────────┘
       │
       ▼
┌─────────────┐
│ Alto-falante│ (Áudio EN)
└─────────────┘

Latência Total: ~465ms (dentro do budget de 700ms)
```

#### 4. Perfis de Uso

**Perfil Gaming**:
```yaml
gaming:
  priority: low_latency
  quality: balanced
  asr_model: vosk-small  # Mais rápido
  translation_model: nllb-distilled
  tts_model: piper-fast
  target_latency: 400ms
  voice_cloning: false  # Desabilitado para velocidade
```

**Perfil Meeting**:
```yaml
meeting:
  priority: high_quality
  quality: high
  asr_model: whisper-medium
  translation_model: nllb-large
  tts_model: coqui-xtts
  target_latency: 600ms
  voice_cloning: true
  noise_cancellation: true
```

**Perfil Movie**:
```yaml
movie:
  priority: maximum_quality
  quality: maximum
  asr_model: whisper-large
  translation_model: deepl-api
  tts_model: elevenlabs-api
  target_latency: 800ms
  voice_cloning: true
  prosody_transfer: enhanced
```

#### 5. Interface de Configuração

**Tela Principal**:
```
╔════════════════════════════════════════════════════════════╗
║  DubbingPT→EN - Dublagem Automática em Tempo Real         ║
╠════════════════════════════════════════════════════════════╣
║                                                            ║
║  Status: ● Ativo                                          ║
║                                                            ║
║  ┌──────────────────────────────────────────────────┐    ║
║  │  Latência End-to-End: 465ms                      │    ║
║  │  ████████████████░░░░░░░░░░ 65%                  │    ║
║  │                                                    │    ║
║  │  Qualidade Geral: 92%                            │    ║
║  │  ██████████████████████░░░░ 92%                  │    ║
║  └──────────────────────────────────────────────────┘    ║
║                                                            ║
║  Módulos:                                                 ║
║  ✓ Audio Interface    [25ms]  [OK]                       ║
║  ✓ ASR (Whisper)      [180ms] [OK]                       ║
║  ✓ Translation (NLLB) [120ms] [OK]                       ║
║  ✓ TTS (Coqui)        [140ms] [OK]                       ║
║                                                            ║
║  Dispositivos:                                            ║
║  🎤 Entrada:  Microfone (Realtek HD Audio)               ║
║  🔊 Saída:    Alto-falantes (Realtek HD Audio)           ║
║                                                            ║
║  Perfil Ativo: [Gaming ▼]                                ║
║                                                            ║
║  [⏸️ Pausar]  [⏹️ Parar]  [⚙️ Configurações]  [📊 Métricas]║
║                                                            ║
╚════════════════════════════════════════════════════════════╝
```

**Painel de Configurações Avançadas**:
```
╔════════════════════════════════════════════════════════════╗
║  Configurações Avançadas                                   ║
╠════════════════════════════════════════════════════════════╣
║                                                            ║
║  ┌─ Modelos ─────────────────────────────────────────┐   ║
║  │                                                     │   ║
║  │  ASR Model:      [Whisper Medium ▼]               │   ║
║  │  Translation:    [NLLB 600M ▼]                    │   ║
║  │  TTS Model:      [Coqui XTTS ▼]                   │   ║
║  │                                                     │   ║
║  │  [📥 Baixar Modelos]  [🗑️ Gerenciar Cache]        │   ║
║  └─────────────────────────────────────────────────────┘   ║
║                                                            ║
║  ┌─ Performance ──────────────────────────────────────┐   ║
║  │                                                     │   ║
║  │  Target Latency:  [500ms]  ◄─────────────► [800ms]│   ║
║  │  Quality Level:   [●●●●○] Balanced                │   ║
║  │  CPU Limit:       [50%]    ◄─────────────► [100%] │   ║
║  │                                                     │   ║
║  │  ☑ Enable GPU Acceleration                        │   ║
║  │  ☑ Enable Voice Cloning                           │   ║
║  │  ☑ Enable Prosody Transfer                        │   ║
║  └─────────────────────────────────────────────────────┘   ║
║                                                            ║
║  ┌─ Audio ────────────────────────────────────────────┐   ║
║  │                                                     │   ║
║  │  Input Device:    [Microfone ▼]                   │   ║
║  │  Output Device:   [Alto-falantes ▼]               │   ║
║  │  Sample Rate:     [16000 Hz ▼]                    │   ║
║  │                                                     │   ║
║  │  ☑ Noise Cancellation                             │   ║
║  │  ☑ Echo Cancellation                              │   ║
║  │  ☑ Auto Gain Control                              │   ║
║  └─────────────────────────────────────────────────────┘   ║
║                                                            ║
║  [💾 Salvar]  [↺ Restaurar Padrões]  [✖ Cancelar]        ║
║                                                            ║
╚════════════════════════════════════════════════════════════╝
```

#### 6. Atalhos de Teclado Globais

```
Ctrl + Alt + D     - Ativar/Desativar dublagem
Ctrl + Alt + P     - Pausar/Retomar
Ctrl + Alt + M     - Mutar/Desmutar saída
Ctrl + Alt + V     - Ajustar volume saída
Ctrl + Alt + S     - Abrir configurações
Ctrl + Alt + Q     - Mostrar/Ocultar overlay
```

#### 7. Notificações do Sistema

```
┌─────────────────────────────────────┐
│ 🎙️ DubbingPT→EN                    │
├─────────────────────────────────────┤
│ Dublagem iniciada com sucesso      │
│ Latência: 465ms | Qualidade: 92%   │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ ⚠️ DubbingPT→EN                     │
├─────────────────────────────────────┤
│ Latência alta detectada (850ms)    │
│ Considere usar perfil "Gaming"     │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ ❌ DubbingPT→EN                     │
├─────────────────────────────────────┤
│ Erro no módulo ASR                 │
│ Tentando recuperar...              │
└─────────────────────────────────────┘
```

#### 8. Dashboard de Métricas em Tempo Real

```
╔════════════════════════════════════════════════════════════╗
║  Métricas em Tempo Real                                    ║
╠════════════════════════════════════════════════════════════╣
║                                                            ║
║  Latência por Módulo (ms):                                ║
║  ┌────────────────────────────────────────────────────┐   ║
║  │ 200│                                                │   ║
║  │    │     ╱╲                                         │   ║
║  │ 150│    ╱  ╲      ╱╲                               │   ║
║  │    │   ╱    ╲    ╱  ╲    ╱╲                        │   ║
║  │ 100│  ╱      ╲  ╱    ╲  ╱  ╲                       │   ║
║  │    │ ╱        ╲╱      ╲╱    ╲                      │   ║
║  │  50│╱                        ╲                     │   ║
║  │    └────────────────────────────────────────────   │   ║
║  │     M6   M2   M3   M4   M6                         │   ║
║  └────────────────────────────────────────────────────┘   ║
║                                                            ║
║  Qualidade:                                               ║
║  ┌────────────────────────────────────────────────────┐   ║
║  │  WER (ASR):        12.3%  ████████████████░░░░     │   ║
║  │  BLEU (Trans):     34.2   ██████████████████░░     │   ║
║  │  MOS (TTS):        4.1    ████████████████████░    │   ║
║  │  Semantic Sim:     0.87   ███████████████████░░    │   ║
║  └────────────────────────────────────────────────────┘   ║
║                                                            ║
║  Recursos do Sistema:                                     ║
║  ┌────────────────────────────────────────────────────┐   ║
║  │  CPU:     28%  ██████░░░░░░░░░░░░░░░░░░░░░░░░░░   │   ║
║  │  Memory:  1.2GB / 4.0GB                            │   ║
║  │  GPU:     45%  ███████████░░░░░░░░░░░░░░░░░░░░░   │   ║
║  └────────────────────────────────────────────────────┘   ║
║                                                            ║
║  Estatísticas da Sessão:                                  ║
║  ┌────────────────────────────────────────────────────┐   ║
║  │  Tempo Ativo:      00:15:32                        │   ║
║  │  Sentenças:        127                             │   ║
║  │  Palavras:         1,543                           │   ║
║  │  Erros:            3                               │   ║
║  └────────────────────────────────────────────────────┘   ║
║                                                            ║
║  [📊 Exportar Relatório]  [🔄 Resetar Estatísticas]       ║
║                                                            ║
╚════════════════════════════════════════════════════════════╝
```

---

## Ordem de Implementação Recomendada

### Fase 1: Completar Specs (1 semana)
1. ✅ M2: ASR Module - COMPLETO
2. 🔄 M3: Translation Module - Finalizar design e tasks
3. 📋 M4: TTS Module - Criar requirements, design, tasks
4. 📋 M0: Main Integration - Criar requirements, design, tasks

### Fase 2: Implementação Core (4-6 semanas)
1. **Semana 1-2**: M2 (ASR Module)
2. **Semana 3**: M4 (TTS Module - básico)
3. **Semana 4**: M3 (Translation Module)
4. **Semana 5-6**: M0 (Main Integration - MVP)

### Fase 3: Features Avançadas (2-3 semanas)
1. Voice Cloning (M4)
2. Prosody Transfer (M3 + M4)
3. UI/UX Polimento (M0)
4. Perfis de Uso (M0)

### Fase 4: Otimização e Testes (1-2 semanas)
1. Performance tuning
2. Testes end-to-end
3. Testes de usuário
4. Documentação final

---

## Tecnologias Recomendadas para M0

### Backend (Go)
- **Fyne** ou **Wails** para UI desktop cross-platform
- **systray** para ícone na bandeja do sistema
- **robotgo** para hotkeys globais
- **gopsutil** para monitoramento de recursos

### Frontend (se usar Wails)
- **React** + **TypeScript** para UI
- **TailwindCSS** para estilização
- **Recharts** para gráficos de métricas
- **Electron** alternativa (mais pesado)

### Comunicação entre Módulos
- **Channels** do Go para comunicação assíncrona
- **gRPC** para comunicação entre processos (se necessário)
- **Protocol Buffers** para serialização eficiente

---

## Requisitos de Sistema

### Mínimo
- CPU: 4 cores (Intel i5 ou equivalente)
- RAM: 4GB
- GPU: Não requerida (CPU-only)
- Disco: 5GB (modelos)
- SO: Windows 10+, Linux, macOS

### Recomendado
- CPU: 8 cores (Intel i7 ou equivalente)
- RAM: 8GB
- GPU: NVIDIA GTX 1060 ou equivalente (CUDA)
- Disco: 10GB SSD
- SO: Windows 11, Ubuntu 22.04+, macOS 12+

---

## Próximos Passos Imediatos

1. **Finalizar M3 Translation Module**:
   - Completar design.md com propriedades
   - Criar tasks.md

2. **Criar M4 TTS Module Spec**:
   - requirements.md
   - design.md
   - tasks.md

3. **Criar M0 Main Integration Spec**:
   - requirements.md (foco em UX)
   - design.md (arquitetura de integração)
   - tasks.md (implementação UI + orquestração)

4. **Começar Implementação**:
   - Seguir ordem: M2 → M4 → M3 → M0
   - Testes incrementais em cada fase
   - Integração contínua

---

## Conclusão

Este plano fornece uma visão completa de como construir o sistema de dublagem automática PT→EN com foco em:

1. **Modularidade**: Cada módulo independente e testável
2. **Qualidade**: Specs completas com propriedades de correção
3. **UX**: Interface intuitiva para usuários finais
4. **Performance**: Latência otimizada e uso eficiente de recursos
5. **Manutenibilidade**: Código limpo seguindo SOLID

O sistema será uma aplicação de sistema operacional profissional, com interface gráfica moderna, perfis de uso, e monitoramento em tempo real.
