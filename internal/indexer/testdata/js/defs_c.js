const { CriticalUserModel } = require("./models/criticalUserModel");

class CriticalUserModelCache {
  constructor() {
    this.entries = new Map();
  }

  put(id, model) {
    this.entries.set(id, model);
  }

  get(id) {
    return this.entries.get(id) || null;
  }

  values() {
    return Array.from(this.entries.values());
  }
}

/** Returns a tagged string for a CriticalUserModel. */
function transformCriticalUserModel(model) {
  if (!(model instanceof CriticalUserModel)) {
    return null;
  }
  return "[" + model.name + "]";
}

/** Creates a wrapped CriticalUserModel entry. */
function wrapCriticalUserModel(name, email) {
  const m = new CriticalUserModel(name, email);
  return { model: m, tag: "wrapped" };
}

module.exports = { CriticalUserModelCache, transformCriticalUserModel, wrapCriticalUserModel };
