/// AppProcessor handles data processing.
pub struct AppProcessor {
    pub name: String,
    pub workers: u32,
}

impl AppProcessor {
    /// Creates a new AppProcessor.
    pub fn new(name: String, workers: u32) -> Self {
        AppProcessor { name, workers }
    }

    /// Runs the processor.
    pub fn run(&self) {
        for i in 0..self.workers {
            println!("Worker {} started for {}", i, self.name);
        }
    }

    /// Stops the processor.
    pub fn stop(&self) {
        println!("Stopping {}", self.name);
    }
}
