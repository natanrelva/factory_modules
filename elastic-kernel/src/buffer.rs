pub struct RingBuffer {
    buffer: Vec<f32>,
    capacity: usize,
    write_pos: usize,
    read_pos: usize, // Agora atua como o "cursor base" (inteiro)
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
        }
    }

    // Pega o valor em 'offset' a partir do read_pos atual
    // offset 0 = read_pos, offset 1 = read_pos + 1
    pub fn get_relative(&self, offset: usize) -> f32 {
        if offset >= self.count {
            return 0.0; // Se pedir além do que temos, retorna silêncio
        }
        let index = (self.read_pos + offset) % self.capacity;
        self.buffer[index]
    }

    // Avança o cursor base e remove dados já consumidos
    pub fn advance(&mut self, amount: usize) {
        let to_remove = amount.min(self.count);
        self.read_pos = (self.read_pos + to_remove) % self.capacity;
        self.count -= to_remove;
    }
    
    pub fn len(&self) -> usize {
        self.count
    }
}
