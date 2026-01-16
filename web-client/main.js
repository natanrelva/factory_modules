// Importa o WASM compilado (assumindo que rodaremos server na raiz)
import init, { ElasticProcessor } from '../elastic-kernel/pkg/elastic_kernel.js';

const startBtn = document.getElementById('startBtn');
const status = document.getElementById('status');

startBtn.onclick = async () => {
    await init(); // Inicializa o WASM
    status.innerText = "WASM Carregado. Iniciando AudioContext...";

    const audioContext = new AudioContext();
    
    // Adiciona o módulo do processador (Thread separada)
    await audioContext.audioWorklet.addModule('processor.js');

    // Cria o Node que usa nosso script
    const elasticNode = new AudioWorkletNode(audioContext, 'elastic-processor-node');

    // Conecta à saída (alto-falantes)
    elasticNode.connect(audioContext.destination);

    // Manda sinal para inicializar o WASM lá dentro
    elasticNode.port.postMessage('init');

    status.innerText = "Rodando! Você deve ouvir um tom de 440Hz (Buffer vazio).";
};
