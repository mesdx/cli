/**
 * AppProcessor handles data processing.
 */
export class AppProcessor {
  name: string;
  workers: number;

  constructor(name: string, workers: number = 1) {
    this.name = name;
    this.workers = workers;
  }

  run(): void {
    for (let i = 0; i < this.workers; i++) {
      console.log(`Worker ${i} started for ${this.name}`);
    }
  }

  stop(): void {
    console.log(`Stopping ${this.name}`);
  }
}
