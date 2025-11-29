# 🎙️ Guia Completo: Usando com Google Meets

## Visão Geral

Este guia explica como configurar e usar o sistema de dublagem em tempo real com Google Meets, permitindo que você fale em português e os outros participantes ouçam em inglês.

## 📋 Pré-requisitos

### 1. Software Necessário

✅ **Python 3.8+**
```powershell
python --version
# Deve mostrar: Python 3.8 ou superior
```

✅ **Go 1.21+**
```powershell
go version
# Deve mostrar: go version go1.21 ou superior
```

✅ **Dependências Python**
```powershell
pip install argostranslate pyttsx3 pywin32 vosk pyaudio
```

✅ **Modelo Vosk PT**
- Baixar de: https://alphacephei.com/vosk/models
- Modelo: `vosk-model-small-pt-0.3.zip` (69 MB)
- Extrair para: `models/vosk-model-small-pt-0.3/`

✅ **Pacote de Tradução Argos**
```powershell
# Instalar pacote PT→EN
python -c "import argostranslate.package; argostranslate.package.update_package_index(); available_packages = argostranslate.package.get_available_packages(); package_to_install = next(filter(lambda x: x.from_code == 'pt' and x.to_code == 'en', available_packages)); argostranslate.package.install_from_path(package_to_install.download())"
```

### 2. Hardware Necessário

✅ **Microfone** - Qualquer microfone USB ou integrado
✅ **Cabo de Áudio Virtual** - Para rotear o áudio para o Google Meets

**Opções de Cabo Virtual**:
1. **VB-Audio Virtual Cable** (Recomendado)
   - Download: https://vb-audio.com/Cable/
   - Gratuito
   - Fácil de usar

2. **VoiceMeeter** (Alternativa)
   - Download: https://vb-audio.com/Voicemeeter/
   - Mais recursos
   - Mais complexo

## 🔧 Configuração Passo a Passo

### Passo 1: Instalar Cabo de Áudio Virtual

1. **Baixar VB-Audio Virtual Cable**
   - Acesse: https://vb-audio.com/Cable/
   - Clique em "Download"
   - Extraia o arquivo ZIP

2. **Instalar**
   - Execute `VBCABLE_Setup_x64.exe` (como Administrador)
   - Clique em "Install Driver"
   - Reinicie o computador se solicitado

3. **Verificar Instalação**
   - Abra "Configurações de Som" do Windows
   - Você deve ver:
     - **Entrada**: "CABLE Output (VB-Audio Virtual Cable)"
     - **Saída**: "CABLE Input (VB-Audio Virtual Cable)"

### Passo 2: Compilar o Sistema

```powershell
# 1. Navegar até o diretório do projeto
cd C:\factory_modules

# 2. Adicionar Python ao PATH (ajuste o caminho se necessário)
$env:PATH = "C:\Users\natan\AppData\Local\Programs\Python\Python313;C:\Users\natan\AppData\Local\Programs\Python\Python313\Scripts;$env:PATH"

# 3. Compilar
go build -o dubbing-mvp.exe cmd/dubbing-mvp/main.go

# 4. Verificar compilação
.\dubbing-mvp.exe --version
```

### Passo 3: Testar o Sistema (Sem Google Meets)

```powershell
# Teste básico com modo balanced
.\dubbing-mvp.exe start --use-vosk --use-argos --use-windows-tts --use-real-audio

# Ou com modo low-latency para menor latência
.\dubbing-mvp.exe start --mode low-latency --use-vosk --use-argos --use-windows-tts --use-real-audio
```

**O que deve acontecer**:
1. Sistema inicia e mostra configurações
2. Fale em português no microfone
3. Você deve ouvir a tradução em inglês nos seus alto-falantes
4. Verifique os logs para confirmar que está funcionando

**Exemplo de log esperado**:
```
--- Processing chunk #1 ---
✓ Captured 48000 audio samples
🎙️  Speech detected - processing...
✓ ASR: 'bom dia' (latency: 2.1s)
✓ Translation: 'Good morning' (latency: 4.3s)
✓ TTS: Generated 32000 audio samples (latency: 0.6s)
✓ Audio played
```

### Passo 4: Configurar Áudio no Windows

#### 4.1 Configurar Dispositivos de Gravação

1. **Abrir Configurações de Som**
   - Clique direito no ícone de som na bandeja
   - Selecione "Configurações de som"
   - Clique em "Painel de controle de som"

2. **Aba "Gravação"**
   - Seu microfone físico deve estar como "Dispositivo Padrão"
   - "CABLE Output" deve estar habilitado (mas não como padrão)

#### 4.2 Configurar Dispositivos de Reprodução

1. **Aba "Reprodução"**
   - Seus alto-falantes/fones devem estar como "Dispositivo Padrão"
   - "CABLE Input" deve estar habilitado

### Passo 5: Configurar Google Meets

1. **Abrir Google Meets**
   - Acesse: https://meet.google.com
   - Crie ou entre em uma reunião

2. **Configurar Áudio no Meets**
   - Clique nos 3 pontos (⋮) no canto inferior direito
   - Selecione "Configurações"
   - Vá para a aba "Áudio"

3. **Configurar Microfone**
   - **Microfone**: Selecione "CABLE Output (VB-Audio Virtual Cable)"
   - Isso fará o Meets capturar o áudio traduzido

4. **Configurar Alto-falante**
   - **Alto-falante**: Selecione seus alto-falantes/fones normais
   - Isso permite você ouvir os outros participantes

5. **Testar Áudio**
   - Clique em "Testar microfone"
   - Fale em português
   - Você deve ver a barra de volume se movendo

## 🚀 Uso em Produção

### Fluxo de Trabalho Completo

```
┌─────────────────────────────────────────────────────────────┐
│                    SEU COMPUTADOR                            │
│                                                              │
│  Você fala PT → [Microfone] → [Sistema Dublagem]           │
│                                      ↓                       │
│                              Traduz para EN                  │
│                                      ↓                       │
│                          [Cabo Virtual] → [Google Meets]    │
│                                                ↓             │
│                                    Transmite EN para outros  │
└─────────────────────────────────────────────────────────────┘
```

### Comando Recomendado para Google Meets

```powershell
# Modo Low-Latency (melhor para conversas em tempo real)
.\dubbing-mvp.exe start `
  --mode low-latency `
  --use-vosk `
  --use-argos `
  --use-windows-tts `
  --use-real-audio `
  --use-silence-detection `
  --use-metrics
```

**Por que este comando?**
- `--mode low-latency`: Chunks de 1s para resposta rápida
- `--use-vosk`: ASR offline e gratuito
- `--use-argos`: Tradução offline e gratuita
- `--use-windows-tts`: TTS nativo do Windows
- `--use-real-audio`: Captura real do microfone
- `--use-silence-detection`: Pula processamento durante silêncio
- `--use-metrics`: Monitora performance em tempo real

### Sequência de Início

1. **Iniciar o Sistema de Dublagem**
   ```powershell
   .\dubbing-mvp.exe start --mode low-latency --use-vosk --use-argos --use-windows-tts --use-real-audio
   ```

2. **Aguardar Inicialização**
   - Espere ver: "🚀 Dubbing started!"
   - Espere ver: "🎙️  Pipeline running - speak in Portuguese!"

3. **Entrar no Google Meets**
   - Abra o navegador
   - Entre na reunião
   - Verifique que o microfone está configurado para "CABLE Output"

4. **Começar a Falar**
   - Fale normalmente em português
   - O sistema traduzirá automaticamente
   - Os outros participantes ouvirão em inglês

## 🎯 Modos de Performance

### Low-Latency Mode (Recomendado para Meets)
```powershell
.\dubbing-mvp.exe start --mode low-latency --use-vosk --use-argos --use-windows-tts --use-real-audio
```
- **Chunk Size**: 1s
- **Latência**: ~2-3s
- **Melhor para**: Conversas em tempo real
- **Auto-ativa**: Silence Detection + Metrics

### Balanced Mode (Padrão)
```powershell
.\dubbing-mvp.exe start --use-vosk --use-argos --use-windows-tts --use-real-audio
```
- **Chunk Size**: 2s
- **Latência**: ~3-4s
- **Melhor para**: Uso geral
- **Auto-ativa**: Silence Detection

### Quality Mode
```powershell
.\dubbing-mvp.exe start --mode quality --use-vosk --use-argos --use-windows-tts --use-real-audio
```
- **Chunk Size**: 3s
- **Latência**: ~4-5s
- **Melhor para**: Precisão sobre velocidade
- **Auto-ativa**: Nenhuma otimização

## 🔍 Troubleshooting

### Problema: "Ninguém me ouve no Meets"

**Solução**:
1. Verificar que o sistema de dublagem está rodando
2. Verificar que o Meets está usando "CABLE Output" como microfone
3. Falar em português e verificar os logs do sistema
4. Verificar volume do cabo virtual:
   - Abrir "Mixer de Volume" do Windows
   - Verificar que "CABLE Input" não está mudo

### Problema: "Latência muito alta"

**Solução**:
1. Usar modo low-latency:
   ```powershell
   .\dubbing-mvp.exe start --mode low-latency --use-vosk --use-argos --use-windows-tts --use-real-audio
   ```
2. Fechar outros programas pesados
3. Verificar uso de CPU (deve estar < 80%)
4. Considerar usar chunk size menor:
   ```powershell
   .\dubbing-mvp.exe start --chunk-size 1 --use-vosk --use-argos --use-windows-tts --use-real-audio
   ```

### Problema: "Sistema não reconhece minha voz"

**Solução**:
1. Verificar que o microfone está funcionando:
   ```powershell
   # Testar gravação do Windows
   # Configurações > Sistema > Som > Testar microfone
   ```
2. Verificar logs do sistema para erros
3. Ajustar threshold de silêncio se necessário
4. Falar mais alto ou mais próximo do microfone
5. Verificar que o modelo Vosk PT está instalado corretamente

### Problema: "Tradução incorreta"

**Solução**:
1. Falar mais devagar e claramente
2. Usar frases mais curtas
3. Verificar que o pacote PT→EN do Argos está instalado
4. Considerar usar modo quality para melhor precisão:
   ```powershell
   .\dubbing-mvp.exe start --mode quality --use-vosk --use-argos --use-windows-tts --use-real-audio
   ```

### Problema: "Áudio cortado ou com falhas"

**Solução**:
1. Aumentar buffer de áudio:
   - Editar `pkg/audio-capture-python/capture.go`
   - Aumentar `BufferSize` de 16000*10 para 16000*20
2. Verificar uso de CPU
3. Fechar outros programas de áudio
4. Reiniciar o sistema de dublagem

### Problema: "Erro ao iniciar Python scripts"

**Solução**:
1. Verificar que Python está no PATH:
   ```powershell
   python --version
   ```
2. Reinstalar dependências:
   ```powershell
   pip install --upgrade argostranslate pyttsx3 pywin32 vosk pyaudio
   ```
3. Verificar que os scripts estão na pasta `scripts/`:
   - `scripts/vosk-asr.py`
   - `scripts/windows-tts.py`
   - `scripts/audio-capture.py`

## 📊 Monitoramento de Performance

### Ver Métricas em Tempo Real

```powershell
.\dubbing-mvp.exe start --use-metrics --use-vosk --use-argos --use-windows-tts --use-real-audio
```

**Métricas exibidas**:
- Total de chunks processados
- Latência média, P50, P95, P99
- Taxa de acerto do cache
- Chunks de silêncio pulados
- Tempo de atividade

**Exemplo de saída**:
```
📊 Statistics:
  ASR:         10 chunks, avg latency: 2.1s
  Translation: 10 sentences, avg latency: 4.3s
  TTS:         10 sentences, avg latency: 0.6s

⚡ Performance Metrics:
  Total Chunks:    10
  Avg Latency:     7.0s
  P50 Latency:     6.8s
  P95 Latency:     8.2s
  P99 Latency:     8.5s
  Cache Hit Rate:  40.0% (4 hits, 6 misses)
  Silence Skips:   2
  Uptime:          2m30s

🔇 Silence Detection:
  Total Checks:    12
  Silence:         2 (16.7%)
  Speech:          10

💾 Translation Cache:
  Size:            6/1000 entries
  Hit Rate:        40.0% (4 hits, 6 misses)
```

## 🎓 Dicas de Uso

### Para Melhor Qualidade

1. **Use um bom microfone**
   - Microfone USB dedicado é melhor que integrado
   - Headset com cancelamento de ruído é ideal

2. **Ambiente silencioso**
   - Minimize ruído de fundo
   - Feche janelas
   - Desligue ventiladores

3. **Fale claramente**
   - Pronuncie bem as palavras
   - Não fale muito rápido
   - Faça pausas entre frases

4. **Use modo quality para apresentações**
   ```powershell
   .\dubbing-mvp.exe start --mode quality --use-vosk --use-argos --use-windows-tts --use-real-audio
   ```

### Para Melhor Performance

1. **Use modo low-latency**
   ```powershell
   .\dubbing-mvp.exe start --mode low-latency --use-vosk --use-argos --use-windows-tts --use-real-audio
   ```

2. **Feche programas desnecessários**
   - Navegadores com muitas abas
   - Editores de vídeo
   - Jogos

3. **Monitore o uso de recursos**
   - CPU deve estar < 80%
   - RAM deve ter pelo menos 2GB livre

4. **Use cabo Ethernet**
   - Conexão com fio é mais estável que WiFi
   - Importante para Google Meets

## 📝 Checklist Pré-Reunião

Antes de entrar em uma reunião do Google Meets:

- [ ] Sistema de dublagem compilado e testado
- [ ] Cabo de áudio virtual instalado e funcionando
- [ ] Modelo Vosk PT baixado e extraído
- [ ] Pacote Argos PT→EN instalado
- [ ] Microfone testado e funcionando
- [ ] Sistema de dublagem iniciado e processando
- [ ] Google Meets configurado para usar "CABLE Output"
- [ ] Teste de áudio realizado (falar PT, ouvir EN)
- [ ] Latência aceitável (< 5s)
- [ ] CPU usage aceitável (< 80%)

## 🆘 Suporte

### Logs e Debugging

Para ver logs detalhados:
```powershell
.\dubbing-mvp.exe start --use-vosk --use-argos --use-windows-tts --use-real-audio --use-metrics 2>&1 | Tee-Object -FilePath dubbing.log
```

Isso salvará todos os logs em `dubbing.log` para análise.

### Reportar Problemas

Se encontrar problemas:
1. Salvar os logs
2. Anotar a configuração usada
3. Descrever o comportamento esperado vs. observado
4. Incluir informações do sistema (Windows version, CPU, RAM)

## 🎉 Conclusão

Com este guia, você deve conseguir usar o sistema de dublagem em tempo real com Google Meets. O sistema permite que você fale em português e os outros participantes ouçam em inglês, tudo de forma automática e em tempo real!

**Lembre-se**:
- Latência típica: 2-3s (modo low-latency)
- 100% gratuito e offline
- Sem custos recorrentes
- Privacidade garantida (tudo local)

**Boa sorte com suas reuniões! 🚀**
