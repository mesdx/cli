use std::fmt;

const MAX_RETRIES: u32 = 3;

pub struct Person {
    pub name: String,
    pub age: u32,
}

pub enum Status {
    Active,
    Inactive,
}

pub trait Greeter {
    fn greet(&self, name: &str) -> String;
}

impl Greeter for Person {
    fn greet(&self, name: &str) -> String {
        format!("Hello, {}! I'm {}.", name, self.name)
    }
}

impl Person {
    pub fn new(name: String, age: u32) -> Self {
        Person { name, age }
    }

    pub fn say_hello(&self) {
        let msg = self.greet("friend");
        println!("{}", msg);
        for _ in 0..MAX_RETRIES {
            println!("{}", self.greet(&self.name));
        }
    }
}

mod helpers {
    pub fn format_name(name: &str) -> String {
        name.to_uppercase()
    }
}

type PersonAlias = Person;

fn main() {
    let p = Person::new("Alice".to_string(), 30);
    p.say_hello();
    let formatted = helpers::format_name(&p.name);
    println!("{}", formatted);
}
