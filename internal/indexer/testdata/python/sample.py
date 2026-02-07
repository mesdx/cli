from typing import List, Optional

MAX_RETRIES = 3
default_name = "world"

class Greeter:
    def greet(self, name: str) -> str:
        raise NotImplementedError

class Person(Greeter):
    def __init__(self, name: str, age: int):
        self.name = name
        self.age = age

    def greet(self, name: str) -> str:
        return f"Hello, {name}! I'm {self.name}."

    @property
    def display_name(self) -> str:
        return self.name.upper()

def say_hello():
    p = Person(default_name, 30)
    msg = p.greet("friend")
    print(msg)
    for i in range(MAX_RETRIES):
        print(p.greet(default_name))
    print(p.display_name)

def format_name(name: str) -> str:
    return name.strip().title()

if __name__ == "__main__":
    say_hello()
    result = format_name(default_name)
    print(result)
