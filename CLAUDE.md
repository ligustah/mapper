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
conversion (`store.GalaxyState(in.State)`) if both types have identical underlying
types, e.g. two separately declared `type X string`. This is narrower than
`types.ConvertibleTo` on purpose: lossy conversions such as `int64` → `int32` and
`string` ↔ `[]byte` are not auto-derived and still need an explicit converter.

### Field Mapping Directives

In method doc comments:
- `// field:TargetField from:"SourceField"` - Map from a different field name
- `// field:TargetField from:"Nested.Path"` - Map from a nested field
- `// field:TargetField from:"."` - Map directly from the input type itself
- `// field:TargetField from:"-"` - Ignore/skip this field
- `// field:TargetField using:"pkg.CustomFunc"` - Use a custom converter function
- `// field:TargetField method:"MethodName"` - Call a no-argument method on the source value
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

Works on plain fields, on `*T` → `*U` (nil stays nil) and on `[]T` → `[]U`.

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
