/**
 * AppProcessorJs handles data processing (JS).
 */
class AppProcessorJs {
  constructor(name, workers = 1) {
    this.name = name;
    this.workers = workers;
  }

  run() {
    for (let i = 0; i < this.workers; i++) {
      console.log(`Worker ${i} started for ${this.name}`);
    }
  }

  stop() {
    console.log(`Stopping ${this.name}`);
  }
}

module.exports = { AppProcessorJs };
