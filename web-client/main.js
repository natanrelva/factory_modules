// Importa o WASM compilado (assumindo que rodaremos server na raiz)
import init, { ElasticProcessor } from '../elastic-kernel/pkg/elastic_kernel.js';

const startBtn = document.getElementById('startBtn');
const status = document.getElementById('status');

startBtn.onclick = async () => {
    try {
        await init(); // Inicializa o WASM
        status.innerText = "WASM Carregado. Iniciando AudioContext...";

        const audioContext = new AudioContext();
        
        // Adiciona o módulo do processador (Thread separada)
        await audioContext.audioWorklet.addModule('processor.js');

        // Cria o Node que usa nosso script
        const elasticNode = new AudioWorkletNode(audioContext, 'elastic-processor-node');

        // Conecta à saída (alto-falantes)
        elasticNode.connect(audioContext.destination);

        // Cria o processador Rust na thread principal e processa aqui
        const kernel = ElasticProcessor.new(44100);
        
        // Envia dados processados para o worklet
        const processAudio = () => {
            const buffer = new Float32Array(128);
            kernel.process(buffer);
            elasticNode.port.postMessage({ type: 'audio', data: buffer });
            requestAnimationFrame(processAudio);
        };
        
        elasticNode.port.onmessage = (event) => {
            if (event.data === 'ready') {
                processAudio();
            }
        };

        status.innerText = "Rodando! Você deve ouvir um tom de 440Hz (Buffer vazio).";
    } catch (error) {
        console.error("Erro ao inicializar:", error);
        status.innerText = "ERRO: " + error.message;
    }
};
