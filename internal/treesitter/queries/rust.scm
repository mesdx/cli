;; Rust symbol definitions and references

;; Function definitions
(function_item
  name: (identifier) @def.function)

;; Struct definitions
(struct_item
  name: (type_identifier) @def.struct)

;; Enum definitions
(enum_item
  name: (type_identifier) @def.enum)

;; Trait definitions
(trait_item
  name: (type_identifier) @def.trait)

;; Impl blocks for methods (simple type)
(impl_item
  type: (type_identifier) @container.name
  body: (declaration_list
    (function_item
      name: (identifier) @def.method)))

;; Impl blocks for methods (generic type, e.g. impl<T> Foo<T>)
(impl_item
  type: (generic_type
    type: (type_identifier) @container.name)
  body: (declaration_list
    (function_item
      name: (identifier) @def.method)))

;; Trait impl for methods (impl Trait for Type)
(impl_item
  trait: (_)
  type: (type_identifier) @container.name
  body: (declaration_list
    (function_item
      name: (identifier) @def.method)))

;; Type aliases
(type_item
  name: (type_identifier) @def.typealias)

;; Const items
(const_item
  name: (identifier) @def.const)

;; Static items
(static_item
  name: (identifier) @def.const)

;; Module definitions
(mod_item
  name: (identifier) @def.module)

;; Let bindings (variables)
(let_declaration
  pattern: (identifier) @def.var)

;; Use declarations (imports)
(use_declaration
  argument: (identifier) @ref.import)

(use_declaration
  argument: (scoped_identifier
    name: (identifier) @ref.import))

;; Use list items (e.g. use std::io::{self, Read, Write})
(use_declaration
  argument: (use_wildcard) @ref.import)

(use_list
  (identifier) @ref.import)

(use_list
  (scoped_identifier
    name: (identifier) @ref.import))

;; Function/method calls
(call_expression
  function: (identifier) @ref.call)

(call_expression
  function: (field_expression
    field: (field_identifier) @ref.call))

(call_expression
  function: (scoped_identifier
    name: (identifier) @ref.call))

;; Assignment left-hand side (writes)
(assignment_expression
  left: (identifier) @ref.write)

;; Identifiers as references
(identifier) @ref.identifier

;; Type identifiers
(type_identifier) @ref.type

;; Field identifiers
(field_identifier) @ref.field
