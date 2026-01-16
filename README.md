# Dubbing POC - Ciclo 1.1: DSP Setup

Projeto de prova de conceito para processamento de áudio em tempo real usando Rust/WASM.

## Estrutura

```
dubbing-poc-cycle1/
├── elastic-kernel/       # Código Rust (Crate)
│   ├── src/
│   │   ├── lib.rs        # Ponto de entrada
│   │   └── buffer.rs     # Ring Buffer
│   └── Cargo.toml
├── web-client/           # Frontend
│   ├── index.html
│   ├── main.js           # Thread principal UI
│   └── processor.js      # AudioWorklet
```

## Instruções de Execução

### 1. Compilar o Rust

No terminal, dentro da pasta `elastic-kernel`:

```bash
wasm-pack build --target web
```

### 2. Rodar Servidor Web

Na raiz do projeto:

**Python:**
```bash
python -m http.server 8080
```

**Ou use a extensão "Live Server" do VSCode**

### 3. Testar

1. Abra `http://localhost:8080/web-client`
2. Abra o Console do navegador (F12)
3. Clique em **INICIAR ÁUDIO**

## Critério de Sucesso

- Você deve ouvir um tom senoidal contínuo (440Hz)
- No console, não deve haver erros de carregamento de WASM
- Mensagem "Kernel Rust inicializado no Worklet!" deve aparecer no console
