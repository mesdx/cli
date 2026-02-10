;; Go symbol definitions and references

;; Package declaration
(package_clause
  (package_identifier) @def.package)

;; Function definitions
(function_declaration
  name: (identifier) @def.function)

;; Method definitions
(method_declaration
  name: (field_identifier) @def.method
  receiver: (parameter_list
    (parameter_declaration
      type: [
        (type_identifier) @container.name
        (pointer_type (type_identifier) @container.name)
      ])))

;; Type definitions
(type_declaration
  (type_spec
    name: (type_identifier) @def.type
    type: (struct_type) @def.struct))

(type_declaration
  (type_spec
    name: (type_identifier) @def.type
    type: (interface_type) @def.interface))

(type_declaration
  (type_spec
    name: (type_identifier) @def.typealias))

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
