/**
 * Tricky JavaScript fixture for deep attribute analysis.
 */

const EventEmitter = require("events");
const path = require("path");

// --- Shadowing ---

const TIMEOUT = 30;

function shadowExample() {
  const TIMEOUT = 10; // shadows module-level
  console.log(TIMEOUT);
}

// --- Prototype-based inheritance ---

function Animal(name, sound) {
  this.name = name;
  this.sound = sound;
}

Animal.prototype.speak = function () {
  return `${this.name} says ${this.sound}`;
};

// --- Class inheritance ---

class BaseEntity {
  constructor(id) {
    this.id = id;
  }

  validate() {
    return this.id != null;
  }
}

class User extends BaseEntity {
  constructor(id, username, email) {
    super(id);
    this.username = username;
    this.email = email;
  }

  validate() {
    return super.validate() && this.username.length > 0;
  }

  // --- Getter/setter ---
  get displayName() {
    return `${this.username} <${this.email}>`;
  }

  set displayName(val) {
    const parts = val.split(" <");
    this.username = parts[0];
    this.email = parts[1]?.replace(">", "") || "";
  }

  // --- Static method ---
  static fromJSON(json) {
    const data = JSON.parse(json);
    return new User(data.id, data.username, data.email);
  }
}

// --- Destructuring and spread ---

function processConfig({ host, port, ...rest }) {
  console.log(host, port, rest);
}

// --- Builtins / globals (should NOT be unresolved external) ---

function useBuiltins() {
  // Global constructors
  const arr = Array.from([1, 2, 3]);
  const str = String(42);
  const num = Number("3.14");
  const bool = Boolean(0);
  const obj = Object.keys({ a: 1 });
  const p = Promise.resolve(42);
  const d = new Date();
  const r = new RegExp("\\d+");
  const m = new Map();
  const s = new Set();
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

function useExternal() {
  const emitter = new EventEmitter();
  emitter.on("data", () => {});
  const joined = path.join("/tmp", "file.txt");
  console.log(joined);
}

// --- IIFE ---

const Singleton = (() => {
  let instance = null;
  return {
    getInstance() {
      if (!instance) {
        instance = { created: Date.now() };
      }
      return instance;
    },
  };
})();

// --- Computed property names ---

const STATUS_KEY = "status";

class DynamicProps {
  constructor() {
    this[STATUS_KEY] = "active";
  }
}

// --- CommonJS exports ---

module.exports = {
  Animal,
  User,
  BaseEntity,
  processConfig,
  useBuiltins,
  Singleton,
  DynamicProps,
};
