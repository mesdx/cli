import { CoreModel } from "../models/coreModel";

/** CoreModelService extends CoreModel and manages a collection of instances.
 * Demonstrates: import (0.40), extends (1.0 via lexical), instantiation new (0.95),
 * method calls (0.65), and type annotations (0.25). */
export class CoreModelService extends CoreModel {
  private store: CoreModel[] = [];

  constructor() {
    super("service-root", 1.0);
    const initial = new CoreModel("initial", 0.5);
    this.store.push(initial);
  }

  /** save stores a CoreModel entry. */
  save(model: CoreModel): boolean {
    if (model.isValid()) {
      this.store.push(model);
      return true;
    }
    return false;
  }

  /** findByTitle looks up a CoreModel by its title. */
  findByTitle(title: string): CoreModel | undefined {
    return this.store.find((m) => m.title === title);
  }

  /** findAll returns every stored CoreModel. */
  findAll(): CoreModel[] {
    return this.store;
  }

  /** create instantiates a fresh CoreModel. */
  static create(title: string, score: number): CoreModel {
    return new CoreModel(title, score);
  }
}

/** storeCoreModel persists a CoreModel instance. */
export function storeCoreModel(model: CoreModel): boolean {
  return model.isValid();
}
