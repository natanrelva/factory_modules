use wasm_bindgen::prelude::*;
mod buffer;
use buffer::RingBuffer;

#[wasm_bindgen]
pub struct ElasticProcessor {
    buffer: RingBuffer,
    cursor_fract: f32, // A parte fracionária da posição de leitura (0.0 a 0.999...)
    playback_rate: f32, // 1.0 = Normal, 0.5 = Lento
}

#[wasm_bindgen]
impl ElasticProcessor {
    pub fn new(buffer_size: usize) -> Self {
        Self {
            buffer: RingBuffer::new(buffer_size),
            cursor_fract: 0.0,
            playback_rate: 1.0,
        }
    }

    pub fn push_data(&mut self, data: &[f32]) {
        self.buffer.push(data);
    }

    pub fn set_playback_rate(&mut self, rate: f32) {
        // Limitamos para evitar crashes ou som inaudível
        self.playback_rate = rate.max(0.1).min(4.0);
    }
    
    // Retorna o nível de preenchimento (0.0 a 1.0) para visualização no JS
    pub fn get_buffer_health(&self) -> f32 {
        self.buffer.len() as f32 / 44100.0 // Assumindo buffer de ~1s
    }

    pub fn process(&mut self, output: &mut [f32]) {
        for sample in output.iter_mut() {
            // Precisamos de pelo menos 2 amostras para interpolar
            if self.buffer.len() < 2 {
                *sample = 0.0;
                continue;
            }

            // 1. Ler as duas amostras vizinhas (Floor e Ceiling)
            let s1 = self.buffer.get_relative(0);
            let s2 = self.buffer.get_relative(1);

            // 2. Matemática da Interpolação Linear
            // y = s1 + (s2 - s1) * fraction
            *sample = s1 + (s2 - s1) * self.cursor_fract;

            // 3. Avançar o cursor virtual baseada na velocidade
            self.cursor_fract += self.playback_rate;

            // 4. Se a fração passou de 1.0, avançamos o cursor real do buffer (inteiro)
            while self.cursor_fract >= 1.0 {
                self.buffer.advance(1);
                self.cursor_fract -= 1.0;
            }
        }
    }
}
