# 🧪 Guia de Validação - Ciclo 1.1

## O que estamos validando?

Queremos provar que o **pipeline Rust → WASM → AudioWorklet → Alto-falantes** está funcionando corretamente.

### Critérios de Sucesso

✅ **Critério 1:** Você ouve um tom contínuo de 440Hz (nota Lá)
✅ **Critério 2:** Não há erros no console do navegador
✅ **Critério 3:** A mensagem "Kernel Rust inicializado no Worklet!" aparece no console
✅ **Critério 4:** O áudio não trava, não corta, não tem cliques

---

## 🚀 Passo a Passo da Validação

### Passo 1: Verificar se o servidor está rodando

Abra o terminal e veja se há uma mensagem como:
```
Starting up http-server, serving ./
Available on:
  http://127.0.0.1:8080
  http://192.168.x.x:8080
```

Se não estiver rodando, execute:
```bash
npx http-server -p 8080
```

---

### Passo 2: Abrir o navegador

1. Abra o **Google Chrome** ou **Microsoft Edge** (recomendado)
2. Digite na barra de endereço: `http://localhost:8080/web-client`
3. Pressione Enter

**O que você deve ver:**
- Uma página com o título "Ciclo 1.1: Validação do Kernel Rust"
- Um botão "INICIAR ÁUDIO (Worklet)"
- Texto "Status: Parado"

---

### Passo 3: Abrir o Console do Desenvolvedor

**Antes de clicar no botão**, abra o console:

- **Atalho:** Pressione `F12` ou `Ctrl + Shift + I`
- Ou clique com botão direito → "Inspecionar" → aba "Console"

**Por que fazer isso?**
Precisamos ver as mensagens de log e possíveis erros.

---

### Passo 4: Clicar no botão "INICIAR ÁUDIO"

1. Clique no botão **"INICIAR ÁUDIO (Worklet)"**
2. **IMPORTANTE:** O navegador pode pedir permissão para reproduzir áudio - clique em "Permitir"

---

### Passo 5: Observar e Anotar

#### 🎧 O que você deve OUVIR:
- Um **tom contínuo** (som de "biiiiiiiiip")
- Frequência: 440Hz (nota Lá - mesma nota de afinação de instrumentos)
- Volume: Baixo (10% do máximo)
- **SEM:** Cliques, estalos, cortes, travamentos

#### 👀 O que você deve VER no Console:
Procure por estas mensagens (na ordem):
```
1. "WASM Carregado. Iniciando AudioContext..."
2. "Kernel Rust inicializado no Worklet!"
3. Status na página muda para: "Rodando! Você deve ouvir um tom de 440Hz..."
```

#### ❌ O que NÃO deve aparecer:
- Erros em vermelho no console
- Mensagens de "Failed to load"
- Mensagens de "CORS error"
- Mensagens de "Module not found"

---

## 📊 Checklist de Validação

Preencha mentalmente ou anote:

### Funcionalidade Básica
- [ ] Página carregou sem erros
- [ ] Botão está visível e clicável
- [ ] Console não mostra erros vermelhos ao carregar

### Após Clicar no Botão
- [ ] Status mudou para "WASM Carregado..."
- [ ] Status mudou para "Rodando!..."
- [ ] Console mostra "Kernel Rust inicializado no Worklet!"
- [ ] Áudio está tocando (tom contínuo)

### Qualidade do Áudio
- [ ] Tom é contínuo (não para)
- [ ] Não há cliques ou estalos
- [ ] Volume está audível mas não alto demais
- [ ] Não há distorção

---

## 🐛 Problemas Comuns e Soluções

### Problema 1: "Não ouço nada"
**Possíveis causas:**
- Volume do sistema está mudo
- Navegador bloqueou autoplay de áudio
- Fones de ouvido desconectados

**Solução:**
1. Verifique o volume do Windows
2. Clique no botão novamente
3. Verifique se o navegador não bloqueou áudio (ícone de som na barra de endereço)

---

### Problema 2: Erro "Failed to fetch" ou "CORS"
**Causa:** Servidor HTTP não está rodando ou arquivo WASM não foi encontrado

**Solução:**
1. Verifique se o servidor está rodando na porta 8080
2. Verifique se a pasta `elastic-kernel/pkg/` existe
3. Se não existir, compile novamente:
   ```bash
   cd elastic-kernel
   wasm-pack build --target web
   ```

---

### Problema 3: Erro "Module not found"
**Causa:** Caminho do import está incorreto

**Solução:**
Verifique se a estrutura de pastas está assim:
```
C:\factory_modules\
├── elastic-kernel/
│   └── pkg/              ← Esta pasta deve existir
│       ├── elastic_kernel.js
│       ├── elastic_kernel_bg.wasm
│       └── ...
└── web-client/
    ├── index.html
    ├── main.js
    └── processor.js
```

---

### Problema 4: Áudio corta ou trava
**Causa:** Possível problema de performance ou buffer

**Solução:**
1. Feche outras abas do navegador
2. Verifique se há mensagens de erro no console
3. Anote o comportamento para ajustar no Ciclo 1.2

---

## 📝 Após a Validação

### Se TUDO funcionou (✅ Sucesso Total):
Atualize o arquivo `docs/ciclo-1.1-feedback.md`:
- Mude o status para: **✅ SUCESSO**
- Adicione as métricas observadas
- Anote qualquer observação sobre qualidade do áudio

### Se funcionou PARCIALMENTE (⚠️):
- Anote quais critérios passaram e quais falharam
- Documente os erros específicos
- Mantenha status como **⚠️ SUCESSO PARCIAL**

### Se NÃO funcionou (❌):
- Copie TODOS os erros do console
- Tire um screenshot da página
- Me envie essas informações para debug

---

## 🎯 Métricas para Coletar (Opcional - Avançado)

Se quiser ir além, cole isso no console do navegador após clicar no botão:

```javascript
// Medir latência do AudioContext
console.log('Latência do AudioContext:', audioContext.baseLatency * 1000, 'ms');
console.log('Sample Rate:', audioContext.sampleRate, 'Hz');
```

Anote esses valores no documento de feedback.

---

## ✅ Validação Completa

Quando você conseguir:
1. ✅ Ouvir o tom de 440Hz
2. ✅ Ver a mensagem "Kernel Rust inicializado no Worklet!"
3. ✅ Não ter erros no console

**Parabéns!** O Ciclo 1.1 está validado e você pode avançar para o **Ciclo 1.2: Algoritmo de Elasticidade**.

---

**Dúvidas?** Me envie:
- Screenshot do console
- Descrição do que você ouve (ou não ouve)
- Mensagens de erro (se houver)
