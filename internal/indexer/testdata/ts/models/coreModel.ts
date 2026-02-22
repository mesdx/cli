/** BaseEntity provides shared identity fields. */
export class BaseEntity {
  id: number;
  createdAt: string;

  constructor(id: number = 0) {
    this.id = id;
    this.createdAt = new Date().toISOString();
  }
}

/** CoreModel is a high-usage domain model referenced by many files.
 * Used as a fixture to verify coupling score distribution for popular symbols. */
export class CoreModel extends BaseEntity {
  title: string;
  score: number;

  constructor(title: string, score: number) {
    super(0);
    this.title = title;
    this.score = score;
  }

  isValid(): boolean {
    return this.title.length > 0 && this.score >= 0;
  }

  describe(): string {
    return this.title;
  }
}

/** CoreModelOptions describes factory options. */
export type CoreModelOptions = {
  title: string;
  score: number;
};

/** createCoreModel instantiates a CoreModel from options. */
export function createCoreModel(opts: CoreModelOptions): CoreModel {
  return new CoreModel(opts.title, opts.score);
}
