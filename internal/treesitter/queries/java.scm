;; Java symbol definitions and references

;; Class declarations
(class_declaration
  name: (identifier) @def.class)

;; Interface declarations
(interface_declaration
  name: (identifier) @def.interface)

;; Enum declarations
(enum_declaration
  name: (identifier) @def.enum)

;; Method declarations
(method_declaration
  name: (identifier) @def.method)

;; Constructor declarations
(constructor_declaration
  name: (identifier) @def.constructor)

;; Field declarations
(field_declaration
  declarator: (variable_declarator
    name: (identifier) @def.field))

;; Local variable declarations
(local_variable_declaration
  declarator: (variable_declarator
    name: (identifier) @def.var))

;; Import declarations
(import_declaration
  (scoped_identifier
    name: (identifier) @ref.import))

(import_declaration
  (identifier) @ref.import)

;; Identifiers as references
(identifier) @ref.identifier

;; Type identifiers
(type_identifier) @ref.type
