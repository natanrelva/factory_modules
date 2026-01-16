import init, { ElasticProcessor } from '../elastic-kernel/pkg/elastic_kernel.js';

class ElasticProcessorNode extends AudioWorkletProcessor {
    constructor() {
        super();
        this.initialized = false;
        this.port.onmessage = async (event) => {
            if (event.data === 'init') {
                // Worklets tem escopo isolado, precisamos reinicializar o WASM aqui ou passar memória
                // Para esta POC 1.1, vamos carregar o WASM direto aqui
                await init('../elastic-kernel/pkg/elastic_kernel_bg.wasm');
                this.kernel = ElasticProcessor.new(44100); // 1 segundo de buffer hipotético
                this.initialized = true;
                console.log("Kernel Rust inicializado no Worklet!");
            }
        };
    }

    process(inputs, outputs, parameters) {
        if (!this.initialized) return true;

        const output = outputs[0];
        const channel = output[0]; // Mono

        // Passamos o array do JS para o Rust preencher
        // Em Rust/WASM isso envolve copiar memória. 
        // Para 1.1, isso é aceitável.
        this.kernel.process(channel);

        // Copiar para os outros canais se houver (estéreo)
        for (let i = 1; i < output.length; i++) {
            output[i].set(channel);
        }

        return true;
    }
}

registerProcessor('elastic-processor-node', ElasticProcessorNode);
