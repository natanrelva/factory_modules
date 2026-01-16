import init, { ElasticProcessor } from '../elastic-kernel/pkg/elastic_kernel.js';

const audioFile = document.getElementById('audioFile');
const speedSlider = document.getElementById('speedSlider');
const speedDisplay = document.getElementById('speedDisplay');
const bufferHealth = document.getElementById('bufferHealth');
const status = document.getElementById('status');

let audioContext = null;
let elasticNode = null;

// Função para enviar chunks de dados para o Worklet
async function sendAudioDataToWorklet(audioBuffer) {
    const rawData = audioBuffer.getChannelData(0); // Pega apenas canal esquerdo (Mono)
    
    // Enviamos em chunks de 4096 amostras para não travar a mensageria
    const chunkSize = 4096;
    for (let i = 0; i < rawData.length; i += chunkSize) {
        const chunk = rawData.slice(i, i + chunkSize);
        elasticNode.port.postMessage({ type: 'data', chunk: chunk });
        
        // Pequeno delay artificial para simular streaming de rede (Opcional, mas bom para teste)
        // await new Promise(r => setTimeout(r, 10)); 
    }
    status.innerText = "Upload completo para o buffer Rust!";
}

async function setup() {
    await init();
    audioContext = new AudioContext();
    await audioContext.audioWorklet.addModule('processor.js');
    
    elasticNode = new AudioWorkletNode(audioContext, 'elastic-processor-node');
    elasticNode.connect(audioContext.destination);
    
    // Inicializa WASM no Worklet
    elasticNode.port.postMessage({ type: 'init' });

    // Monitoramento do Buffer (Loop de UI)
    setInterval(() => {
        // Pedimos o status para o worklet (feedback loop)
        elasticNode.port.postMessage({ type: 'get_health' });
    }, 100);

    // Recebe atualizações do Worklet
    elasticNode.port.onmessage = (e) => {
        if (e.data.type === 'health') {
            bufferHealth.value = e.data.value;
        }
    };
}

// Evento: Mudar Velocidade
speedSlider.oninput = (e) => {
    const val = parseFloat(e.target.value);
    speedDisplay.innerText = val + "x";
    if (elasticNode) {
        elasticNode.port.postMessage({ type: 'speed', value: val });
    }
};

// Evento: Carregar Arquivo
audioFile.onchange = async (e) => {
    if (!audioContext) await setup();
    
    const file = e.target.files[0];
    const arrayBuffer = await file.arrayBuffer();
    const audioBuffer = await audioContext.decodeAudioData(arrayBuffer);
    
    status.innerText = `Áudio decodificado: ${audioBuffer.duration.toFixed(2)}s. Enviando...`;
    
    // Retomar contexto se estiver suspenso (política de autoplay)
    if (audioContext.state === 'suspended') audioContext.resume();
    
    sendAudioDataToWorklet(audioBuffer);
};
