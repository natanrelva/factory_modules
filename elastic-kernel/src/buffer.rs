pub struct RingBuffer {
    buffer: Vec<f32>,
    capacity: usize,
    write_pos: usize,
    read_pos: usize,
    count: usize,
}

impl RingBuffer {
    pub fn new(capacity: usize) -> Self {
        Self {
            buffer: vec![0.0; capacity],
            capacity,
            write_pos: 0,
            read_pos: 0,
            count: 0,
        }
    }

    pub fn push(&mut self, data: &[f32]) {
        for &sample in data {
            if self.count < self.capacity {
                self.buffer[self.write_pos] = sample;
                self.write_pos = (self.write_pos + 1) % self.capacity;
                self.count += 1;
            }
            // Em produção, lidaríamos com overflow (sobrescrever antigo ou rejeitar)
        }
    }

    pub fn pop(&mut self) -> Option<f32> {
        if self.count > 0 {
            let sample = self.buffer[self.read_pos];
            self.read_pos = (self.read_pos + 1) % self.capacity;
            self.count -= 1;
            Some(sample)
        } else {
            None
        }
    }
    
    // Útil para saber se estamos ficando sem dados (Underrun risk)
    pub fn len(&self) -> usize {
        self.count
    }
}
