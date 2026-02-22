import { CriticalUserModel } from "../models/criticalUserModel";

/** CriticalUserModelRepository manages CriticalUserModel persistence. */
export class CriticalUserModelRepository {
  private store: CriticalUserModel[] = [];

  save(model: CriticalUserModel): boolean {
    this.store.push(model);
    return true;
  }

  findByEmail(email: string): CriticalUserModel | undefined {
    return this.store.find((m) => m.email === email);
  }

  findAll(): CriticalUserModel[] {
    return this.store;
  }
}

/** storeCriticalUserModel persists a CriticalUserModel. */
export function storeCriticalUserModel(model: CriticalUserModel): boolean {
  return model.validate();
}
