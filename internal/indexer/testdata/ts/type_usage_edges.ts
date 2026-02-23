/**
 * type_usage_edges.ts — fixture testing that CoreModel is found as a ref
 * when it appears inside generic/container type expressions.
 *
 * Every function/type below uses CoreModel exclusively in type positions so
 * the fixture exercises TypeScript's type-identifier capture
 * (`(type_identifier) @ref.type`).
 */

import { CoreModel } from "./models/coreModel";

/** Receives an array of CoreModel values. */
export function takesArray(models: CoreModel[]): number {
    return models.length;
}

/** Receives a generic Array<CoreModel>. */
export function takesGenericArray(models: Array<CoreModel>): number {
    return models.length;
}

/** Returns a Promise that resolves to CoreModel. */
export async function returnsPromise(title: string): Promise<CoreModel> {
    return new CoreModel(title, 1.0);
}

/** Returns a Promise that resolves to an array of CoreModel. */
export async function returnsPromiseArray(): Promise<CoreModel[]> {
    return [];
}

/** ReadonlyArray<CoreModel> as parameter. */
export function takesReadonly(models: ReadonlyArray<CoreModel>): CoreModel | undefined {
    return models[0];
}

/** Record with CoreModel values. */
export function takesRecord(map: Record<string, CoreModel>): string[] {
    return Object.keys(map);
}

/** Map<string, CoreModel> as parameter. */
export function takesMap(map: Map<string, CoreModel>): number {
    return map.size;
}

/** Set<CoreModel> as parameter. */
export function takesSet(models: Set<CoreModel>): number {
    return models.size;
}

/** Deeply nested: Map<string, Array<CoreModel>>. */
export function deepNested(data: Map<string, Array<CoreModel>>): number {
    let total = 0;
    data.forEach(arr => { total += arr.length; });
    return total;
}

/** Local variable with CoreModel type. */
export function localVar(title: string): boolean {
    const m: CoreModel = new CoreModel(title, 0.5);
    return m.isValid();
}

/** Type alias using CoreModel. */
export type CoreModelList = CoreModel[];

/** Type alias: optional CoreModel. */
export type MaybeModel = CoreModel | null;

/** Interface with CoreModel property. */
export interface WithModel {
    model: CoreModel;
    models: CoreModel[];
}

/** Generic function constrained to CoreModel. */
export function identity<T extends CoreModel>(item: T): T {
    return item;
}
