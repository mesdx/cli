;; JavaScript symbol definitions and references

;; Class declarations
(class_declaration
  name: (identifier) @def.class)

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

;; Import statements
(import_statement
  (import_clause
    (identifier) @ref.import))

(import_statement
  (import_clause
    (named_imports
      (import_specifier
        name: (identifier) @ref.import))))

;; Prototype property access
(member_expression
  object: (identifier)
  property: (property_identifier) @ref.prototype
  (#eq? @ref.prototype "prototype"))

;; Function/method calls
(call_expression
  function: (identifier) @ref.call)

(call_expression
  function: (member_expression
    property: (property_identifier) @ref.call))

;; Assignment left-hand side (writes)
(assignment_expression
  left: (identifier) @ref.write)

(assignment_expression
  left: (member_expression
    property: (property_identifier) @ref.write))

;; Identifiers as references
(identifier) @ref.identifier

;; Property identifiers
(property_identifier) @ref.property
