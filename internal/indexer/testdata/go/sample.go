package sample

import "fmt"

// MaxRetries is a package-level constant.
const MaxRetries = 3

// DefaultName is a package-level variable.
var DefaultName = "world"

// Greeter is an interface for greeting.
type Greeter interface {
	Greet(name string) string
}

// Person represents a person with a name and age.
type Person struct {
	Name string
	Age  int
}

// NewPerson creates a new Person.
func NewPerson(name string, age int) *Person {
	return &Person{Name: name, Age: age}
}

// Greet implements the Greeter interface.
func (p *Person) Greet(name string) string {
	return fmt.Sprintf("Hello, %s! I'm %s.", name, p.Name)
}

// SayHello is a standalone function.
func SayHello() {
	p := NewPerson(DefaultName, 30)
	msg := p.Greet("friend")
	fmt.Println(msg)
	for i := 0; i < MaxRetries; i++ {
		fmt.Println(p.Greet(DefaultName))
	}
}
