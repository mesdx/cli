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

;; Type definitions
(type_declaration
  (type_spec
    name: (type_identifier) @def.struct
    type: (struct_type)))

(type_declaration
  (type_spec
    name: (type_identifier) @def.interface
    type: (interface_type)))

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

;; Identifiers as references
(identifier) @ref.identifier

(type_identifier) @ref.type

(field_identifier) @ref.field
