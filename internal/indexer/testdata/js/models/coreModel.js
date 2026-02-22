/** BaseEntity provides shared identity fields. */
class BaseEntity {
  constructor(id = 0) {
    this.id = id;
    this.createdAt = new Date().toISOString();
  }
}

/** CoreModel is a high-usage domain model referenced by many files.
 * Used as a fixture to verify coupling score distribution for popular symbols. */
class CoreModel extends BaseEntity {
  constructor(title, score) {
    super(0);
    this.title = title;
    this.score = score;
  }

  isValid() {
    return this.title.length > 0 && this.score >= 0;
  }

  describe() {
    return this.title;
  }
}

/** createCoreModel instantiates a CoreModel from raw fields. */
function createCoreModel(title, score) {
  return new CoreModel(title, score);
}

module.exports = { BaseEntity, CoreModel, createCoreModel };
