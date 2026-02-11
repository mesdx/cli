package sample

import (
	"fmt"
	"os"
	"strings"
)

// --- Shadowing: local var shadows package-level ---

// GlobalTimeout is a package-level constant.
const GlobalTimeout = 30

// ShadowExample demonstrates variable shadowing.
func ShadowExample() {
	// local shadows the package-level constant
	GlobalTimeout := 10
	fmt.Println(GlobalTimeout)
}

// --- Multiple definitions with same name (function overloading by receiver) ---

// Animal is a struct with methods.
type Animal struct {
	Species string
	Sound   string
}

// String satisfies fmt.Stringer – common name, many types have it.
func (a *Animal) String() string {
	return fmt.Sprintf("%s says %s", a.Species, a.Sound)
}

// Vehicle is an unrelated struct that also has a String() method.
type Vehicle struct {
	Make  string
	Model string
}

// String satisfies fmt.Stringer on a different type.
func (v *Vehicle) String() string {
	return fmt.Sprintf("%s %s", v.Make, v.Model)
}

// --- Nested / embedded types ---

// Outer contains an inner struct.
type Outer struct {
	Inner struct {
		Value int
	}
	Name string
}

// --- Interface embedding ---

// Reader is a custom interface.
type Reader interface {
	Read(p []byte) (n int, err error)
}

// ReadCloser embeds Reader plus a Close method.
type ReadCloser interface {
	Reader
	Close() error
}

// --- Builtins usage ---

// UseBuiltins demonstrates references to Go builtins that should NOT be
// treated as unresolved external refs.
func UseBuiltins() {
	// builtin functions
	s := make([]int, 10)
	s = append(s, 42)
	_ = len(s)
	_ = cap(s)
	m := make(map[string]int)
	delete(m, "key")
	ch := make(chan int, 1)
	close(ch)
	println("debug")

	// builtin types used as values
	var b bool
	var i int
	var f float64
	var str string
	var e error
	_ = b
	_ = i
	_ = f
	_ = str
	_ = e
}

// --- External / stdlib references ---

// UseExternalStdlib demonstrates references to Go stdlib packages.
func UseExternalStdlib() {
	// os.Getenv is an external stdlib reference
	home := os.Getenv("HOME")
	fmt.Println(home)

	// strings.HasPrefix is external stdlib
	if strings.HasPrefix(home, "/") {
		fmt.Println("absolute path")
	}
}

// --- Type alias vs new type ---

// MyString is a new type (not an alias), wrapping string.
type MyString string

// MyStringAlias is a true alias for string.
type MyStringAlias = string

// --- Iota constants ---

type Color int

const (
	Red Color = iota
	Green
	Blue
)

// --- Function type ---

// TransformFunc is a function type definition.
type TransformFunc func(input string) (output string, err error)

// --- Init function (special, no explicit call) ---

func init() {
	fmt.Println("package initialized")
}
