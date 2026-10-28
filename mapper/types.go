package mapper

import (
	"errors"
	"fmt"
	"go/token"
	"go/types"
	"path"
	"strings"

	"github.com/dave/jennifer/jen"
)

var (
	StringType        = getGoType(types.Typ[types.String])
	IsZeroerInterface = types.NewInterfaceType([]*types.Func{
		types.NewFunc(token.NoPos, types.NewPackage("", ""), "IsZero",
			types.NewSignatureType(nil, nil, nil, nil,
				types.NewTuple(types.NewVar(token.NoPos, types.NewPackage("", ""), "", types.Typ[types.Bool])),
				false),
		),
	}, nil)
)

type GeneratorFunc func(ctx *FunctionContext, input jen.Code) jen.Code

type FunctionContext struct {
	targetType *GoType
	body       *jen.Statement
	canFail    bool
	varCounter int
}

func (f *FunctionContext) add(code ...jen.Code) {
	f.body.Add(code...)
	f.canFail = true
}

func (f *FunctionContext) tmpVar() string {
	name := fmt.Sprintf("tmp%d", f.varCounter)
	f.varCounter++

	return name
}

func (f *FunctionContext) handleError(input jen.Code) *jen.Statement {
	tmp := f.tmpVar()

	f.add(jen.List(jen.Id(tmp), jen.Err()).Op(":=").Add(input).Line())
	f.add(jen.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.List(f.targetType.zero(), jen.Err())),
	).Line().Line())

	return jen.Id(tmp)
}

type Interface struct {
	Name    string
	Mappers []*TypeMapping
}

type TypePair struct {
	from *GoType
	to   *GoType
}

func (t TypePair) String() string {
	return fmt.Sprintf("%s:%s", t.from, t.to)
}

type TypeMapping struct {
	types TypePair

	generator GeneratorFunc
	canFail   bool
}

type ResolvedMapping struct {
	output    *GoField
	input     *jen.Statement
	canFail   bool
	generator GeneratorFunc
	// produces is the type the generator hands back when it is not the output
	// field's own, as a method: or using: result may be
	produces *GoType
}

type GoInterface struct {
	Name    string
	Pkg     string
	Type    *types.Interface
	Methods []*GoFunction
	// skipUnmatched is the interface's own mapper:skip-unmatched, which also holds
	// for the struct conversions derived for it
	skipUnmatched bool
}

type GoFunction struct {
	name     string
	hasError bool
	inputs   []*GoParameter
	output   *GoType
	options  map[string]*FieldOptions
	// skipOptional leaves a builder's optional setters unset when no source field
	// has their name, instead of failing
	skipOptional bool
	// skipUnmatched leaves every target field unset that has neither a directive
	// nor a source field of its name, instead of failing
	skipUnmatched bool
}

type GoFieldPath []*GoField

func (g GoFieldPath) typ() *GoType {
	if len(g) == 0 {
		return nil
	}

	return g[len(g)-1].typ
}

func (g GoFieldPath) id() *jen.Statement {
	if len(g) == 0 {
		return jen.Empty()
	}

	id := jen.Id(g[0].name)

	for _, field := range g[1:] {
		id = id.Dot(field.name)
	}

	return id
}

type GoField struct {
	name string
	typ  *GoType
	raw  *types.Var
}

type GoParameter struct {
	name string
	typ  *GoType
}

type FieldOptions struct {
	from    string
	using   string
	method  string
	mapping string
}

type GoKind int

const (
	InvalidKind GoKind = iota
	BasicKind
	StructKind
	SliceKind
	PointerKind
	MapKind
	ArrayKind
)

type GoType struct {
	Kind GoKind
	Raw  types.Type

	Name     string
	Pkg      string
	TypeName string

	IsProtoEnum bool

	// if Kind is struct
	Fields []*GoField

	// if Kind is SliceKind or PointerKind
	Element *GoType

	// if Kind is MapKind
	Key   *GoType
	Value *GoType

	// if Kind is basic
	BasicInfo types.BasicInfo
}

func (g *GoType) String() string {
	return g.TypeName
}

// findField resolves a dotted path such as `Edges.Set.Name` to the fields along it.
// A nil path means the source has no such field, an error means it is ambiguous.
func (g *GoType) findField(source string) (GoFieldPath, error) {
	return g.find(strings.Split(source, "."))
}

func (g *GoType) find(source []string) (GoFieldPath, error) {
	if len(source) == 0 {
		return nil, nil
	}

	field, err := g.getField(source[0])
	if err != nil || field == nil {
		return nil, err
	}

	fields := GoFieldPath{field}
	if len(source) == 1 {
		return fields, nil
	}

	rest, err := field.typ.find(source[1:])
	if err != nil || rest == nil {
		return nil, err
	}

	return append(fields, rest...), nil
}

// getField finds a field by name. Only when no field matches exactly does it fall
// back to comparing case-insensitively, so `SetId` finds `SetID`. Two fields that
// both match that way, e.g. `Id` next to `ID`, are reported instead of picked from.
func (g *GoType) getField(name string) (*GoField, error) {
	// go one lever deeper, if this is an indirect type
	fields := g.Fields
	if g.Element != nil {
		fields = g.Element.Fields
	}

	for _, field := range fields {
		if field.name == name {
			return field, nil
		}
	}

	var found *GoField
	for _, field := range fields {
		if !strings.EqualFold(field.name, name) {
			continue
		}

		if found != nil {
			return nil, fmt.Errorf("field %q matches both %s and %s of %s, spell it exactly",
				name, found.name, field.name, g)
		}

		found = field
	}

	return found, nil
}

func (g *GoType) zero() *jen.Statement {
	if g.Kind == PointerKind {
		return jen.Nil()
	}

	if g.Kind == BasicKind {
		info := g.BasicInfo

		if info&types.IsInteger != 0 {
			return jen.Lit(0)
		} else if info&types.IsFloat != 0 {
			return jen.Lit(0.0)
		} else if info&types.IsComplex != 0 {
			return jen.Lit(0)
		} else if info&types.IsString != 0 {
			return jen.Lit("")
		} else if info&types.IsBoolean != 0 {
			return jen.Lit(false)
		}
		// fallback for other basic types (shouldn't happen in practice)
		return jen.Lit(0)
	}

	return jen.Qual(g.Pkg, g.Name).Values()
}

func (g *GoType) isZero() *jen.Statement {
	if types.Implements(g.Raw, IsZeroerInterface) {
		return jen.Dot("IsZero").Call()
	}

	return jen.Op("==").Add(g.zero())
}

func (g *GoType) qual() *jen.Statement {
	return jen.Qual(g.Pkg, g.Name)
}

// ref renders the type as it is spelled in Go source, including any pointer, slice,
// array or map wrapping around a named type. qual() only names the named type
// itself and drops that wrapping, so it is correct only where the caller spells the
// indirection out.
func (g *GoType) ref() *jen.Statement {
	// Name carries the name of the nearest named type, which for `*T`, `[]T` and
	// `[N]T` is the element's, so it only spells this type if this type is the
	// defined one. A defined type is spelled by its name whatever its shape.
	if _, defined := g.Raw.(*types.Named); !defined {
		switch g.Kind {
		case PointerKind:
			return jen.Op("*").Add(g.Element.ref())
		case SliceKind:
			return jen.Index().Add(g.Element.ref())
		case ArrayKind:
			length := 0
			if arr, ok := g.Raw.(*types.Array); ok {
				length = int(arr.Len())
			}

			return jen.Index(jen.Lit(length)).Add(g.Element.ref())
		case MapKind:
			return jen.Map(g.Key.ref()).Add(g.Value.ref())
		}
	}

	return jen.Qual(g.Pkg, g.Name)
}

func (g *GoType) pointer() *GoType {
	ptr := types.NewPointer(g.Raw)
	return &GoType{
		Kind:     PointerKind,
		Raw:      ptr,
		Name:     g.Name,
		Pkg:      g.Pkg,
		TypeName: types.TypeString(ptr, nil),
		Element:  g,
	}
}

// arrayLen returns the number of elements of an array type, or -1 for anything
// that is not one. Two arrays only map onto each other if they are the same length,
// the generated loop runs over the source and writes into the target.
func arrayLen(g *GoType) int64 {
	if g == nil || g.Raw == nil {
		return -1
	}

	if arr, ok := g.Raw.Underlying().(*types.Array); ok {
		return arr.Len()
	}

	return -1
}

// assignableTo reports whether a value of `from` can be assigned to `to`.
//
// Types that come out of two different packages.Load calls have distinct identities
// even when they name the same declaration, and go/types compares named types by
// identity, so a `using:` converter loaded on its own never matches the interface's
// own view of a named type. Fall back to comparing the fully qualified names, which
// name the same declaration across loads.
func assignableTo(from, to types.Type) bool {
	if types.AssignableTo(from, to) {
		return true
	}

	return types.TypeString(from, nil) == types.TypeString(to, nil)
}

// qualifyReference expands the short forms of a `using:` reference into the full
// `import/path.Name` form. A bare name refers to pkg itself, a `local.Name` whose
// prefix is one of imports refers to the imported package. Anything else, including
// a path with a slash, is returned unchanged.
func qualifyReference(ref, pkg string, imports map[string]string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.Contains(ref, "/") {
		return ref
	}

	head, rest, found := strings.Cut(ref, ".")
	if !found || strings.HasPrefix(head, "(") {
		return pkg + "." + ref
	}

	if full, ok := imports[head]; ok {
		return full + "." + rest
	}

	return ref
}

func splitPackageAndType(name string) (canonical string, pkg string, typeName, methodName string, pointer bool, err error) {
	trimmed := strings.TrimSpace(name)

	if len(trimmed) == 0 {
		err = errors.New("name cannot be empty")
		return
	}

	canonical = trimmed

	if trimmed[0] == '*' {
		pointer = true
		trimmed = trimmed[1:]
	}

	parent, base := path.Split(trimmed)
	parts := strings.Split(base, ".")

	switch len(parts) {
	// this is a type
	case 2:
		typeName = parts[1]

	// this is a method on a type
	case 3:
		// typeName would be something like (*Type) or (Type)
		typeName = parts[1]
		if len(typeName) <= 3 || typeName[0] != '(' || typeName[len(typeName)-1] != ')' {
			err = errors.New("expected method name to be enclosed in parentheses")
			return
		}

		typeName = typeName[1 : len(typeName)-1]

		if typeName[0] == '*' {
			pointer = true
			typeName = typeName[1:]
		}

		methodName = parts[2]

	default:
		err = errors.New("expected package and type Name separated by period")
	}

	pkg = path.Join(parent, parts[0])

	return
}

var typeCache = make(map[types.Type]*GoType)

func getGoType(typ types.Type) *GoType {
	// check cache to prevent infinite recursion on circular types
	if cached, ok := typeCache[typ]; ok {
		return cached
	}

	underlying := typ

	out := &GoType{
		Raw:      typ,
		TypeName: types.TypeString(typ, nil),
	}

	// add to cache before processing to handle circular references
	typeCache[typ] = out

	if pointer, ok := underlying.(*types.Pointer); ok {
		out.Kind = PointerKind
		out.Element = getGoType(pointer.Elem())
		underlying = pointer.Elem()
	} else if slice, ok := underlying.(*types.Slice); ok {
		out.Kind = SliceKind
		out.Element = getGoType(slice.Elem())
		underlying = slice.Elem()
	} else if arr, ok := underlying.(*types.Array); ok {
		out.Kind = ArrayKind
		out.Element = getGoType(arr.Elem())
		return out
	} else if m, ok := underlying.(*types.Map); ok {
		out.Kind = MapKind
		out.Key = getGoType(m.Key())
		out.Value = getGoType(m.Elem())
		return out
	}

	if named, ok := underlying.(*types.Named); ok {
		out.Name += named.Obj().Name()
		out.Pkg = named.Obj().Pkg().Path()
		underlying = named.Underlying()

		for i := 0; i < named.NumMethods(); i++ {
			method := named.Method(i)

			if method.Name() == "Enum" {
				out.IsProtoEnum = true
				return out
			}
		}
	}

	if out.Element != nil {
		return out
	}

	switch u := underlying.(type) {
	case *types.Struct:
		out.Kind = StructKind
		for i := 0; i < u.NumFields(); i++ {
			field := u.Field(i)

			if !field.Exported() {
				continue
			}

			goField := &GoField{
				name: field.Name(),
				typ:  getGoType(field.Type()),
				raw:  field,
			}

			out.Fields = append(out.Fields, goField)
		}
	case *types.Basic:
		out.Kind = BasicKind
		out.BasicInfo = u.Info()
		// for a named basic type (e.g. `type State string`) the name was already
		// taken from the type name above, don't overwrite it with the underlying one
		if out.Name == "" {
			out.Name = u.Name()
		}
	case *types.Slice:
		// this is a named slice type
		out.Kind = SliceKind
		out.Element = getGoType(u.Elem())
	case *types.Map:
		// this is a named map type
		out.Kind = MapKind
		out.Key = getGoType(u.Key())
		out.Value = getGoType(u.Elem())
	}

	return out
}
