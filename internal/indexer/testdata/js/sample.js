const MAX_RETRIES = 3;
let defaultName = "world";

class Person {
  constructor(name, age) {
    this.name = name;
    this.age = age;
  }

  greet(name) {
    return `Hello, ${name}! I'm ${this.name}.`;
  }

  get displayName() {
    return this.name.toUpperCase();
  }
}

function sayHello() {
  const p = new Person(defaultName, 30);
  const msg = p.greet("friend");
  console.log(msg);
  for (let i = 0; i < MAX_RETRIES; i++) {
    console.log(p.greet(defaultName));
  }
  console.log(p.displayName);
}

const formatName = (name) => {
  return name.trim();
};

module.exports = { Person, sayHello, formatName };
