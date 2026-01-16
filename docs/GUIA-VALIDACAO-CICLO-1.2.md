# 🧪 Guia de Validação - Ciclo 1.2

## O que estamos validando?

Queremos provar que o **algoritmo de Linear Resampling** funciona corretamente, permitindo alterar a velocidade de reprodução de áudio em tempo real.

### Critérios de Sucesso

✅ **Critério 1:** Consegue carregar um arquivo de áudio (MP3/WAV)
✅ **Critério 2:** O áudio toca normalmente em 1.0x
✅ **Critério 3:** Em 0.5x, a voz fica grave e lenta ("efeito monstro")
✅ **Critério 4:** Em 1.5x, a voz fica aguda e rápida ("efeito esquilo")
✅ **Critério 5:** Mudanças no slider são instantâneas, sem cliques ou cortes
✅ **Critério 6:** A barra de progresso do buffer funciona

---

## 🚀 Passo a Passo da Validação

### Passo 1: Preparar um arquivo de áudio

Você precisa de um arquivo de áudio curto (10-30 segundos) com voz falada.

**Opções:**
- Grave sua própria voz dizendo uma frase
- Use um arquivo MP3/WAV que você já tenha
- Baixe um sample de voz da internet

**Recomendação:** Voz falada funciona melhor para testar porque é fácil perceber mudanças de tom.

---

### Passo 2: Abrir o navegador

1. Certifique-se de que o servidor está rodando em `http://localhost:8080`
2. Abra o navegador (Chrome ou Edge)
3. Vá para: `http://localhost:8080/web-client`
4. Pressione `F12` para abrir o console

**O que você deve ver:**
- Título "Ciclo 1.2: Controle de Elasticidade"
- Botão "Escolher arquivo" para carregar áudio
- Slider de velocidade (0.1x a 2.0x)
- Barra de progresso do buffer

---

### Passo 3: Carregar o arquivo de áudio

1. Clique em "Escolher arquivo"
2. Selecione seu arquivo de áudio
3. Aguarde a mensagem "Upload completo para o buffer Rust!"

**O que observar:**
- Status deve mostrar: "Áudio decodificado: X.XXs. Enviando..."
- Depois: "Upload completo para o buffer Rust!"
- Barra de progresso deve começar a encher
- Console deve mostrar: "Kernel Rust inicializado no Worklet!"

---

### Passo 4: Testar velocidade normal (1.0x)

O áudio deve começar a tocar automaticamente após o upload.

**Verifique:**
- [ ] Áudio toca normalmente
- [ ] Sem distorção ou cliques
- [ ] Barra de progresso diminui conforme o áudio toca

---

### Passo 5: Testar velocidade lenta (0.5x)

1. Mova o slider para **0.5x** enquanto o áudio toca
2. Observe a mudança instantânea

**O que você deve ouvir:**
- Voz **grave** (tom mais baixo)
- Voz **lenta** (metade da velocidade)
- Efeito "monstro" ou "voz de demônio"
- **SEM:** Cliques, estalos ou interrupções

---

### Passo 6: Testar velocidade rápida (1.5x)

1. Mova o slider para **1.5x**
2. Observe a mudança instantânea

**O que você deve ouvir:**
- Voz **aguda** (tom mais alto)
- Voz **rápida** (1.5x mais rápida)
- Efeito "esquilo" ou "chipmunk"
- **SEM:** Cliques, estalos ou interrupções

---

### Passo 7: Testar mudanças dinâmicas

1. Mova o slider para frente e para trás várias vezes
2. Teste valores extremos: 0.1x, 2.0x

**O que verificar:**
- [ ] Mudanças são instantâneas
- [ ] Não há "pulos" ou "travamentos"
- [ ] O áudio continua fluindo suavemente
- [ ] Display mostra o valor correto (ex: "1.5x")

---

## 📊 Checklist de Validação

### Funcionalidade Básica
- [ ] Página carregou sem erros
- [ ] Console não mostra erros vermelhos
- [ ] Consegui carregar um arquivo de áudio

### Reprodução Normal (1.0x)
- [ ] Áudio toca normalmente
- [ ] Sem distorção
- [ ] Barra de progresso funciona

### Velocidade Lenta (0.5x)
- [ ] Voz fica grave
- [ ] Voz fica lenta
- [ ] Sem cliques ou cortes

### Velocidade Rápida (1.5x)
- [ ] Voz fica aguda
- [ ] Voz fica rápida
- [ ] Sem cliques ou cortes

### Mudanças Dinâmicas
- [ ] Mudanças são instantâneas
- [ ] Slider responde suavemente
- [ ] Sem travamentos

---

## 🐛 Problemas Comuns e Soluções

### Problema 1: "Não consigo carregar o arquivo"
**Possíveis causas:**
- Formato de arquivo não suportado
- Arquivo muito grande

**Solução:**
1. Use arquivos MP3 ou WAV
2. Mantenha o arquivo abaixo de 1 minuto
3. Verifique o console para erros

---

### Problema 2: "Áudio não toca"
**Causa:** AudioContext pode estar suspenso

**Solução:**
1. Clique em qualquer lugar da página
2. Recarregue o arquivo
3. Verifique se o volume do sistema não está mudo

---

### Problema 3: "Áudio corta ou picotar"
**Causa:** Buffer underrun (buffer vazio)

**Solução:**
1. Observe a barra de progresso
2. Se estiver vazia, o arquivo não foi carregado corretamente
3. Recarregue o arquivo

---

### Problema 4: "Mudanças no slider não fazem efeito"
**Causa:** WASM não inicializado ou erro de comunicação

**Solução:**
1. Verifique o console para erros
2. Recarregue a página (Ctrl + Shift + R)
3. Carregue o arquivo novamente

---

## 🎯 Testes Avançados (Opcional)

### Teste 1: Limites de Qualidade
Teste diferentes velocidades e anote quando a qualidade começa a degradar:

| Velocidade | Qualidade | Observações |
|------------|-----------|-------------|
| 0.1x | ? | Muito lento |
| 0.5x | ? | Efeito monstro |
| 0.8x | ? | Levemente grave |
| 1.0x | ✅ | Normal |
| 1.2x | ? | Levemente agudo |
| 1.5x | ? | Efeito esquilo |
| 2.0x | ? | Muito rápido |

### Teste 2: Latência de Mudança
Use o console para medir o tempo de resposta:

```javascript
// Cole no console
let startTime = performance.now();
// Mova o slider
// Quando ouvir a mudança, cole:
console.log('Latência:', performance.now() - startTime, 'ms');
```

---

## ✅ Validação Completa

Quando você conseguir:
1. ✅ Carregar um arquivo de áudio
2. ✅ Ouvir o áudio em 1.0x normalmente
3. ✅ Ouvir efeito "monstro" em 0.5x
4. ✅ Ouvir efeito "esquilo" em 1.5x
5. ✅ Mudanças instantâneas sem cliques

**Parabéns!** O Ciclo 1.2 está validado e você pode avançar para o **Ciclo 2: Jitter Buffer e Sincronização**.

---

## 📝 Relatório de Validação

Após testar, anote:

**Funcionou?** [ ] Sim [ ] Parcialmente [ ] Não

**Velocidades testadas:**
- 0.5x: [ ] OK [ ] Problemas
- 1.0x: [ ] OK [ ] Problemas
- 1.5x: [ ] OK [ ] Problemas

**Problemas encontrados:**
_______________________________________
_______________________________________

**Observações sobre qualidade:**
_______________________________________
_______________________________________

---

**Dúvidas?** Me envie:
- Screenshot do console
- Descrição do que você ouve
- Velocidades que funcionaram/não funcionaram
