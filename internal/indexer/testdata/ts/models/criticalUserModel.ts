/** BaseModel provides base entity identity fields. */
export class BaseModel {
  id: number;
  createdAt: string;

  constructor(id: number = 0) {
    this.id = id;
    this.createdAt = new Date().toISOString();
  }
}

/** CriticalUserModel is a heavily-used user representation that many files depend on. */
export class CriticalUserModel extends BaseModel {
  name: string;
  email: string;

  constructor(name: string, email: string) {
    super(0);
    this.name = name;
    this.email = email;
  }

  validate(): boolean {
    return this.name.length > 0 && this.email.includes("@");
  }

  display(): string {
    return `${this.name} <${this.email}>`;
  }
}

/** CriticalUserModelOptions describes creation options. */
export type CriticalUserModelOptions = {
  name: string;
  email: string;
};

/** createCriticalUserModel is a factory function. */
export function createCriticalUserModel(opts: CriticalUserModelOptions): CriticalUserModel {
  return new CriticalUserModel(opts.name, opts.email);
}
