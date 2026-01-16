import init, { ElasticProcessor } from '../elastic-kernel/pkg/elastic_kernel.js';

class ElasticProcessorNode extends AudioWorkletProcessor {
    constructor() {
        super();
        this.initialized = false;
        
        this.port.onmessage = async (event) => {
            const msg = event.data;
            
            if (msg.type === 'init') {
                await init('../elastic-kernel/pkg/elastic_kernel_bg.wasm');
                // Buffer grande de 10 segundos para testar carregamento de arquivo
                this.kernel = ElasticProcessor.new(44100 * 10); 
                this.initialized = true;
                console.log("Kernel Rust inicializado no Worklet!");
            } 
            else if (msg.type === 'data' && this.initialized) {
                this.kernel.push_data(msg.chunk);
            }
            else if (msg.type === 'speed' && this.initialized) {
                this.kernel.set_playback_rate(msg.value);
            }
            else if (msg.type === 'get_health' && this.initialized) {
                const health = this.kernel.get_buffer_health();
                this.port.postMessage({ type: 'health', value: health });
            }
        };
    }

    process(inputs, outputs, parameters) {
        if (!this.initialized) return true;

        const output = outputs[0];
        const channel = output[0]; // Mono

        this.kernel.process(channel);

        // Copia mono para todos canais de saída
        for (let i = 1; i < output.length; i++) {
            output[i].set(channel);
        }

        return true;
    }
}

registerProcessor('elastic-processor-node', ElasticProcessorNode);
