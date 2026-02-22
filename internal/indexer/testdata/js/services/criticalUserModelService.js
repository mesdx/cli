const { CriticalUserModel } = require("../models/criticalUserModel");

class CriticalUserModelRepository {
  constructor() {
    this.store = [];
  }

  save(model) {
    this.store.push(model);
    return true;
  }

  findByEmail(email) {
    return this.store.find((m) => m.email === email) || null;
  }

  findAll() {
    return this.store;
  }
}

function storeCriticalUserModel(model) {
  return model instanceof CriticalUserModel && model.validate();
}

module.exports = { CriticalUserModelRepository, storeCriticalUserModel };
