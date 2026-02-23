package scmsearch

import (
	"fmt"
	"strings"
)

// Stub is a predefined, per-language query template.
type Stub struct {
	ID          string
	Description string
	Args        []string          // required placeholder names
	Templates   map[string]string // language -> SCM query with {{placeholders}}
}

var stubRegistry = map[string]*Stub{}

func init() {
	register(defsFunctionNamed)
	register(defsClassNamed)
	register(defsInterfaceNamed)
	register(refsTypeNamed)
	register(refsCallNamed)
	register(defsMethodNamed)
	register(refsWriteNamed)
	register(refsImportNamed)
}

func register(s *Stub) { stubRegistry[s.ID] = s }

// LookupStub returns a stub by ID, or nil if not found.
func LookupStub(id string) *Stub { return stubRegistry[id] }

// ListStubs returns all registered stubs.
func ListStubs() []*Stub {
	out := make([]*Stub, 0, len(stubRegistry))
	for _, s := range stubRegistry {
		out = append(out, s)
	}
	return out
}

// Render resolves placeholders in the template for the given language.
func (s *Stub) Render(language string, args map[string]string) (string, error) {
	tpl, ok := s.Templates[language]
	if !ok {
		return "", fmt.Errorf("stub %q has no template for language %q", s.ID, language)
	}
	for _, a := range s.Args {
		val, ok := args[a]
		if !ok || val == "" {
			return "", fmt.Errorf("stub %q requires arg %q", s.ID, a)
		}
		tpl = strings.ReplaceAll(tpl, "{{"+a+"}}", val)
	}
	return tpl, nil
}

// ---------------------------------------------------------------------------
// Stubs
// ---------------------------------------------------------------------------

var defsFunctionNamed = &Stub{
	ID:          "defs.function.named",
	Description: "Find function definitions by name.",
	Args:        []string{"name"},
	Templates: map[string]string{
		"go":         `(function_declaration name: (identifier) @def.function (#eq? @def.function "{{name}}"))`,
		"java":       `(method_declaration name: (identifier) @def.function (#eq? @def.function "{{name}}"))`,
		"rust":       `(function_item name: (identifier) @def.function (#eq? @def.function "{{name}}"))`,
		"python":     `(function_definition name: (identifier) @def.function (#eq? @def.function "{{name}}"))`,
		"typescript": `(function_declaration name: (identifier) @def.function (#eq? @def.function "{{name}}"))`,
		"javascript": `(function_declaration name: (identifier) @def.function (#eq? @def.function "{{name}}"))`,
	},
}

var defsClassNamed = &Stub{
	ID:          "defs.class.named",
	Description: "Find class/struct definitions by name.",
	Args:        []string{"name"},
	Templates: map[string]string{
		"go":         `(type_declaration (type_spec name: (type_identifier) @def.class (#eq? @def.class "{{name}}") type: (struct_type)))`,
		"java":       `(class_declaration name: (identifier) @def.class (#eq? @def.class "{{name}}"))`,
		"rust":       `(struct_item name: (type_identifier) @def.class (#eq? @def.class "{{name}}"))`,
		"python":     `(class_definition name: (identifier) @def.class (#eq? @def.class "{{name}}"))`,
		"typescript": `(class_declaration name: (type_identifier) @def.class (#eq? @def.class "{{name}}"))`,
		"javascript": `(class_declaration name: (identifier) @def.class (#eq? @def.class "{{name}}"))`,
	},
}

var defsInterfaceNamed = &Stub{
	ID:          "defs.interface.named",
	Description: "Find interface/trait definitions by name.",
	Args:        []string{"name"},
	Templates: map[string]string{
		"go":         `(type_declaration (type_spec name: (type_identifier) @def.interface (#eq? @def.interface "{{name}}") type: (interface_type)))`,
		"java":       `(interface_declaration name: (identifier) @def.interface (#eq? @def.interface "{{name}}"))`,
		"rust":       `(trait_item name: (type_identifier) @def.interface (#eq? @def.interface "{{name}}"))`,
		"python":     `(class_definition name: (identifier) @def.interface (#eq? @def.interface "{{name}}"))`,
		"typescript": `(interface_declaration name: (type_identifier) @def.interface (#eq? @def.interface "{{name}}"))`,
		"javascript": `(class_declaration name: (identifier) @def.interface (#eq? @def.interface "{{name}}"))`,
	},
}

var refsTypeNamed = &Stub{
	ID:          "refs.type.named",
	Description: "Find type references by name.",
	Args:        []string{"name"},
	Templates: map[string]string{
		"go":         `(type_identifier) @ref.type (#eq? @ref.type "{{name}}")`,
		"java":       `(type_identifier) @ref.type (#eq? @ref.type "{{name}}")`,
		"rust":       `(type_identifier) @ref.type (#eq? @ref.type "{{name}}")`,
		"python":     `(identifier) @ref.type (#eq? @ref.type "{{name}}")`,
		"typescript": `(type_identifier) @ref.type (#eq? @ref.type "{{name}}")`,
		"javascript": `(identifier) @ref.type (#eq? @ref.type "{{name}}")`,
	},
}

var refsCallNamed = &Stub{
	ID:          "refs.call.named",
	Description: "Find call sites of a function by name.",
	Args:        []string{"name"},
	Templates: map[string]string{
		"go":         `(call_expression function: (identifier) @ref.call (#eq? @ref.call "{{name}}"))`,
		"java":       `(method_invocation name: (identifier) @ref.call (#eq? @ref.call "{{name}}"))`,
		"rust":       `(call_expression function: (identifier) @ref.call (#eq? @ref.call "{{name}}"))`,
		"python":     `(call function: (identifier) @ref.call (#eq? @ref.call "{{name}}"))`,
		"typescript": `(call_expression function: (identifier) @ref.call (#eq? @ref.call "{{name}}"))`,
		"javascript": `(call_expression function: (identifier) @ref.call (#eq? @ref.call "{{name}}"))`,
	},
}

var defsMethodNamed = &Stub{
	ID:          "defs.method.named",
	Description: "Find method definitions by name.",
	Args:        []string{"name"},
	Templates: map[string]string{
		"go":         `(method_declaration name: (field_identifier) @def.method (#eq? @def.method "{{name}}"))`,
		"java":       `(method_declaration name: (identifier) @def.method (#eq? @def.method "{{name}}"))`,
		"rust":       `(impl_item body: (declaration_list (function_item name: (identifier) @def.method (#eq? @def.method "{{name}}"))))`,
		"python":     `(class_definition body: (block (function_definition name: (identifier) @def.method (#eq? @def.method "{{name}}"))))`,
		"typescript": `(method_definition name: (property_identifier) @def.method (#eq? @def.method "{{name}}"))`,
		"javascript": `(method_definition name: (property_identifier) @def.method (#eq? @def.method "{{name}}"))`,
	},
}

var refsWriteNamed = &Stub{
	ID:          "refs.write.named",
	Description: "Find write/assignment sites for a variable by name.",
	Args:        []string{"name"},
	Templates: map[string]string{
		"go":         `(assignment_statement left: (expression_list (identifier) @ref.write (#eq? @ref.write "{{name}}")))`,
		"java":       `(assignment_expression left: (identifier) @ref.write (#eq? @ref.write "{{name}}"))`,
		"rust":       `(assignment_expression left: (identifier) @ref.write (#eq? @ref.write "{{name}}"))`,
		"python":     `(assignment left: (identifier) @ref.write (#eq? @ref.write "{{name}}"))`,
		"typescript": `(assignment_expression left: (identifier) @ref.write (#eq? @ref.write "{{name}}"))`,
		"javascript": `(assignment_expression left: (identifier) @ref.write (#eq? @ref.write "{{name}}"))`,
	},
}

var refsImportNamed = &Stub{
	ID:          "refs.import.named",
	Description: "Find import statements referencing a name.",
	Args:        []string{"name"},
	Templates: map[string]string{
		"go":         `(import_spec path: (interpreted_string_literal) @ref.import (#match? @ref.import "{{name}}"))`,
		"java":       `(import_declaration (scoped_identifier name: (identifier) @ref.import (#eq? @ref.import "{{name}}")))`,
		"rust":       `(use_declaration argument: (scoped_identifier name: (identifier) @ref.import (#eq? @ref.import "{{name}}")))`,
		"python":     `(import_from_statement name: (dotted_name (identifier) @ref.import (#eq? @ref.import "{{name}}")))`,
		"typescript": `(import_statement (import_clause (named_imports (import_specifier name: (identifier) @ref.import (#eq? @ref.import "{{name}}")))))`,
		"javascript": `(import_statement (import_clause (named_imports (import_specifier name: (identifier) @ref.import (#eq? @ref.import "{{name}}")))))`,
	},
}
