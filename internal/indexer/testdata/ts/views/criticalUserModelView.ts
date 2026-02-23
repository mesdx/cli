import { CriticalUserModel, createCriticalUserModel } from "../models/criticalUserModel";

/** CriticalUserModelRenderer renders a CriticalUserModel for display. */
export class CriticalUserModelRenderer {
  model: CriticalUserModel;

  constructor(model: CriticalUserModel) {
    this.model = model;
  }

  render(): string {
    return this.model.display();
  }
}

/** renderCriticalUserModel formats a CriticalUserModel as a display string. */
export function renderCriticalUserModel(m: CriticalUserModel): string {
  return m.name + " <" + m.email + ">";
}

/** buildCriticalUserModel creates a CriticalUserModel from parts. */
export function buildCriticalUserModel(name: string, email: string): CriticalUserModel {
  return createCriticalUserModel({ name, email });
}
