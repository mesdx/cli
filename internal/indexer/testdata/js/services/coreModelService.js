const { CoreModel } = require("../models/coreModel");

/** CoreModelService extends CoreModel and manages a store of instances.
 * Demonstrates: import-require (0.40), extends (1.0 via lexical),
 * new CoreModel instantiation (0.95 via lexical), and instanceof (0.10). */
class CoreModelService extends CoreModel {
  constructor() {
    super("service-root", 1.0);
    this.store = [];
    const initial = new CoreModel("initial", 0.5);
    this.store.push(initial);
  }

  /** save stores a CoreModel if valid. */
  save(model) {
    if (model instanceof CoreModel && model.isValid()) {
      this.store.push(model);
      return true;
    }
    return false;
  }

  /** findByTitle returns the first CoreModel with matching title. */
  findByTitle(title) {
    return this.store.find((m) => m instanceof CoreModel && m.title === title) || null;
  }

  /** findAll returns every stored CoreModel. */
  findAll() {
    return this.store.filter((m) => m instanceof CoreModel);
  }

  /** create instantiates a fresh CoreModel. */
  static create(title, score) {
    return new CoreModel(title, score);
  }
}

/** storeCoreModel persists a CoreModel instance. */
function storeCoreModel(model) {
  return model instanceof CoreModel && model.isValid();
}

module.exports = { CoreModelService, storeCoreModel };
