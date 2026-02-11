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

;; Identifiers as references
(identifier) @ref.identifier

;; Attribute references
(attribute
  attribute: (identifier) @ref.attribute)
