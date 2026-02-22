const { CriticalUserModel } = require("../models/criticalUserModel");

class CriticalUserModelRenderer {
  constructor(model) {
    this.model = model;
  }

  render() {
    return this.model.display();
  }
}

function renderCriticalUserModel(m) {
  return m.name + " <" + m.email + ">";
}

/** Creates a new CriticalUserModel directly. */
function buildCriticalUserModel(name, email) {
  return new CriticalUserModel(name, email);
}

/** Checks if the value is a CriticalUserModel instance. */
function isCriticalUserModel(val) {
  return val instanceof CriticalUserModel;
}

module.exports = { CriticalUserModelRenderer, renderCriticalUserModel, buildCriticalUserModel, isCriticalUserModel };
