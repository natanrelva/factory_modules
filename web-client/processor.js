class ElasticProcessorNode extends AudioWorkletProcessor {
    constructor() {
        super();
        this.audioBuffer = new Float32Array(128);
        this.port.onmessage = (event) => {
            if (event.data.type === 'audio') {
                this.audioBuffer = event.data.data;
            }
        };
        // Sinaliza que está pronto
        this.port.postMessage('ready');
        console.log("AudioWorklet inicializado!");
    }

    process(inputs, outputs, parameters) {
        const output = outputs[0];
        const channel = output[0]; // Mono

        // Copia o buffer recebido da thread principal
        channel.set(this.audioBuffer);

        // Copiar para os outros canais se houver (estéreo)
        for (let i = 1; i < output.length; i++) {
            output[i].set(channel);
        }

        return true;
    }
}

registerProcessor('elastic-processor-node', ElasticProcessorNode);
