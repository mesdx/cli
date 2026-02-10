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

;; Impl blocks for methods
(impl_item
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

;; Identifiers as references
(identifier) @ref.identifier

;; Type identifiers
(type_identifier) @ref.type

;; Field identifiers
(field_identifier) @ref.field
