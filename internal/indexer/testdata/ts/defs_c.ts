import { CriticalUserModel } from "./models/criticalUserModel";

/** CriticalUserModelCache caches CriticalUserModel instances. */
export class CriticalUserModelCache {
  private entries: Map<number, CriticalUserModel> = new Map();

  put(id: number, model: CriticalUserModel): void {
    this.entries.set(id, model);
  }

  get(id: number): CriticalUserModel | undefined {
    return this.entries.get(id);
  }

  values(): CriticalUserModel[] {
    return Array.from(this.entries.values());
  }
}

/** transformCriticalUserModel transforms a CriticalUserModel. */
export function transformCriticalUserModel(model: CriticalUserModel): string {
  return "[" + model.name + "]";
}
