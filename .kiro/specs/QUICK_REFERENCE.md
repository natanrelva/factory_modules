# Referência Rápida - Sistema de Dublagem PT→EN

## 🚀 Início Rápido

### Para Entender o Projeto
1. Leia: `EXECUTIVE_SUMMARY.md` (5 min)
2. Veja: `VISUAL_ARCHITECTURE.md` (3 min)
3. Explore: `README.md` (2 min)

### Para Implementar
1. Escolha um módulo (M2, M3, M4, ou M0)
2. Leia: `{module}/requirements.md`
3. Leia: `{module}/design.md`
4. Siga: `{module}/tasks.md`

---

## 📊 Status Rápido

| Módulo | Status | Próximo Passo |
|--------|--------|---------------|
| M6 Audio | ✅ 100% | Nenhum |
| M2 ASR | 📋 Spec pronta | Implementar |
| M3 Translation | 📋 Spec pronta | Implementar |
| M4 TTS | ❌ 0% | Criar spec |
| M0 Main | ❌ 0% | Criar spec |

---

## ⚡ Comandos Rápidos (Futuro)

```bash
# Iniciar sistema
dubbing-pten start

# Usar perfil específico
dubbing-pten start --profile gaming

# Ver status
dubbing-pten status

# Pausar
dubbing-pten pause

# Parar
dubbing-pten stop

# Ver métricas
dubbing-pten metrics

# Configurar
dubbing-pten config
```

---

## 🎯 Métricas Principais

### Latência
- **Target**: < 700ms
- **Atual**: ~465ms ✅
- **Por módulo**: M6(25ms) + M2(180ms) + M3(120ms) + M4(140ms) + M6(30ms)

### Qualidade
- **WER (ASR)**: < 15% 🔴
- **BLEU (Translation)**: > 30 🟡
- **MOS (TTS)**: > 4.0 🟡
- **Semantic Similarity**: > 0.85 🟡

---

## 🛠️ Stack Tecnológico

### Backend
- Go 1.21+
- WASAPI (Windows)
- ONNX Runtime

### Modelos
- ASR: Whisper/Vosk
- Translation: NLLB/DeepL
- TTS: Coqui/Piper
- Vocoder: HiFi-GAN

### Frontend
- Fyne/Wails
- systray
- robotgo

---

## 📁 Estrutura de Arquivos

```
.kiro/specs/
├── README.md                    # Índice
├── EXECUTIVE_SUMMARY.md         # Resumo executivo
├── SYSTEM_INTEGRATION_PLAN.md   # Plano de integração
├── VISUAL_ARCHITECTURE.md       # Diagramas
├── WORK_COMPLETED.md            # Trabalho realizado
├── QUICK_REFERENCE.md           # Este arquivo
│
├── asr-module/                  # M2 ✅
│   ├── requirements.md
│   ├── design.md
│   └── tasks.md
│
├── translation-module/          # M3 ✅
│   ├── requirements.md
│   ├── design.md
│   └── tasks.md
│
├── tts-module/                  # M4 📋
└── main-integration/            # M0 📋
```

---

## 🎨 Interface do Usuário

### System Tray
- Ícone na bandeja do sistema
- Menu com status e controles
- Notificações do sistema

### Overlay
- Indicador transparente
- Latência e qualidade
- Sempre visível

### CLI
- Automação e scripting
- Integração com outros sistemas
- Logs detalhados

---

## ⌨️ Atalhos Globais

| Atalho | Ação |
|--------|------|
| `Ctrl+Alt+D` | Ativar/Desativar |
| `Ctrl+Alt+P` | Pausar/Retomar |
| `Ctrl+Alt+M` | Mutar/Desmutar |
| `Ctrl+Alt+S` | Configurações |
| `Ctrl+Alt+Q` | Overlay |

---

## 🎮 Perfis de Uso

### Gaming
- Latência: 400ms
- Qualidade: Balanceada
- Voice Cloning: Não

### Meeting
- Latência: 600ms
- Qualidade: Alta
- Voice Cloning: Sim

### Movie
- Latência: 800ms
- Qualidade: Máxima
- Voice Cloning: Sim

---

## 🧪 Testing

### Property-Based Tests
- M2: 25 properties
- M3: 35 properties
- M4: ~30 properties
- M0: ~20 properties

### Coverage
- Target: > 80%
- Iterations: 100+ per property

---

## 📅 Timeline

### Fase 1: Specs (1 semana)
- [x] M6, M2, M3 ✅
- [ ] M4, M0 (2 dias)

### Fase 2: Core (4-6 semanas)
- [ ] M2 (2 semanas)
- [ ] M4 (1 semana)
- [ ] M3 (1 semana)
- [ ] M0 (2 semanas)

### Fase 3: Features (2-3 semanas)
- [ ] Voice cloning
- [ ] Prosody transfer
- [ ] UI polish

### Fase 4: Otimização (1-2 semanas)
- [ ] Performance
- [ ] Testing
- [ ] Docs

**Total**: 8-12 semanas

---

## 🔗 Links Úteis

### Documentação
- [README.md](README.md) - Índice completo
- [EXECUTIVE_SUMMARY.md](EXECUTIVE_SUMMARY.md) - Resumo executivo
- [VISUAL_ARCHITECTURE.md](VISUAL_ARCHITECTURE.md) - Diagramas

### Specs Completas
- [M2 ASR](asr-module/) - Reconhecimento de fala
- [M3 Translation](translation-module/) - Tradução PT→EN

### Código
- [M6 Audio Interface](../audio-interface/) - Implementado

---

## 💡 Dicas

### Para Desenvolvedores
1. Sempre leia requirements antes de design
2. Sempre leia design antes de tasks
3. Siga as tasks em ordem
4. Escreva testes primeiro (TDD)
5. Use property-based testing

### Para Testers
1. Verifique todas as properties
2. Teste edge cases
3. Teste integração entre módulos
4. Meça latência e qualidade
5. Reporte bugs com contexto

### Para Usuários
1. Escolha o perfil adequado
2. Configure dispositivos de áudio
3. Monitore métricas
4. Reporte problemas
5. Sugira melhorias

---

## ❓ FAQ

**Q: Qual a latência esperada?**
A: ~465ms end-to-end (target: <700ms)

**Q: Funciona offline?**
A: Sim, com modelos locais

**Q: Suporta GPU?**
A: Sim, opcional para melhor performance

**Q: Qual a qualidade da tradução?**
A: BLEU > 30, Semantic Similarity > 0.85

**Q: Clona minha voz?**
A: Sim, com similaridade > 70%

**Q: Funciona em tempo real?**
A: Sim, RTF < 0.5 (2x mais rápido que tempo real)

---

## 📞 Próximos Passos

1. ✅ Finalizar specs M3 - COMPLETO
2. 📋 Criar spec M4 TTS
3. 📋 Criar spec M0 Main
4. 📋 Implementar M2 ASR
5. 📋 Integrar tudo

---

**Última Atualização**: 2025-11-29
**Versão**: 1.0.0
