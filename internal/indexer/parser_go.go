package indexer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/mesdx/cli/internal/symbols"
)

// GoParser uses go/parser + go/ast for Go symbol extraction.
type GoParser struct{}

func (p *GoParser) Parse(filename string, src []byte) (*symbols.FileResult, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.AllErrors)
	if err != nil {
		// Even with errors, we may have a partial AST.
		if f == nil {
			return &symbols.FileResult{}, nil
		}
	}

	result := &symbols.FileResult{}
	var containerStack []string

	pushContainer := func(name string) { containerStack = append(containerStack, name) }
	popContainer := func() {
		if len(containerStack) > 0 {
			containerStack = containerStack[:len(containerStack)-1]
		}
	}
	currentContainer := func() string {
		if len(containerStack) > 0 {
			return containerStack[len(containerStack)-1]
		}
		return ""
	}

	// Collect definitions
	ast.Inspect(f, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			kind := symbols.KindFunction
			container := ""
			sig := ""
			if decl.Recv != nil && len(decl.Recv.List) > 0 {
				kind = symbols.KindMethod
				container = receiverTypeName(decl.Recv.List[0].Type)
			}
			if decl.Type != nil {
				sig = funcSignature(fset, decl)
			}
			pos := fset.Position(decl.Name.Pos())
			// EndLine = end of body (or end of signature if no body)
			declEnd := decl.End()
			if decl.Body != nil {
				declEnd = decl.Body.End()
			}
			endPos := fset.Position(declEnd)
			result.Symbols = append(result.Symbols, symbols.Symbol{
				Name:          decl.Name.Name,
				Kind:          kind,
				ContainerName: container,
				Signature:     sig,
				StartLine:     pos.Line,
				StartCol:      pos.Column - 1,
				EndLine:       endPos.Line,
				EndCol:        pos.Column - 1 + len(decl.Name.Name),
			})

		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					kind := symbols.KindTypeAlias
					switch s.Type.(type) {
					case *ast.StructType:
						kind = symbols.KindStruct
					case *ast.InterfaceType:
						kind = symbols.KindInterface
					}
					pos := fset.Position(s.Name.Pos())
					// EndLine = end of the full type (struct body, interface body, etc.)
					endPos := fset.Position(s.End())
					result.Symbols = append(result.Symbols, symbols.Symbol{
						Name:      s.Name.Name,
						Kind:      kind,
						StartLine: pos.Line,
						StartCol:  pos.Column - 1,
						EndLine:   endPos.Line,
						EndCol:    pos.Column - 1 + len(s.Name.Name),
					})

					// If struct, extract fields (keep identifier-only spans)
					if st, ok := s.Type.(*ast.StructType); ok && st.Fields != nil {
						for _, field := range st.Fields.List {
							for _, name := range field.Names {
								fp := fset.Position(name.Pos())
								fe := fset.Position(name.End())
								result.Symbols = append(result.Symbols, symbols.Symbol{
									Name:          name.Name,
									Kind:          symbols.KindField,
									ContainerName: s.Name.Name,
									StartLine:     fp.Line,
									StartCol:      fp.Column - 1,
									EndLine:       fe.Line,
									EndCol:        fe.Column - 1,
								})
							}
						}
					}

					// If interface, extract methods (keep identifier-only spans)
					if it, ok := s.Type.(*ast.InterfaceType); ok && it.Methods != nil {
						for _, method := range it.Methods.List {
							for _, name := range method.Names {
								mp := fset.Position(name.Pos())
								me := fset.Position(name.End())
								result.Symbols = append(result.Symbols, symbols.Symbol{
									Name:          name.Name,
									Kind:          symbols.KindMethod,
									ContainerName: s.Name.Name,
									StartLine:     mp.Line,
									StartCol:      mp.Column - 1,
									EndLine:       me.Line,
									EndCol:        me.Column - 1,
								})
							}
						}
					}

				case *ast.ValueSpec:
					kind := symbols.KindVariable
					if decl.Tok == token.CONST {
						kind = symbols.KindConstant
					}
					// EndLine = end of the full value spec
					specEnd := fset.Position(s.End())
					for _, name := range s.Names {
						if name.Name == "_" {
							continue
						}
						pos := fset.Position(name.Pos())
						result.Symbols = append(result.Symbols, symbols.Symbol{
							Name:      name.Name,
							Kind:      kind,
							StartLine: pos.Line,
							StartCol:  pos.Column - 1,
							EndLine:   specEnd.Line,
							EndCol:    pos.Column - 1 + len(name.Name),
						})
					}
				}
			}
		}
		return true
	})

	// Collect references (identifiers that are not definitions).
	_ = pushContainer
	_ = popContainer
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			pushContainer(node.Name.Name)
		}

		ident, ok := n.(*ast.Ident)
		if !ok || ident.Obj != nil && ident.Obj.Pos() == ident.Pos() {
			// Skip definitions (where Obj.Pos == Ident.Pos) — they are
			// already captured above.
			return true
		}
		if ident.Name == "_" || ident.Name == "nil" || ident.Name == "true" || ident.Name == "false" {
			return true
		}
		// Skip single-letter or very short names to reduce noise.
		if len(ident.Name) <= 1 {
			return true
		}
		// Skip built-in type names.
		if isGoBuiltin(ident.Name) {
			return true
		}

		pos := fset.Position(ident.Pos())
		endPos := fset.Position(ident.End())
		result.Refs = append(result.Refs, symbols.Ref{
			Name:             ident.Name,
			Kind:             symbols.RefOther,
			StartLine:        pos.Line,
			StartCol:         pos.Column - 1,
			EndLine:          endPos.Line,
			EndCol:           endPos.Column - 1,
			ContextContainer: currentContainer(),
		})
		return true
	})

	return result, nil
}

func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	case *ast.IndexExpr:
		return receiverTypeName(t.X)
	}
	return ""
}

func funcSignature(fset *token.FileSet, decl *ast.FuncDecl) string {
	var b strings.Builder
	b.WriteString("func ")
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		b.WriteString("(")
		b.WriteString(receiverTypeName(decl.Recv.List[0].Type))
		b.WriteString(") ")
	}
	b.WriteString(decl.Name.Name)
	b.WriteString("(")
	if decl.Type.Params != nil {
		for i, p := range decl.Type.Params.List {
			if i > 0 {
				b.WriteString(", ")
			}
			for j, name := range p.Names {
				if j > 0 {
					b.WriteString(", ")
				}
				b.WriteString(name.Name)
			}
			if len(p.Names) > 0 {
				b.WriteString(" ")
			}
			b.WriteString(typeString(p.Type))
		}
	}
	b.WriteString(")")
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		b.WriteString(" ")
		if len(decl.Type.Results.List) > 1 {
			b.WriteString("(")
		}
		for i, r := range decl.Type.Results.List {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(typeString(r.Type))
		}
		if len(decl.Type.Results.List) > 1 {
			b.WriteString(")")
		}
	}
	return b.String()
}

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	case *ast.SelectorExpr:
		return typeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeString(t.Elt)
	case *ast.MapType:
		return "map[" + typeString(t.Key) + "]" + typeString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + typeString(t.Elt)
	case *ast.FuncType:
		return "func(...)"
	case *ast.ChanType:
		return "chan " + typeString(t.Value)
	}
	return "any"
}

func isGoBuiltin(name string) bool {
	switch name {
	case "bool", "byte", "complex64", "complex128",
		"error", "float32", "float64",
		"int", "int8", "int16", "int32", "int64",
		"rune", "string",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"any", "comparable",
		"append", "cap", "close", "complex", "copy", "delete",
		"imag", "len", "make", "new", "panic", "print", "println",
		"real", "recover":
		return true
	}
	return false
}
