# Go Type Mapper

The `Go Type Mapper` is a code generation tool that automates the creation of type mapping logic between different Go structs. It eliminates the need for writing repetitive boilerplate code when converting data structures, especially useful when dealing with DTOs, database models, or Protocol Buffer messages.

By defining simple Go interfaces with special comments, you can instruct the `mapper` tool to generate efficient and type-safe conversion functions. A key feature of the `mapper` is its ability to discover indirect mapping routes between types using a sophisticated type graph, allowing for flexible and powerful transformations even when direct conversions are not immediately apparent.

## Installation

To install the `mapper` tool, you can use `go install`:

```bash
go install github.com/ligustah/mapper@latest
```

Make sure your `GOPATH/bin` directory is in your system's `PATH` to easily run the `mapper` executable.

## Quick Start Example

Let's say you have two structs, `UserSource` and `UserTarget`, and you want to map fields from `UserSource` to `UserTarget`.

**1. Define your source and target structs:**

```go
// models/source.go
package models

type UserSource struct {
	ID    string
	Name  string
	Email string
}

// models/target.go
package models

type UserTarget struct {
	UserID string
	FullName string
	EmailAddress string
}
```

**2. Create a mapping interface:**

Create a new Go file (e.g., `mapper/user_mapper.go`) and define an interface with the `// mapper:generate` annotation. Each method in this interface will represent a mapping function.

```go
// mapper/user_mapper.go
package mapper

import "your_module/models"

//go:generate mapper .
// mapper:generate
type UserMapper interface {
	MapUser(source *models.UserSource) *models.UserTarget
}
```

**3. Run the `mapper` tool:**

Navigate to your project's root directory in your terminal and run:

```bash
mapper ./mapper
```

This command will parse the `mapper` directory, find your `UserMapper` interface, and generate a new file (e.g., `mapper/gen/user_mapper_gen.go`) containing the implementation of `UserMapper`.

**4. Use the generated mapper:**

You can now use the generated `UserMapperImpl` in your application:

```go
package main

import (
	"fmt"
	"your_module/mapper"
	"your_module/models"
)

func main() {
	userSource := &models.UserSource{
		ID    : "123",
		Name  : "John Doe",
		Email : "john.doe@example.com",
	}

	userMapper := &mapper.UserMapperImpl{} // Use the generated implementation
	userTarget := userMapper.MapUser(userSource)

	fmt.Printf("Mapped User: %+v\n", userTarget)
	// Expected output (assuming default mapping):
	// Mapped User: &{UserID:123 FullName:John Doe EmailAddress:john.doe@example.com}
}
```

## Core Concepts

### `// mapper:generate` Annotation

This special comment placed directly above a Go interface tells the `mapper` tool that this interface should be processed for code generation. Only interfaces with this annotation will have their mapping implementations generated.

```go
// mapper:generate
type MyMappingInterface interface {
    // ... mapping methods ...
}
```

### Mapping Interfaces

Each method within a `// mapper:generate` annotated interface defines a single mapping operation. The method's signature (input parameters and return values) dictates the types involved in the mapping.

*   **Input:** Typically, a single input parameter representing the source type.
*   **Output:** A single return value representing the target type.
*   **Error Handling:** An optional second return value of type `error` can be used to indicate that the mapping might fail.

### Generated Code Structure

The `mapper` tool generates a new Go file (by default in a `gen` subdirectory relative to the parsed package) that contains an implementation struct for your mapping interface. For each method in your interface, a corresponding method will be generated on this implementation struct.

For example, for `type MyMapper interface { MapFoo(in *Foo) *Bar }`, the tool might generate:

```go
// gen/my_mapper_gen.go
package mypackagegen

import (
    "your_module/mypackage"
    "your_module/models" // Assuming Foo and Bar are in models
)

type MyMapperImpl struct{}

func (m *MyMapperImpl) MapFoo(in *models.Foo) *models.Bar {
    // ... generated mapping logic ...
}
```

## Detailed Usage & Capabilities

The `mapper` tool offers several ways to customize how fields are mapped, primarily through doc comments on the mapping interface methods.

### 1. Basic Field Mapping (Direct Assignment)

By default, the `mapper` tool attempts to map fields with identical names and compatible types directly.

**Example Interface:**

```go
// mapper:generate
type SimpleMapper interface {
	MapData(source *SourceData) *TargetData
}

type SourceData struct {
	ID   string
	Name string
}

type TargetData struct {
	ID   string
	Name string
}
```

**Generated Logic:**

```go
func (m *SimpleMapperImpl) MapData(source *SourceData) *TargetData {
	return &TargetData{
		ID:   source.ID,
		Name: source.Name,
	}
}
```

### 2. Renaming Fields (`from` tag)

Use the `// field:OutputFieldName from:"InputFieldName"` comment to map a field from a source field with a different name.

**Example Interface:**

```go
// mapper:generate
type RenamingMapper interface {
	// field:TargetID from:"SourceIdentifier"
	// field:TargetName from:"SourceName"
	MapRenamedData(source *SourceRenamedData) *TargetRenamedData
}

type SourceRenamedData struct {
	SourceIdentifier string
	SourceName       string
}

type TargetRenamedData struct {
	TargetID   string
	TargetName string
}
```

**Generated Logic:**

```go
func (m *RenamingMapperImpl) MapRenamedData(source *SourceRenamedData) *TargetRenamedData {
	return &TargetRenamedData{
		TargetID:   source.SourceIdentifier,
		TargetName: source.SourceName,
	}
}
```

### 3. Ignoring Fields (`-` tag)

Use the `// field:FieldName from:"-"` comment to explicitly ignore a field in the target struct, preventing it from being mapped.

**Example Interface:**

```go
// mapper:generate
type IgnoringMapper interface {
	// field:IgnoredField from:"-"
	MapIgnoredData(source *SourceIgnoredData) *TargetIgnoredData
}

type SourceIgnoredData struct {
	FieldA string
	FieldB string
}

type TargetIgnoredData struct {
	FieldA       string
	IgnoredField string // This field will not be mapped
}
```

**Generated Logic:**

```go
func (m *IgnoringMapperImpl) MapIgnoredData(source *SourceIgnoredData) *TargetIgnoredData {
	return &TargetIgnoredData{
		FieldA: source.FieldA,
		// IgnoredField is omitted
	}
}
```

### 4. Pointer Conversions

The tool can handle conversions between pointer and non-pointer types automatically, including dereferencing and taking addresses as needed.

**Example Interface:**

```go
// mapper:generate
type PointerMapper interface {
	MapToPtr(source *SourcePtr) *TargetPtr
	MapFromPtr(source *SourcePtr) TargetNonPtr
	MapPtrToNonPtr(source *SourceNonPtr) *TargetNonPtr
}

type SourcePtr struct {
	Value string
}

type TargetPtr struct {
	Value *string
}

type SourceNonPtr struct {
	Value string
}

type TargetNonPtr struct {
	Value string
}
```

**Generated Logic:**

```go
func (m *PointerMapperImpl) MapToPtr(source *SourcePtr) *TargetPtr {
	if source == nil {
		return nil
	}
	val := source.Value
	return &TargetPtr{
		Value: &val,
	}
}

func (m *PointerMapperImpl) MapFromPtr(source *SourcePtr) TargetNonPtr {
	if source == nil {
		return TargetNonPtr{} // Or handle as appropriate
	}
	return TargetNonPtr{
		Value: source.Value,
	}
}

func (m *PointerMapperImpl) MapPtrToNonPtr(source *SourceNonPtr) *TargetNonPtr {
	if source == nil {
		return nil
	}
	return &TargetNonPtr{
		Value: source.Value,
	}
}
```

### 5. Slice Conversions

The `mapper` tool can generate code to convert slices of one type to slices of another, provided a mapping exists for the element types.

**Example Interface:**

```go
// mapper:generate
type SliceMapper interface {
	MapSlice(sources []*SourceItem) []*TargetItem
}

type SourceItem struct {
	Name string
}

type TargetItem struct {
	ItemName string
}
```

**Generated Logic:**

```go
func (m *SliceMapperImpl) MapSlice(sources []*SourceItem) []*TargetItem {
	if sources == nil {
		return nil
	}
	targets := make([]*TargetItem, len(sources))
	for i := range sources {
		targets[i] = &TargetItem{
			ItemName: sources[i].Name,
		}
	}
	return targets
}
```

### 6. Error Handling

If a mapping method in your interface includes an `error` as its second return value, the generated code will include error handling. This is particularly useful when underlying conversions (e.g., custom type conversions) might fail.

**Example Interface:**

```go
// mapper:generate
type ErrorHandlingMapper interface {
	MapWithError(source *SourceWithError) (*TargetWithError, error)
}

type SourceWithError struct {
	Value string
}

type TargetWithError struct {
	ParsedValue int
}
```

**Generated Logic (assuming a custom converter for `Value` to `ParsedValue` that can error):**

```go
func (m *ErrorHandlingMapperImpl) MapWithError(source *SourceWithError) (*TargetWithError, error) {
	if source == nil {
		return nil, nil // Or an appropriate error
	}
	parsedValue, err := strconv.Atoi(source.Value) // Example custom conversion
	if err != nil {
		return nil, fmt.Errorf("failed to parse value: %w", err)
	}
	return &TargetWithError{
		ParsedValue: parsedValue,
	}, nil
}
```

### 7. Protobuf Enum Handling

The `mapper` tool automatically discovers and handles Protocol Buffer enums when they are part of the types being mapped. This simplifies conversions between Go structs and Protobuf messages that contain enum fields.

*(Note: Specific examples for Protobuf enum handling would require defining `.proto` files and generated Go code, which is beyond a simple example here. The tool integrates with `go/types` to recognize these enums.)*

An enum on the other side, such as the string enums ent generates, is mapped onto a Protobuf enum without annotation when their constants pair up by name. Case, underscores and the type name are ignored in the comparison, so `StateActive = "active"` of type `State` meets `SetState_SET_STATE_ACTIVE` of type `SetState` as `ACTIVE`. A string constant also matches by its value. The generated code is a switch with an error branch for values without a counterpart, so the method has to return an error. Constants that find no counterpart, like the usual `..._UNSPECIFIED`, are reported as a warning, or as an error with `-strict-enums`. Where names do not line up, spell the pairs out with a `map:` directive.

### 8. Indirect Type Mapping via Type Graph

A powerful feature of the `mapper` tool is its ability to find indirect mapping routes between types. Instead of requiring a direct conversion function for every possible type pair, the `mapper` constructs a graph of all available type conversions (both direct and those defined by your mapping interfaces).

If you request a mapping from `TypeA` to `TypeC`, and there's no direct `TypeA -> TypeC` conversion, the `mapper` will attempt to find an intermediate path, such as `TypeA -> TypeB -> TypeC`.

**Example:**

Suppose you have the following types and direct mappings:

*   `func (m *MyMapperImpl) MapAToB(a *TypeA) *TypeB`
*   `func (m *MyMapperImpl) MapBToC(b *TypeB) *TypeC`

If you then define a method in your interface like:

```go
// mapper:generate
type ComplexMapper interface {
	MapAToC(a *TypeA) *TypeC
}
```

The `mapper` tool will automatically chain `MapAToB` and `MapBToC` to fulfill the `MapAToC` request, generating code similar to:

```go
func (m *ComplexMapperImpl) MapAToC(a *TypeA) *TypeC {
	b := m.MapAToB(a)
	c := m.MapBToC(b)
	return c
}
```

This graph-based approach allows you to define smaller, focused mapping functions, and the `mapper` will intelligently combine them to achieve more complex transformations.

### 9. Custom Mapper Functions

The `mapper` tool allows you to define and register your own custom mapping functions for specific type conversions. This is useful when the default mapping logic (direct assignment, pointer/slice conversions) is insufficient, and you need to implement custom transformation logic.

To use a custom mapper function:

1.  **Define the Custom Mapper Function:**
    Create a Go function in any package. The function must have one input parameter (the source type) and either one return value (the target type) or two return values (the target type and an `error`).

    **Example Custom Mapper Function:**

    ```go
    // mypackage/converters.go
    package mypackage

    import (
    	"fmt"
    	"strconv"
    )

    // StringToInt converts a string to an int, returning an error if parsing fails.
    func StringToInt(s string) (int, error) {
    	return strconv.Atoi(s)
    }

    // BoolToString converts a boolean to its string representation.
    func BoolToString(b bool) string {
    	if b {
    		return "true"
    	}
    	return "false"
    }
    ```

2.  **Register the Custom Mapper Function:**
    Before running the `mapper` tool, you need to register your custom function using `mapper.RegisterCustomMapper`. This is typically done in an `init()` function or at the beginning of your `main` function. The `mapper` tool will use reflection to automatically determine the input and output types, as well as whether the function returns an error.

    ```go
    // mypackage/init.go
    package mypackage

    import "github.com/ligustah/mapper/mapper"

    func init() {
    	// Register StringToInt: maps string to int
    	mapper.RegisterCustomMapper(StringToInt)

    	// Register BoolToString: maps bool to string
    	mapper.RegisterCustomMapper(BoolToString)
    }
    ```
    *   The `customMapperFunc` argument is the function itself. The `mapper` tool will use reflection to extract its input and output types, its name, package, and error-returning capability.

3.  **Use the `using` Field Annotation:**
    In your mapping interface, use the `// field:OutputFieldName using:"CustomMapperFunc"` annotation on the target field that requires the custom conversion. `CustomMapperFunc` names your custom mapper function in one of three ways:
    *   by its full import path, e.g. `"github.com/you/app/mypackage.StringToInt"`
    *   by a bare name, e.g. `"StringToInt"`, for a function declared in the same package as the interface
    *   by the local name of one of the file's imports, e.g. `"mypackage.StringToInt"` when the file declaring the interface imports `mypackage`

    **Example Mapping Interface using Registered Custom Mappers:**

    ```go
    // mypackage/mappers.go
    package mypackage

    // mapper:generate
    type MyCustomMapper interface {
    	// field:Age using:"mypackage.StringToInt"
    	// field:IsActiveText using:"mypackage.BoolToString"
    	MapUser(source *UserSource) (*UserTarget, error)
    }

    type UserSource struct {
    	ID       string
    	AgeStr   string
    	IsActive bool
    }

    type UserTarget struct {
    	ID         string
    	Age        int
    	IsActiveText string
    }
    ```

**Generated Logic:**

```go
func (m *MyCustomMapperImpl) MapUser(source *UserSource) (*UserTarget, error) {
	if source == nil {
		return nil, nil
	}
	target := &UserTarget{
		ID: source.ID,
	}

	// Generated call to StringToInt
	age, err := mypackage.StringToInt(source.AgeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to map AgeStr: %w", err)
	}
	target.Age = age

	// Generated call to BoolToString
	target.IsActiveText = mypackage.BoolToString(source.IsActive)

	return target, nil
}
```

This feature provides a powerful way to extend the `mapper` tool's capabilities with your own domain-specific conversion logic.

### 10. Filling Builders

Some targets are not structs but builders, such as the `Create` and `UpdateOne` builders ent generates. A method whose output type is also one of its inputs fills that builder instead of returning a struct:

```go
// mapper:generate
type Mapper interface {
	// field:ID from:"-"
	CommitRow(in *proto.Commit, b *ent.CommitRowCreate) *ent.CommitRowCreate
}
```

Every method `SetX(v T)` on the builder that returns the builder is a target field named `X` of type `T`. It is resolved with the same rules and directives as a struct field and rendered as one setter call. The builder is returned for chaining, so the caller adds what the source does not carry:

```go
err := m.CommitRow(commit, tx.CommitRow.Create()).SetRef(ref.Hash).SetSetID(setID).Exec(ctx)
```

A nil source leaves the builder untouched. Two ent conventions are recognised: `SetNillableX(*T)` is used instead of `SetX(T)` when the source field or the converter hands back a pointer, so nil never turns into a zero value; and where an edge has both `SetX(*Entity)` and `SetXID(id)`, the id form is the field unless a directive addresses `X`. Variadic adders such as `AddFileIDs(...)` are not fields. As with structs, a setter without a source is an error until it is ignored with `from:"-"`.

Ent entities often have many optional columns a given source never carries. Rather than ignoring each of them, put `// mapper:skip-optional` on the method, or on the interface to cover all of its methods. An optional setter, one with a `SetNillableX` twin, that has neither a directive nor a source field of its name is then left unset. Required setters are still an error.

## Building and Running the Generator

To build the `mapper` tool from source:

```bash
go build -o mapper .
```

To run the generator on your project:

```bash
./mapper <path_to_package_to_parse>
```

For example, if your mapping interfaces are in `./pkg/mymap`, you would run:

```bash
./mapper ./pkg/mymap
```

The generated code will typically appear in a `gen` subdirectory within the parsed package (e.g., `./pkg/mymap/gen`).

## Development Conventions

*   **Interface-driven Mapping:** Define your mapping logic using Go interfaces.
*   **Doc Comment Annotations:** Use `// mapper:generate` on interfaces and `// field:Name from:"Source"` on methods to guide code generation.
*   **Error Handling:** Design mapping methods to return `error` when conversions might fail.
*   **Dependencies:** Project dependencies are managed via `go.mod`.