;; TypeScript symbol definitions and references

;; Class declarations
(class_declaration
  name: (type_identifier) @def.class)

;; Interface declarations
(interface_declaration
  name: (type_identifier) @def.interface)

;; Type alias declarations
(type_alias_declaration
  name: (type_identifier) @def.typealias)

;; Enum declarations
(enum_declaration
  name: (identifier) @def.enum)

;; Function declarations
(function_declaration
  name: (identifier) @def.function)

;; Method definitions
(method_definition
  name: (property_identifier) @def.method)

;; Constructor
(method_definition
  name: (property_identifier) @def.constructor
  (#eq? @def.constructor "constructor"))

;; Arrow functions assigned to variables
(lexical_declaration
  (variable_declarator
    name: (identifier) @def.function
    value: (arrow_function)))

;; Variable declarations
(lexical_declaration
  (variable_declarator
    name: (identifier) @def.var))

(variable_declaration
  (variable_declarator
    name: (identifier) @def.var))

;; Property definitions (fields)
(public_field_definition
  name: (property_identifier) @def.field)

;; Namespace/module declarations
(internal_module
  name: (identifier) @def.module)

;; Import statements
(import_statement
  (import_clause
    (identifier) @ref.import))

(import_statement
  (import_clause
    (named_imports
      (import_specifier
        name: (identifier) @ref.import))))

;; Identifiers as references
(identifier) @ref.identifier

;; Type identifiers
(type_identifier) @ref.type

;; Property identifiers
(property_identifier) @ref.property
