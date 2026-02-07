import { EventEmitter } from "events";

const MAX_RETRIES = 3;
let defaultName = "world";

interface Greeter {
  greet(name: string): string;
}

type PersonOptions = {
  name: string;
  age: number;
};

enum Status {
  Active = "active",
  Inactive = "inactive",
}

export class Person implements Greeter {
  name: string;
  age: number;

  constructor(options: PersonOptions) {
    this.name = options.name;
    this.age = options.age;
  }

  greet(name: string): string {
    return `Hello, ${name}! I'm ${this.name}.`;
  }

  get displayName(): string {
    return this.name.toUpperCase();
  }
}

export function sayHello(): void {
  const p = new Person({ name: defaultName, age: 30 });
  const msg = p.greet("friend");
  console.log(msg);
  for (let i = 0; i < MAX_RETRIES; i++) {
    console.log(p.greet(defaultName));
  }
  console.log(p.displayName);
}

export const formatName = (name: string): string => {
  return name.trim();
};

namespace Utils {
  export function helper(): void {}
}
