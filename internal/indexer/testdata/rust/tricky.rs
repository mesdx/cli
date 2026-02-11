//! Tricky Rust fixture for deep attribute analysis.

use std::collections::HashMap;
use std::fmt;
use std::io::{self, Read, Write};

// --- Shadowing: variable shadowing is idiomatic in Rust ---

const MAX_RETRIES: u32 = 5;

fn shadow_example() {
    let max_retries = 10; // shadows constant (different case, but same concept)
    let max_retries = max_retries + 1; // re-shadow within same scope
    println!("{}", max_retries);
}

// --- Trait inheritance ---

trait Serializable {
    fn serialize(&self) -> String;
}

trait Printable: fmt::Display {
    fn display(&self) {
        println!("{}", self);
    }
}

// --- Multiple trait impls ---

struct Document {
    title: String,
    content: String,
}

impl Serializable for Document {
    fn serialize(&self) -> String {
        format!("{{\"title\":\"{}\"}}", self.title)
    }
}

impl fmt::Display for Document {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "Document({})", self.title)
    }
}

impl Printable for Document {}

// --- Enum with data ---

enum Shape {
    Circle(f64),
    Rectangle(f64, f64),
    Triangle { base: f64, height: f64 },
}

impl Shape {
    fn area(&self) -> f64 {
        match self {
            Shape::Circle(r) => std::f64::consts::PI * r * r,
            Shape::Rectangle(w, h) => w * h,
            Shape::Triangle { base, height } => 0.5 * base * height,
        }
    }
}

// --- Generic struct + trait bound ---

struct Repository<T: Clone> {
    items: HashMap<u64, T>,
}

impl<T: Clone> Repository<T> {
    fn new() -> Self {
        Repository {
            items: HashMap::new(),
        }
    }

    fn save(&mut self, id: u64, item: T) {
        self.items.insert(id, item);
    }

    fn find(&self, id: u64) -> Option<&T> {
        self.items.get(&id)
    }
}

// --- Builtins / prelude usage (should NOT be unresolved external) ---

fn use_builtins() {
    // Prelude types
    let _opt: Option<i32> = Some(42);
    let _res: Result<i32, String> = Ok(42);
    let _v: Vec<i32> = vec![1, 2, 3];
    let _s: String = String::from("hello");
    let _b: Box<i32> = Box::new(42);

    // Common traits used implicitly
    let _cloned = _v.clone();
    let _debug = format!("{:?}", _v);
    let _display = _s.to_string();

    // Iterator methods
    let _sum: i32 = vec![1, 2, 3].iter().sum();
    let _mapped: Vec<i32> = vec![1, 2, 3].iter().map(|x| x * 2).collect();
    let _filtered: Vec<&i32> = vec![1, 2, 3].iter().filter(|&&x| x > 1).collect();

    // println! is a macro, not a function — but fmt is used
    println!("builtins test: {:?}", _opt);
}

// --- External stdlib references ---

fn use_external_stdlib() -> io::Result<()> {
    // std::collections::HashMap
    let mut map: HashMap<String, i32> = HashMap::new();
    map.insert("key".to_string(), 42);

    // std::io
    let mut buffer = Vec::new();
    let _ = io::stdout().write_all(b"hello");
    let _ = io::stdin().read_to_end(&mut buffer);

    Ok(())
}

// --- Type alias ---

type UserId = u64;
type UserMap = HashMap<UserId, Document>;

// --- Struct with lifetime ---

struct Borrowed<'a> {
    data: &'a str,
}

impl<'a> Borrowed<'a> {
    fn new(data: &'a str) -> Self {
        Borrowed { data }
    }

    fn get(&self) -> &str {
        self.data
    }
}

// --- Closure / function pointer ---

type Transformer = fn(i32) -> i32;

fn apply(f: Transformer, val: i32) -> i32 {
    f(val)
}

fn double(x: i32) -> i32 {
    x * 2
}

// --- Static and const ---

static APP_NAME: &str = "TrickyApp";
const VERSION: u32 = 1;

fn main() {
    shadow_example();

    let doc = Document {
        title: "Hello".to_string(),
        content: "World".to_string(),
    };
    println!("{}", doc.serialize());
    doc.display();

    let circle = Shape::Circle(5.0);
    println!("Area: {}", circle.area());

    let mut repo = Repository::new();
    repo.save(1, doc.title.clone());

    use_builtins();
    let _ = use_external_stdlib();

    let result = apply(double, 21);
    println!("{} v{}: {}", APP_NAME, VERSION, result);
}
