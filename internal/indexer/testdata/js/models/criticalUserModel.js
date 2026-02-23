/** BaseModel provides base entity identity fields. */
class BaseModel {
  constructor(id = 0) {
    this.id = id;
    this.createdAt = new Date().toISOString();
  }
}

/** CriticalUserModel is a heavily-used user representation that many files depend on. */
class CriticalUserModel extends BaseModel {
  constructor(name, email) {
    super(0);
    this.name = name;
    this.email = email;
  }

  validate() {
    return this.name.length > 0 && this.email.includes("@");
  }

  display() {
    return `${this.name} <${this.email}>`;
  }
}

function createCriticalUserModel(name, email) {
  return new CriticalUserModel(name, email);
}

module.exports = { BaseModel, CriticalUserModel, createCriticalUserModel };
