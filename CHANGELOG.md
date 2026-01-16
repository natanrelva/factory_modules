# Changelog - Dubbing POC

## [Ciclo 1.2] - 2026-01-16 ⏳ IMPLEMENTADO

### Objetivo
Implementar algoritmo de Linear Resampling para controle de velocidade de reprodução

### Implementado
- ✅ Ring Buffer com leitura não-destrutiva (`get_relative`, `advance`)
- ✅ Algoritmo de interpolação linear para resampling
- ✅ Controle de playback_rate (0.1x a 4.0x)
- ✅ Interface web com slider de velocidade
- ✅ Upload e processamento de arquivos de áudio (MP3/WAV)
- ✅ Monitoramento de buffer health em tempo real
- ✅ Comunicação bidirecional main thread ↔ AudioWorklet

### Funcionalidades
- Carregamento de arquivos de áudio via input file
- Ajuste de velocidade em tempo real (0.1x a 2.0x)
- Efeito "monstro" (0.5x - grave e lento)
- Efeito "esquilo" (1.5x - agudo e rápido)
- Barra de progresso do buffer
- Display de velocidade atual

### Descobertas Técnicas
- **Cursor Fracionário:** Permite interpolação sub-sample precisa
- **Interpolação Linear:** Fórmula `y = s1 + (s2 - s1) * fraction`
- **Upload em Chunks:** 4096 samples por mensagem evita travamento
- **Compilação Incremental:** 0.42s (23x mais rápido que build inicial)

### Pendente Validação
- ⏳ Teste com arquivo de áudio real
- ⏳ Medição de qualidade em diferentes velocidades
- ⏳ Identificação de ponto de ruptura (artefatos)
- ⏳ Latência de mudança de velocidade

### Documentação
- 📄 `docs/ciclo-1.2-feedback.md` - ADR completo
- 📄 `docs/GUIA-VALIDACAO-CICLO-1.2.md` - Guia de teste

### Próximo Ciclo
**Após validação:** Ciclo 2 - Jitter Buffer e Sincronização
- Implementar buffer adaptativo
- Adicionar detecção de underrun/overflow
- Preparar para streaming de IA em tempo real

---

## [Ciclo 1.1] - 2026-01-16 ✅ CONCLUÍDO

### Objetivo
Validar pipeline Rust → WASM → AudioWorklet → Alto-falantes

### Implementado
- ✅ Ring Buffer circular em Rust (`elastic-kernel/src/buffer.rs`)
- ✅ Interface WASM com oscilador de teste 440Hz (`elastic-kernel/src/lib.rs`)
- ✅ AudioWorklet com comunicação via postMessage (`web-client/processor.js`)
- ✅ Thread principal processando WASM (`web-client/main.js`)
- ✅ Interface HTML simples para teste (`web-client/index.html`)

### Validado
- ✅ Tom de 440Hz audível e contínuo
- ✅ Sem erros no console
- ✅ Sem latência perceptível
- ✅ Sem cliques ou artefatos

### Descobertas Arquiteturais
- **Crítico:** AudioWorklets não suportam imports ES6 de WASM
- **Solução:** WASM na thread principal + postMessage para worklet
- **Performance:** requestAnimationFrame (~60fps) suficiente para alimentar worklet

### Documentação
- 📄 `docs/ciclo-1.1-feedback.md` - ADR completo
- 📄 `docs/GUIA-VALIDACAO-CICLO-1.1.md` - Guia de teste
- 📄 `docs/RESUMO-CICLO-1.1.md` - Resumo executivo

### Próximo Ciclo
**Ciclo 1.2:** Algoritmo de Elasticidade (Linear Resampling)
- Implementar controle de velocidade (0.8x a 1.2x)
- Adicionar sliders de controle em tempo real
- Testar limites de qualidade

---

## [Setup Inicial] - 2026-01-16

### Configuração do Ambiente
- Instalado wasm-pack 0.13.1
- Configurado Rust 1.91.1 com target wasm32-unknown-unknown
- Servidor HTTP via npx http-server

### Estrutura do Projeto
```
factory_modules/
├── elastic-kernel/       # Crate Rust
│   ├── src/
│   │   ├── lib.rs
│   │   └── buffer.rs
│   ├── Cargo.toml
│   └── pkg/              # Gerado por wasm-pack
├── web-client/           # Frontend
│   ├── index.html
│   ├── main.js
│   └── processor.js
├── docs/                 # Documentação
└── README.md
```
