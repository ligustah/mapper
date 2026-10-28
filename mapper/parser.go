package mapper

import (
	"cmp"
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/token"
	"go/types"
	"log"
	"path"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	// skipOptionalDirective, on an interface or one of its methods, leaves a builder's
	// optional setters unset when no source field has their name.
	skipOptionalDirective = "mapper:skip-optional"
	// skipUnmatchedDirective, on an interface or one of its methods, leaves every
	// target field or setter unset that has neither a directive nor a source field
	// of its name.
	skipUnmatchedDirective = "mapper:skip-unmatched"
	// inlineDirective, followed by the names of sub-structs such as `Report`, looks
	// a target field up as `Report.<Name>` when no input has a field of its name.
	inlineDirective = "mapper:inline"
)

// applyToggle sets the toggle a doc comment line names, and reports whether it
// named one.
func applyToggle(f *GoFunction, line string) bool {
	if rest, ok := strings.CutPrefix(line, inlineDirective+" "); ok {
		f.inline = append(f.inline, strings.Fields(rest)...)
		return true
	}

	switch line {
	case skipOptionalDirective:
		f.skipOptional = true
	case skipUnmatchedDirective:
		f.skipUnmatched = true
	default:
		return false
	}

	return true
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func MustLookupType(name string) *GoType {
	typ, err := LookupType(name)

	if err != nil {
		panic(fmt.Errorf("failed looking up type: %w", err))
	}

	return typ
}

func LookupType(name string) (*GoType, error) {
	_, pkg, typeName, _, pointer, err := splitPackageAndType(name)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}, pkg)

	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected to find exactly one package '%s', found %d", pkg, len(pkgs))
	}

	obj := pkgs[0].Types.Scope().Lookup(typeName)

	if obj == nil {
		return nil, fmt.Errorf("symbol '%s' not found in package '%s'", typeName, pkg)
	}

	goType := getGoType(obj.Type())

	if pointer {
		return goType.pointer(), nil
	}

	return goType, nil
}

func (p *Parser) Parse(dir string, patterns ...string) ([]*GoInterface, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  dir,
	}, patterns...)

	if err != nil {
		return nil, err
	}

	var interfaces []*GoInterface
	for _, pkg := range pkgs {
		interfaces = append(interfaces, p.parsePackage(pkg)...)
	}

	// interfaces are discovered by walking a map, sort them so that generation, and
	// with it the generated file, does not depend on iteration order
	slices.SortStableFunc(interfaces, func(a, b *GoInterface) int {
		if c := cmp.Compare(a.Pkg, b.Pkg); c != 0 {
			return c
		}

		return cmp.Compare(a.Name, b.Name)
	})

	return interfaces, nil
}

type interfaceDoc struct {
	name  string
	doc   string
	funcs map[string]funcDoc
	// imports of the file declaring the interface, local name to package path
	imports map[string]string
}

type funcDoc struct {
	name string
	doc  []string
}

func (p *Parser) parsePackage(pkg *packages.Package) []*GoInterface {
	defs := pkg.TypesInfo.Defs

	// extract doc comments for all interface types
	docs, err := p.getDocComments(pkg)
	if err != nil {
		log.Fatal(err)
	}

	var interfaces []*GoInterface
	for _, object := range defs {
		if object == nil {
			continue
		}

		switch typ := object.Type().(type) {
		case *types.Named:
			if iface, ok := typ.Underlying().(*types.Interface); ok {
				interfaceDocs := docs[typ.Obj().Name()]

				hasGenerateDirective := false
				var toggles []string
				for _, line := range strings.Split(interfaceDocs.doc, "\n") {
					switch line = strings.TrimSpace(line); line {
					case "mapper:generate":
						hasGenerateDirective = true
					default:
						toggles = append(toggles, line)
					}
				}

				if hasGenerateDirective {
					goIface := &GoInterface{
						Name:          typ.Obj().Name(),
						Pkg:           typ.Obj().Pkg().Path(),
						Type:          iface,
						skipUnmatched: slices.Contains(toggles, skipUnmatchedDirective),
					}

					for i := 0; i < iface.NumMethods(); i++ {
						method := iface.Method(i)
						goFunc, err := p.getGoFunction(method, interfaceDocs.funcs[method.Name()], goIface.Pkg, interfaceDocs.imports)
						if err != nil {
							log.Fatal(err)
						}
						for _, toggle := range toggles {
							applyToggle(goFunc, toggle)
						}

						goIface.Methods = append(goIface.Methods, goFunc)
					}

					interfaces = append(interfaces, goIface)
				}
			}
		}
	}

	return interfaces
}

// getGoFunction builds the function for an interface method. pkg and imports are the
// declaring package and its file's imports, which qualify short `using:` references.
func (p *Parser) getGoFunction(method *types.Func, f funcDoc, pkg string, imports map[string]string) (*GoFunction, error) {
	// method type is always *types.Signature
	signature := method.Type().(*types.Signature)
	goFunc := &GoFunction{
		name:    method.Name(),
		options: map[string]*FieldOptions{},
	}

	params := signature.Params()
	for p := 0; p < params.Len(); p++ {
		param := params.At(p)
		goFunc.inputs = append(goFunc.inputs, &GoParameter{
			name: param.Name(),
			typ:  getGoType(param.Type()),
		})
	}

	results := signature.Results()
	if results.Len() != 1 && results.Len() != 2 {
		return nil, fmt.Errorf("only support one or two return values, got %d", results.Len())
	}

	goFunc.output = getGoType(results.At(0).Type())

	if results.Len() > 1 {
		lastReturnValue := results.At(1)

		if named, ok := lastReturnValue.Type().(*types.Named); ok && named.Obj().Name() == "error" {
			goFunc.hasError = true
		} else {
			return nil, fmt.Errorf("only support second return value of type error")
		}
	}

	// parse all comments for options
	for _, s := range f.doc {
		trimmed := strings.TrimSpace(s)

		if toggle, ok := strings.CutPrefix(trimmed, "// "); ok && applyToggle(goFunc, strings.TrimSpace(toggle)) {
			continue
		}

		const fieldPrefix = "// field:"
		if strings.HasPrefix(trimmed, fieldPrefix) {
			parts := strings.SplitN(strings.TrimPrefix(trimmed, fieldPrefix), " ", 2)

			if len(parts) > 1 {
				fieldName := parts[0]
				tag := reflect.StructTag(parts[1])

				if _, ok := goFunc.options[fieldName]; !ok {
					goFunc.options[fieldName] = &FieldOptions{}
				}

				goFunc.options[fieldName].from = tag.Get("from")
				goFunc.options[fieldName].using = qualifyReference(tag.Get("using"), pkg, imports)
				goFunc.options[fieldName].method = tag.Get("method")
				goFunc.options[fieldName].mapping = tag.Get("map")
				goFunc.options[fieldName].fallback = tag.Get("default")

				switch zero, ok := tag.Lookup("zero"); {
				case !ok:
				case zero == "nil":
					goFunc.options[fieldName].zeroNil = true
				default:
					return nil, fmt.Errorf("field %s: zero:%q, the only form is zero:\"nil\"", fieldName, zero)
				}
			}
		}
	}

	if len(goFunc.inputs) > 1 {
		for _, input := range goFunc.inputs {
			if input.name == "" {
				return nil, errors.New("input parameters need to be named, when there is more than one")
			}
		}
	}

	return goFunc, nil
}

func (p *Parser) getDocComments(pkg *packages.Package) (map[string]interfaceDoc, error) {
	docPkg, err := doc.NewFromFiles(pkg.Fset, pkg.Syntax, pkg.PkgPath)
	if err != nil {
		return nil, err
	}

	interfaceDocs := make(map[string]interfaceDoc)

	for _, t := range docPkg.Types {
		if t.Decl.Tok == token.TYPE {
			if len(t.Decl.Specs) == 1 {
				spec := t.Decl.Specs[0]

				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if iface, ok := typeSpec.Type.(*ast.InterfaceType); ok {
						ifaceDoc := interfaceDoc{
							name:    t.Name,
							doc:     t.Doc,
							funcs:   map[string]funcDoc{},
							imports: fileImports(pkg, t.Decl.Pos()),
						}

						for _, method := range iface.Methods.List {
							if len(method.Names) != 1 {
								continue
							}

							fDoc := funcDoc{
								name: method.Names[0].Name,
							}

							if method.Doc != nil {
								for _, comment := range method.Doc.List {
									fDoc.doc = append(fDoc.doc, comment.Text)
								}
							}

							ifaceDoc.funcs[fDoc.name] = fDoc
						}

						interfaceDocs[ifaceDoc.name] = ifaceDoc
					}
				}
			}
		}
	}

	return interfaceDocs, nil
}

// fileImports maps the local names the file containing pos imports to the package
// paths behind them. Blank and dot imports are left out.
func fileImports(pkg *packages.Package, pos token.Pos) map[string]string {
	file := pkg.Fset.File(pos)
	if file == nil {
		return nil
	}

	imports := map[string]string{}
	for _, syntax := range pkg.Syntax {
		if pkg.Fset.File(syntax.Pos()) != file {
			continue
		}

		for _, spec := range syntax.Imports {
			pkgPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				continue
			}

			name := ""
			if spec.Name != nil {
				name = spec.Name.Name
			} else if imported, ok := pkg.Imports[pkgPath]; ok {
				name = imported.Name
			} else {
				name = path.Base(pkgPath)
			}

			if name == "_" || name == "." {
				continue
			}

			imports[name] = pkgPath
		}
	}

	return imports
}
