;; Go symbol definitions and references

;; Package declaration
(package_clause
  (package_identifier) @def.package)

;; Function definitions
(function_declaration
  name: (identifier) @def.function)

;; Method definitions (simplified - just capture the method name)
(method_declaration
  name: (field_identifier) @def.method)

;; Struct type definitions
(type_declaration
  (type_spec
    name: (type_identifier) @def.struct
    type: (struct_type)))

;; Interface type definitions
(type_declaration
  (type_spec
    name: (type_identifier) @def.interface
    type: (interface_type)))

;; Function type definitions
(type_declaration
  (type_spec
    name: (type_identifier) @def.typealias
    type: (function_type)))

;; Other type definitions (type aliases, new types, etc.)
(type_declaration
  (type_spec
    name: (type_identifier) @def.typealias
    type: (type_identifier)))

;; Slices, maps, channels, etc. as new types
(type_declaration
  (type_spec
    name: (type_identifier) @def.typealias
    type: (slice_type)))

(type_declaration
  (type_spec
    name: (type_identifier) @def.typealias
    type: (map_type)))

(type_declaration
  (type_spec
    name: (type_identifier) @def.typealias
    type: (channel_type)))

(type_declaration
  (type_spec
    name: (type_identifier) @def.typealias
    type: (pointer_type)))

;; Struct fields
(field_declaration
  name: (field_identifier) @def.field)

;; Const and var declarations
(const_declaration
  (const_spec
    name: (identifier) @def.const))

(var_declaration
  (var_spec
    name: (identifier) @def.var))

;; Import declarations (as references)
(import_spec
  path: (interpreted_string_literal) @ref.import)

;; Function/method calls
(call_expression
  function: (identifier) @ref.call)

(call_expression
  function: (selector_expression
    field: (field_identifier) @ref.call))

;; Assignment left-hand side (writes)
(assignment_statement
  left: (expression_list
    (identifier) @ref.write))

;; Identifiers as references
(identifier) @ref.identifier

(type_identifier) @ref.type

(field_identifier) @ref.field
