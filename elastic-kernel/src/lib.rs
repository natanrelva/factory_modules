use wasm_bindgen::prelude::*;
mod buffer;
use buffer::RingBuffer;

#[wasm_bindgen]
pub struct ElasticProcessor {
    buffer: RingBuffer,
    // Apenas para teste desta etapa: um oscilador interno
    // para provar que o Rust está gerando som se o buffer estiver vazio.
    test_phase: f32, 
}

#[wasm_bindgen]
impl ElasticProcessor {
    pub fn new(buffer_size: usize) -> Self {
        Self {
            buffer: RingBuffer::new(buffer_size),
            test_phase: 0.0,
        }
    }

    pub fn push_data(&mut self, data: &[f32]) {
        self.buffer.push(data);
    }

    // Chamado a cada frame de áudio pelo navegador (geralmente 128 amostras)
    pub fn process(&mut self, output: &mut [f32]) {
        for sample in output.iter_mut() {
            // Tenta pegar do buffer
            if let Some(val) = self.buffer.pop() {
                *sample = val;
            } else {
                // FALLBACK DE TESTE: 
                // Se o buffer estiver vazio, gera uma onda senoidal suave (440Hz)
                // Isso confirma que o Rust está rodando, mesmo sem input.
                *sample = (self.test_phase * 2.0 * std::f32::consts::PI).sin() * 0.1;
                self.test_phase = (self.test_phase + 440.0 / 44100.0) % 1.0;
            }
        }
    }
}
