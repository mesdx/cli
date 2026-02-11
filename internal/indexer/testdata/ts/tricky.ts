/**
 * Tricky TypeScript fixture for deep attribute analysis.
 */

import { EventEmitter } from "events";
import * as path from "path";

// --- Shadowing: block-scoped shadows module-level ---

const TIMEOUT = 30;

function shadowExample(): void {
  const TIMEOUT = 10; // shadows module-level
  console.log(TIMEOUT);
}

// --- Inheritance and implements ---

interface Serializable {
  serialize(): string;
}

interface Printable {
  display(): void;
}

abstract class BaseEntity {
  abstract id: number;
  abstract validate(): boolean;
}

class UserEntity extends BaseEntity implements Serializable, Printable {
  id: number;
  username: string;
  email: string;

  constructor(id: number, username: string, email: string) {
    super();
    this.id = id;
    this.username = username;
    this.email = email;
  }

  validate(): boolean {
    return this.username.length > 0 && this.email.includes("@");
  }

  serialize(): string {
    return JSON.stringify({ id: this.id, username: this.username });
  }

  display(): void {
    console.log(`User: ${this.username}`);
  }

  // --- Method overloading (TS-style) ---
  format(): string;
  format(includeEmail: boolean): string;
  format(includeEmail?: boolean): string {
    if (includeEmail) {
      return `${this.username} <${this.email}>`;
    }
    return this.username;
  }
}

// --- Generics ---

interface Repository<T extends BaseEntity> {
  save(entity: T): void;
  findById(id: number): T | undefined;
}

class InMemoryRepository<T extends BaseEntity> implements Repository<T> {
  private store: Map<number, T> = new Map();

  save(entity: T): void {
    this.store.set(entity.id, entity);
  }

  findById(id: number): T | undefined {
    return this.store.get(id);
  }
}

// --- Decorators (experimental) ---

function log(target: any, key: string, descriptor: PropertyDescriptor): PropertyDescriptor {
  const original = descriptor.value;
  descriptor.value = function (...args: any[]) {
    console.log(`Calling ${key}`);
    return original.apply(this, args);
  };
  return descriptor;
}

// --- Enum with computed values ---

enum HttpStatus {
  OK = 200,
  NotFound = 404,
  InternalError = 500,
}

// --- Const enum ---

const enum Direction {
  Up = "UP",
  Down = "DOWN",
  Left = "LEFT",
  Right = "RIGHT",
}

// --- Builtins / globals (should NOT be unresolved external) ---

function useBuiltins(): void {
  // Global constructors and objects
  const arr = Array.from([1, 2, 3]);
  const str = String(42);
  const num = Number("3.14");
  const bool = Boolean(0);
  const obj = Object.keys({ a: 1 });
  const p = Promise.resolve(42);
  const d = new Date();
  const r = new RegExp("\\d+");
  const m = new Map<string, number>();
  const s = new Set<number>();
  const e = new Error("test");
  const j = JSON.stringify({ key: "value" });
  const parsed = JSON.parse(j);

  // Global functions
  const t = setTimeout(() => {}, 1000);
  clearTimeout(t);
  const i = setInterval(() => {}, 1000);
  clearInterval(i);
  console.log(arr, str, num, bool, obj, p, d, r, m, s, e, parsed);

  // Math
  const abs = Math.abs(-1);
  const pi = Math.PI;
  console.log(abs, pi);
}

// --- External references ---

function useExternal(): void {
  const emitter = new EventEmitter();
  emitter.on("data", () => {});
  const joined = path.join("/tmp", "file.txt");
  console.log(joined);
}

// --- Type aliases and union types ---

type StringOrNumber = string | number;
type Nullable<T> = T | null;
type UserDTO = Pick<UserEntity, "username" | "email">;

// --- Mapped / conditional types ---

type ReadonlyUser = Readonly<UserEntity>;
type PartialUser = Partial<UserEntity>;

// --- Intersection types ---

type Named = { name: string };
type Aged = { age: number };
type NamedAged = Named & Aged;

// --- Re-export pattern ---

export { UserEntity, BaseEntity, HttpStatus };
export type { Repository, Serializable };

// --- Default export ---

export default class App {
  name: string;
  constructor(name: string) {
    this.name = name;
  }
}
