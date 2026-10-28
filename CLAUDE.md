# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`mapper` is a Go code generation tool that automates type mapping between different struct types. It eliminates boilerplate conversion code by generating type-safe mapping functions from annotated Go interfaces.

Key capabilities:
- Generates mapping implementations from `// mapper:generate` annotated interfaces
- Discovers indirect mapping routes via a type graph (e.g., TypeA → TypeB → TypeC)
- Supports custom mapper functions registered via `mapper.RegisterCustomMapper()`
- Handles pointer/slice conversions, field renaming, Protocol Buffer enums

## Building and Testing

Build the generator:
```bash
go build -o mapper .
```

Run code generation:
```bash
./mapper [flags] <package_path>
# Example: ./mapper ./example
# Example: ./mapper -strict-enums ./example
```

Flags:
- `-strict-enums`: fail instead of warning when a generated conversion between two enums can produce a value the target does not declare

Run tests:
```bash
go test ./mapper
```

The main entry point is `main.go`, which parses a package and generates mapper implementations in a `gen` subdirectory.

## Architecture

### Core Components

**Parser (`mapper/parser.go`)**
- Uses `golang.org/x/tools/go/packages` to load and parse Go packages
- Extracts interfaces annotated with `// mapper:generate`
- Parses method doc comments for field mapping directives (`// field:Name from:"Source"` or `// field:Name using:"CustomFunc"`)
- Builds `GoInterface` and `GoFunction` representations with type information

**Generator (`mapper/generator.go`)**
- Creates implementation structs for mapping interfaces (e.g., `MyMapperImpl`)
- Uses `github.com/dave/jennifer` for code generation
- Maintains a type graph (`TypeGraph`) to find indirect conversion paths
- Composes multi-step conversions when direct mappings don't exist
- Generates methods that handle pointer conversions, slice mappings, and error handling

**Type System (`mapper/types.go`)**
- Wraps Go's `types.Type` into `GoType` with kind classification (Basic, Struct, Slice, Pointer)
- Provides utilities for type comparison, zero values, and type graph navigation
- Supports field path resolution (e.g., `Payload.ContentType`)

**Type Graph (`mapper/graph.go`)**
- Maintains a directed graph of available type conversions
- Uses `github.com/yourbasic/graph` for shortest-path finding
- Enables automatic chaining of converters to map between indirectly-related types

**Built-in Mappers (`mapper/builtin.go`)**
- Registers default type conversions
- Handles `volatiletech/null` types for nullable database fields

**Protocol Buffer Support (`mapper/proto.go`)**
- Auto-discovers and registers converters for protobuf enum types
- Detects enums by the presence of an `Enum()` method

**Null Type Handling (`mapper/null.go`)**
- Provides converters for `null.String`, `null.Int`, etc. from `volatiletech/null`

**Enum Compatibility (`mapper/enum.go`)**
- Collects the constants declared for a named type from its package scope
- Compares both sides of a generated type conversion, reporting source constants the target does not declare and constants whose name matches a target constant with a different value
- Parses `map:` directives and resolves each side to a constant by name or by value
- Only warns by default, `-strict-enums` / `Generator.SetStrictEnums()` makes it an error

### Code Generation Flow

1. **Parse**: `Parser.Parse()` loads package and extracts `// mapper:generate` interfaces
2. **Build Mappings**: For each method, `Generator.buildMapping()` resolves how to map each output field
3. **Find Converters**: `Generator.buildGenerator()` queries the type graph for conversion paths
4. **Compose**: If indirect, `compose()` chains multiple converters together
5. **Render**: `Generator.generateMethod()` emits the final method implementation using `jennifer`

### Direct Type Conversions

When no converter is registered for a pair, the generator falls back to a plain Go
conversion (`store.GalaxyState(in.State)`) in three cases: both types have identical
underlying types, e.g. two separately declared `type X string`; both are integers of
any width, named or not, matching the builtin converters between unnamed integers;
or the pair is `string` ↔ `[]byte`. This is narrower than `types.ConvertibleTo` on
purpose: `int` ↔ `float64` or `string` ↔ `int` still need an explicit converter.
`time.Duration` takes no part in any of these: a plain conversion would carry
nanoseconds into an integer that usually counts seconds or milliseconds.

### Derived Struct Conversions

Two named structs that nothing else converts are mapped field by field through a
helper method on the implementation, `func (m XImpl) mapGoalToGoalProto(in Goal) GoalProto`,
generated once per interface and then used like any converter: for a nested field,
`*T` → `*U`, `[]T` → `[]*U`, map values and so on. The helper follows the same field
rules as a declared method and the interface's own toggles, it takes no directives;
a pair whose fields do not line up still needs a declared method. It returns an
error when one of its fields can fail. A struct that refers to itself, a tree, calls
its own helper. Derivation only fills in where the generator would otherwise have
failed: a declared method, a registered converter or a direct conversion always
wins, and the fallback that converts a whole input into a field without a source
never takes a struct apart, only an explicit `from:"."` does.

Converters coming from an interface's own methods and its derived helpers are
scoped to that interface, `GenerateInterface` takes them back out of the graph when
it is done, so a second interface never calls a method of the first one's
implementation.

### Nil Pointers

A generated method whose only input is a pointer opens with a nil guard: nil in
gives nil out for a pointer result, the zero value for a value result. A `*T` → `*U`
field is derived from a `T` → `U` converter by wrapping it in a nil check, so a nil
field stays nil. A converter registered for `*T` itself is called directly and is
expected to handle nil on its own. A converter from `T` that already returns `*U`,
like `timestamppb.New`, has its pointer kept rather than copied and re-addressed.
A `*T` → `U` field mapped through `method:` or `using:` keeps the zero value of `U`
when the pointer is nil, the method or function never runs on a zero `T`.

### Builder Targets

A method whose output type is also one of its inputs fills a builder instead of
returning a struct, the shape ent's `Create` and `UpdateOne` builders have:

```go
// field:ID from:"-"
CommitRow(in *proto.Commit, b *ent.CommitRowCreate) *ent.CommitRowCreate
```

Every method `SetX(v T)` on the builder that returns the builder is a target field
named `X` of type `T`, resolved with the same rules and directives as a struct field,
and rendered as one setter call each. The builder comes back for chaining, so the
caller adds what the source does not carry: `m.CommitRow(c, tx.CommitRow.Create()).SetRef(ref).Exec(ctx)`.
A nil source leaves the builder untouched. Detection is generic, two ent conventions
are layered on top: `SetNillableX(*T)` is folded into `X` and used instead of `SetX`
when the source field or the converter hands back a pointer, so nil never becomes
a zero value; and for an edge with both `SetX(*Entity)` and `SetXID(id)` the id form
is the field, unless a directive addresses `X` and not `XID`. Variadic adders like
`AddFileIDs(...)` are not setters. Everything else stays strict, a setter without a
source has to be ignored with `from:"-"`. The one opt-out is `// mapper:skip-optional`,
on a method or on the whole interface: an optional setter, one with a
`SetNillableX` twin, that has neither a directive nor a source field of its name is
then left unset. Required setters still fail.

### Unmatched Fields

By default every target field needs a source or a directive. `// mapper:skip-unmatched`,
on a method or on the whole interface, leaves any field or builder setter without
either unset instead, struct targets and builders alike. A field whose source exists
but cannot be converted, and a directive pointing at a missing source, still fail.
Toggles are parsed by `applyToggle` in `parser.go`, a new one goes there.

`// mapper:inline Report Host`, on a method or the interface, flattens sub-structs:
a target field that no input has a field of its name for is read from
`Report.<Name>`, then `Host.<Name>`, as if a `from:` said so, with the usual nil
checks for a pointer on the way. A directive on the field still wins. Every field
lookup goes through `Generator.sourceName`, which applies both.

### Nested Paths

A `from:"Edges.Set.Name"` path that crosses a pointer never dereferences nil. When
the method returns an error and the type holding the pointer declares
`SetOrErr() (*Set, error)`, the accessor ent generates for every edge, the pointer
is read through it and an edge that was not loaded comes back as that error.
Otherwise the pointer is nil-checked and the target field keeps its zero value.
The leaf itself may be a pointer, that is left to the converter as usual.

### Field Matching

An output field is matched to an input field of the same name. Only when there is no
exact match does the lookup fall back to a case-insensitive one, so protoc's `SetId`
finds ent's `SetID` without a directive. Two candidates that both match that way
(`Id` next to `ID`) are an error rather than a silent pick, spell the name exactly
with `from:` in that case.

### Field Mapping Directives

In method doc comments:
- `// field:TargetField from:"SourceField"` - Map from a different field name
- `// field:TargetField from:"Nested.Path"` - Map from a nested field
- `// field:TargetField from:"."` - Map directly from the input type itself
- `// field:TargetField from:"-"` - Ignore/skip this field
- `// field:TargetField using:"import/path.CustomFunc"` - Use a custom converter function
- `// field:TargetField using:"CustomFunc"` - A bare name is a function of the interface's own package
- `// field:TargetField using:"alias.CustomFunc"` - An import of the interface's file, by its local name
- `// field:TargetField method:"MethodName"` - Call a no-argument method on the source value,
  which may return `(T, error)`; with `from:"."` it is called on the input itself

A `using:` function or `method:` of the element type is applied to every element when
the source and target are both slices (or, for `using:`, both maps with assignable
keys) and the collection itself does not fit: `[]int` → `[]string` with
`using:"Hex"` for a `func Hex(int) string`.
- `// field:TargetField map:"a=X,b=Y"` - Map listed values explicitly (see below)

### Enum Mapping

Two named types that declare the same values need no annotation, the generator emits
a direct conversion. When the values differ, list them with `map:`:

```go
// field:Visibility map:"public=VisibilityProtoPublic,private=VisibilityProtoPrivate"
MapEnumExample(src *models.EnumExample) (*protos.EnumExampleProto, error)
```

Either side of a pair is spelled as the name of a constant or as its value, and the
generated code refers to the constant wherever one exists, so it keeps following that
constant if its value changes later. The two sides do not need the same underlying
type, which is how a string enum maps onto a numbered one.

The directive generates a switch. Go enums are not closed types, so a value outside
the declared constants is always representable and the switch keeps a `default`
branch returning an error — a mapped field therefore requires the method to return
`error`. Source constants missing from the directive are reported the same way as
other enum problems: a warning, or an error under `-strict-enums`.

A `*=X` entry makes the default branch yield `X` instead, the field can then no
longer fail and unlisted constants are not reported. With `*` as the only entry the
pairs come from matching constant names, as below. Literals may be strings,
integers or `true`/`false`, so an enum maps onto a plain string or bool too:
`map:"KindFile=file,KindDir=dir,*=unknown"`.

Works on plain fields, on `*T` → `*U` (nil stays nil) and on `[]T` → `[]U`.

Two enums whose underlying types differ need no `map:` either when their constants
pair up by name. Case and underscores are ignored and the type name is stripped, so
ent's `set.StateActive = "active"` and protoc's `SetState_SET_STATE_ACTIVE` are both
`ACTIVE`. A string constant also matches by its value. A constant still without a
counterpart is tried once more with the underscore-terminated prefix all constants
of its type share cut off, which pairs `SetState_SET_ACTIVE` with `StateActive`;
this second pass never undoes a pair the first one found. The generator then emits the
same switch a `map:` directive would, error branch included. Source constants with
no counterpart, or with more than one, are reported like constants a `map:` leaves
out, and a directive still takes over wherever the names do not line up.

## Custom Mappers

To extend with custom type conversions:

1. Define a function with signature `func(SourceType) TargetType` or `func(SourceType) (TargetType, error)`
2. Register it in an `init()` function: `mapper.RegisterCustomMapper(YourFunction)`
3. Reference it in field directives: `// field:MyField using:"yourpackage.YourFunction"`

See `mapper/builtin.go` for examples of custom mapper registration.

## Example Structure

The `example/` directory demonstrates usage:
- `example/mapper.go`: Interface with `// mapper:generate` and field directives
- `example/gen/`: Generated implementation code (created by running `./mapper ./example`)
- `example/models/`: Source structs (database models)
- `example/protos/`: Target structs (protobuf messages)

## Determinism

Regenerating an unchanged package produces a byte-identical file. Interfaces are
discovered by walking `pkg.TypesInfo.Defs`, which is a map, so `Parser.Parse()`
sorts them by package path and name before returning; methods within an interface
already come back sorted by `go/types`. `TestParseOrderIsStable` guards this — a
change that reintroduces map iteration order makes the generated file reshuffle
between runs.

## Important Notes

- Generated implementations are in a `gen` subdirectory by default (configurable via `NewGenerator()`)
- The generator logs conversion chains for debugging (e.g., "converting X to Y using 2 converters")
- Methods requiring error handling must have `error` as second return value
- Type graph enables powerful composition: define small mappers, tool chains them automatically
