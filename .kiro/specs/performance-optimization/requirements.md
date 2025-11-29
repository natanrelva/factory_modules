# Requirements Document - Performance Optimization

## Introduction

Este documento define os requisitos para otimização de performance do MVP de dublagem em tempo real. O objetivo é reduzir a latência total de ~10s para ~5s, mantendo a qualidade e funcionalidade.

## Glossary

- **System**: Sistema de dublagem em tempo real PT→EN
- **Pipeline**: Sequência de processamento (Captura → ASR → Tradução → TTS → Output)
- **Latency**: Tempo entre captura de áudio e reprodução da tradução
- **Chunk**: Segmento de áudio processado (atualmente 3 segundos)
- **Throughput**: Quantidade de áudio processado por segundo
- **Cache**: Armazenamento temporário de traduções para reutilização

## Requirements

### Requirement 1: Redução de Latência Total

**User Story:** Como usuário, quero que o sistema responda mais rápido, para que a dublagem seja mais natural e em tempo real.

#### Acceptance Criteria

1. WHEN o sistema processa um chunk de áudio THEN a latência total SHALL ser menor que 6 segundos
2. WHEN o sistema traduz uma frase já traduzida anteriormente THEN a latência SHALL ser menor que 2 segundos (usando cache)
3. WHEN o sistema processa chunks em paralelo THEN o throughput SHALL aumentar em pelo menos 30%
4. WHEN o sistema detecta silêncio THEN o processamento SHALL ser pulado imediatamente
5. WHEN o sistema usa cache de traduções THEN a taxa de acerto SHALL ser maior que 40% em conversas típicas

### Requirement 2: Processamento Paralelo

**User Story:** Como desenvolvedor, quero que os componentes processem em paralelo, para que a latência seja reduzida.

#### Acceptance Criteria

1. WHEN o ASR está processando THEN a tradução do chunk anterior SHALL estar em andamento simultaneamente
2. WHEN múltiplos chunks estão disponíveis THEN o sistema SHALL processar até 3 chunks em paralelo
3. WHEN o processamento paralelo está ativo THEN a ordem dos resultados SHALL ser mantida
4. WHEN um componente falha THEN os outros componentes SHALL continuar processando
5. WHEN o sistema está sob carga THEN o uso de CPU SHALL ser distribuído entre os componentes

### Requirement 3: Cache de Traduções

**User Story:** Como usuário, quero que frases repetidas sejam traduzidas instantaneamente, para que conversas com repetições sejam mais fluidas.

#### Acceptance Criteria

1. WHEN uma frase é traduzida THEN o resultado SHALL ser armazenado em cache
2. WHEN uma frase em cache é solicitada THEN a tradução SHALL ser retornada em menos de 100ms
3. WHEN o cache atinge 1000 entradas THEN as entradas mais antigas SHALL ser removidas (LRU)
4. WHEN o sistema reinicia THEN o cache SHALL ser persistido em disco
5. WHEN uma tradução em cache é usada THEN o sistema SHALL registrar a economia de tempo

### Requirement 4: Detecção Inteligente de Silêncio

**User Story:** Como usuário, quero que o sistema não processe silêncio, para que a latência seja reduzida e recursos sejam economizados.

#### Acceptance Criteria

1. WHEN o áudio capturado tem energia menor que 0.01 THEN o sistema SHALL classificar como silêncio
2. WHEN silêncio é detectado THEN o processamento de ASR SHALL ser pulado
3. WHEN silêncio é detectado THEN o sistema SHALL registrar a economia de tempo
4. WHEN fala é detectada após silêncio THEN o processamento SHALL retomar imediatamente
5. WHEN a detecção de silêncio está ativa THEN a taxa de falsos positivos SHALL ser menor que 5%

### Requirement 5: Otimização de Chunks

**User Story:** Como desenvolvedor, quero chunks menores e mais frequentes, para que a latência percebida seja menor.

#### Acceptance Criteria

1. WHEN o tamanho do chunk é configurado THEN o sistema SHALL aceitar valores entre 1 e 5 segundos
2. WHEN chunks menores são usados THEN a latência percebida SHALL ser menor
3. WHEN chunks menores são usados THEN o throughput total SHALL ser mantido
4. WHEN o sistema processa chunks de 2s THEN a latência total SHALL ser menor que 5 segundos
5. WHEN chunks são processados THEN o sistema SHALL medir e reportar a latência de cada componente

### Requirement 6: Métricas e Monitoramento

**User Story:** Como desenvolvedor, quero métricas detalhadas de performance, para que eu possa identificar gargalos e otimizar.

#### Acceptance Criteria

1. WHEN o sistema processa um chunk THEN as métricas de latência de cada componente SHALL ser registradas
2. WHEN o sistema usa cache THEN a taxa de acerto SHALL ser reportada
3. WHEN o sistema detecta silêncio THEN a economia de tempo SHALL ser reportada
4. WHEN o sistema processa em paralelo THEN o ganho de throughput SHALL ser reportado
5. WHEN o usuário solicita estatísticas THEN o sistema SHALL exibir métricas agregadas dos últimos 100 chunks

### Requirement 7: Configuração de Performance

**User Story:** Como usuário, quero configurar o nível de performance, para que eu possa balancear latência vs qualidade.

#### Acceptance Criteria

1. WHEN o modo "low-latency" é ativado THEN o sistema SHALL usar chunks de 2s e processamento paralelo
2. WHEN o modo "balanced" é ativado THEN o sistema SHALL usar chunks de 3s e cache
3. WHEN o modo "quality" é ativado THEN o sistema SHALL usar chunks de 4s e sem paralelismo
4. WHEN um modo é selecionado THEN o sistema SHALL aplicar as configurações automaticamente
5. WHEN o modo é alterado THEN o sistema SHALL reportar a latência esperada

### Requirement 8: Testes de Performance

**User Story:** Como desenvolvedor, quero testes automatizados de performance, para que eu possa validar otimizações.

#### Acceptance Criteria

1. WHEN os testes de performance são executados THEN a latência média SHALL ser medida
2. WHEN os testes de cache são executados THEN a taxa de acerto SHALL ser validada
3. WHEN os testes de paralelismo são executados THEN o ganho de throughput SHALL ser medido
4. WHEN os testes de silêncio são executados THEN a taxa de detecção SHALL ser validada
5. WHEN todos os testes passam THEN o sistema SHALL garantir latência menor que 6 segundos

## Non-Functional Requirements

### Performance
- Latência total: < 6s (objetivo: 5s)
- Throughput: > 0.5 chunks/segundo
- Uso de CPU: < 50% em média
- Uso de memória: < 500 MB

### Reliability
- Taxa de erro: < 1%
- Disponibilidade: > 99%
- Recuperação de falhas: < 1s

### Maintainability
- Código testável com TDD
- Cobertura de testes: > 80%
- Documentação atualizada

## Success Metrics

1. **Latência Total**: Redução de 10s para 5s (50% de melhoria)
2. **Cache Hit Rate**: > 40% em conversas típicas
3. **Throughput**: Aumento de 30% com paralelismo
4. **Detecção de Silêncio**: > 95% de precisão
5. **Satisfação do Usuário**: Latência percebida como "aceitável"

## Constraints

1. Manter 100% de funcionalidade existente
2. Manter compatibilidade com código atual
3. Não aumentar uso de memória em mais de 20%
4. Manter código limpo e testável
5. Documentar todas as otimizações

## Assumptions

1. Conversas típicas têm 30-40% de repetição
2. Silêncio representa 20-30% do tempo total
3. Processamento paralelo é possível sem race conditions
4. Cache de 1000 entradas é suficiente
5. Chunks de 2s mantêm qualidade aceitável

## Dependencies

1. Go 1.21+ (goroutines para paralelismo)
2. Python 3.8+ (scripts existentes)
3. Bibliotecas existentes (Vosk, Argos, PyAudio)
4. Sistema de arquivos (para cache persistente)

## Risks

1. **Paralelismo**: Race conditions e bugs de concorrência
2. **Cache**: Uso excessivo de memória
3. **Chunks menores**: Possível degradação de qualidade do ASR
4. **Complexidade**: Código mais difícil de manter

## Mitigation

1. Usar testes extensivos de concorrência
2. Implementar LRU cache com limite de tamanho
3. Testar qualidade com diferentes tamanhos de chunk
4. Manter código modular e bem documentado
