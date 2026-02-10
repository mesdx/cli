;; Python symbol definitions and references

;; Class definitions
(class_definition
  name: (identifier) @def.class)

;; Function definitions
(function_definition
  name: (identifier) @def.function)

;; Decorated functions (including @property)
(decorated_definition
  (decorator
    (identifier) @decorator.name)
  definition: (function_definition
    name: (identifier) @def.method))

;; Parameters
(parameters
  (identifier) @def.parameter)

;; Assignment statements (variables)
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
