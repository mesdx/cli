;; Python symbol definitions and references

;; Class definitions
(class_definition
  name: (identifier) @def.class)

;; Method definitions (functions inside class)
(class_definition
  body: (block
    (function_definition
      name: (identifier) @def.method)))

;; Property decorators
(class_definition
  body: (block
    (decorated_definition
      (decorator
        (identifier) @decorator.property (#eq? @decorator.property "property"))
      definition: (function_definition
        name: (identifier) @def.property))))

;; Regular decorated methods
(class_definition
  body: (block
    (decorated_definition
      definition: (function_definition
        name: (identifier) @def.method))))

;; Top-level function definitions
(function_definition
  name: (identifier) @def.function)

;; Parameters
(parameters
  (identifier) @def.parameter)

;; Assignment statements (variables and constants)
(assignment
  left: (identifier) @def.var)

(assignment
  left: (pattern_list
    (identifier) @def.var))

;; Augmented assignments
(augmented_assignment
  left: (identifier) @def.var)

;; Import statements
(import_statement
  name: (dotted_name
    (identifier) @ref.import))

(import_from_statement
  name: (dotted_name
    (identifier) @ref.import))

(aliased_import
  name: (dotted_name
    (identifier) @ref.import))

;; Decorator references (annotations)
(decorator
  (identifier) @ref.annotation)

;; Function/method calls
(call
  function: (identifier) @ref.call)

(call
  function: (attribute
    attribute: (identifier) @ref.call))

;; Assignment left-hand side (writes)
(assignment
  left: (identifier) @ref.write)

(assignment
  left: (pattern_list
    (identifier) @ref.write))

;; Identifiers as references
(identifier) @ref.identifier

;; Attribute references
(attribute
  attribute: (identifier) @ref.attribute)

;; Quoted forward references in type annotations.
;; Python allows strings as type annotations to break circular import cycles:
;;   def foo() -> "MyClass": ...
;;   def bar(x: "MyClass"): ...
;;   def baz() -> tuple["MyClass", bool]: ...   (string inside generic)
;; These string nodes are NOT identifiers, so (identifier) @ref.identifier
;; misses them. We capture the enclosing (type) node and walk its subtree
;; in the extractor to find all string literals inside.
(type) @ref.annotation
