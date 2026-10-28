# mapper

`mapper` generates type-safe conversion code between Go types. Declare the conversions you need as methods of an annotated interface, and the generator writes the implementation: field by field, through pointers, slices, maps and nested structs, with enums, custom converters and error handling.

It is built for the boundary code that piles up between layers: database models and Protocol Buffer messages, ent entities and API types, one service's structs and another's.

```go
// mapper:generate
type UserMapper interface {
	// field:FullName from:"Name"
	// field:Password from:"-"
	ToView(in *models.User) (*models.UserView, error)
	ToViews(in []*models.User) ([]*models.UserView, error)
}
```

## Installation

mapper requires Go 1.27 or newer. Add it as a tool of your module:

```bash
go get -tool github.com/ligustah/mapper@latest
```

and run it with `go generate`, from a `//go:generate` line in the package that declares your interfaces:

```go
//go:generate go tool mapper .
```

Alternatively install the binary with `go install github.com/ligustah/mapper@latest` and run `mapper <package>`.

## Quick start

Given two structs:

```go
package models

type User struct {
	ID       string
	Name     string
	Email    string
	Age      int
	Password string
	Address  *Address
	Tags     []Tag
	Score    string
}

func (u *User) DisplayName() string { return u.Name + " <" + u.Email + ">" }

type Address struct{ Street, City string }

type Tag struct{ Label string }

type UserView struct {
	ID          string
	FullName    string
	Email       string
	Age         int32
	Password    string
	Address     *AddressView
	Tags        []*TagView
	Score       int64
	DisplayName string
}

type AddressView struct{ Street, City string }

type TagView struct{ Label string }
```

declare the conversion in a package of your own, annotated with `// mapper:generate`:

```go
package mapping

import "example.com/app/models"

//go:generate go tool mapper .

// mapper:generate
type UserMapper interface {
	// field:FullName from:"Name"
	// field:Password from:"-"
	// field:DisplayName from:"." method:"DisplayName"
	ToView(in *models.User) (*models.UserView, error)

	ToViews(in []*models.User) ([]*models.UserView, error)
}
```

`go generate ./...` writes `mapping/gen/gen.go`:

```go
func (m UserMapperImpl) ToView(in *models.User) (*models.UserView, error) {
	if in == nil {
		return nil, nil
	}
	var tmp0 *models.AddressView
	if tmp1 := in.Address; tmp1 != nil {
		tmp2 := *tmp1
		tmp3 := models.AddressView(tmp2)
		tmp0 = &tmp3
	}
	tmp4 := make([]*models.TagView, len(in.Tags))
	for i := range in.Tags {
		tmp5 := m.mapTagToTagView(in.Tags[i])
		tmp4[i] = &tmp5
	}

	tmp6, err := strconv.ParseInt(in.Score, 10, 64)
	if err != nil {
		return nil, err
	}

	return &models.UserView{
		ID:          in.ID,
		FullName:    in.Name,
		Email:       in.Email,
		Age:         int32(in.Age),
		Address:     tmp0,
		Tags:        tmp4,
		Score:       tmp6,
		DisplayName: in.DisplayName(),
	}, nil
}

func (m UserMapperImpl) ToViews(in []*models.User) ([]*models.UserView, error) {
	tmp0 := make([]*models.UserView, len(in))
	for i := range in {
		tmp1, err := m.ToView(in[i])
		if err != nil {
			return nil, err
		}

		tmp0[i] = tmp1
	}

	return tmp0, nil
}

func (m UserMapperImpl) mapTagToTagView(in models.Tag) models.TagView {
	return models.TagView{
		Label: in.Label,
	}
}

type UserMapperImpl struct{}

var _ mapping.UserMapper = (*UserMapperImpl)(nil)
```

and you use it through the interface:

```go
var m mapping.UserMapper = gen.UserMapperImpl{}
view, err := m.ToView(user)
```

The generated file is deterministic: regenerating an unchanged package gives the same bytes.

## How fields are mapped

Every exported field of the target is filled from the inputs:

1. A [directive](#directives) on the field decides, if there is one.
2. Otherwise the input field of the same name is used. Only if there is none is the name compared case-insensitively, so protoc's `Id` finds `ID`. Two fields that both match that way are an error.
3. With several inputs, a field may also come from a parameter of its name: `Row(in *Policy, closedAtNs int64) *Row` sets `ClosedAtNs`.
4. A field nothing provides is an error, unless [`mapper:skip-unmatched`](#toggles) is set.

The value is then converted to the field's type by the first rule that applies:

- **Assignable** types are copied, with nil checks and addresses taken between `T` and `*T` as needed. A nil pointer becomes the zero value.
- **A converter**: a method of the same interface, a [custom converter](#custom-converters) or a builtin one, possibly chained. `TypeA → TypeB → TypeC` is found when the interface declares both steps.
- **A plain Go conversion** between types with the same underlying type (two `type X string`), integers of any width, and `string` ↔ `[]byte`. `int` ↔ `float64` or `string` ↔ `int` need a converter. `time.Duration` never converts to or from an integer on its own, since a plain conversion would silently carry nanoseconds.
- **Enums** whose constants pair up by name, see [Enums](#enums).
- **Containers**: `[]T` → `[]U`, `[N]T` → `[N]U` and `map[K]V` → `map[K2]V2` convert element by element; `*T` → `*U` converts the value behind the pointer and keeps nil.
- **Nested structs**: two named structs nothing else converts are mapped field by field through a helper generated on the implementation (`mapTagToTagView` above), used wherever the pair appears, including structs that refer to themselves. A struct holding a lock, such as any protobuf message, gets a helper over pointers so that it is never copied. Directives don't reach inside a helper; declare a method when a nested pair needs them.

Builtin converters cover `string` ↔ integers via `strconv` (parsing can fail), `time.Time` ↔ `*timestamppb.Timestamp`, `time.Duration` ↔ `*durationpb.Duration`, and the `null.String`/`null.Time` types of `volatiletech/null`. The protobuf and null converters are only available when your module depends on those packages.

A method whose result is not a struct, such as `ToViews` above, converts its single input as a whole by the same rules.

## Directives

Directives are comments on an interface method, one per target field: `// field:<Target> <tag>:"<value>"`.

| Directive | Effect |
| --- | --- |
| `from:"Source"` | read from another field |
| `from:"Edges.Owner.Name"` | read from a nested field, nil-safe (see [Nested paths](#nested-paths)) |
| `from:"."` | convert the input itself, e.g. to build a struct out of the input's own fields |
| `from:"-"` | leave the field unset |
| `using:"Func"` | convert with a function, see [Custom converters](#custom-converters) |
| `method:"Name"` | call a method on the source value, see below |
| `map:"a=X,b=Y,*=Z"` | map listed values explicitly, see [Enums](#enums) |
| `default:"X"` | what a nil pointer source gives instead of the zero value, see below |
| `zero:"nil"` | a zero source value gives a nil pointer field, see below |

`method:"Decode"` calls a method without arguments on the source field. It may return `(T, error)`. A pointer source is nil-checked first, a slice or map has the method called on every element or value, and the result is converted to the field's type like any value. With `from:"."` the method is called on the input itself:

```go
// field:Body method:"Decode"
// field:Passed from:"." method:"Passed"
ToResult(in *Probe) (*Result, error)
```

### Nil and zero values

A nil pointer source gives the zero value of a value field. `default:"X"` gives `X` instead, a constant of the field's type by name or value, or a literal (strings, numbers, `true`/`false`). The other way round, a pointer field filled from a value is always set, and `zero:"nil"` leaves it nil when the source is its zero value; a type with an `IsZero()` method, like `time.Time`, is asked through it. Both work with `using:` and `method:` too, and on builders, where `zero:"nil"` goes through `SetNillableX`:

```go
// field:Share default:"true"
// field:Kind default:"KindDir"
// field:At zero:"nil"
// field:Created using:"Unix" zero:"nil"
ToRow(in *Settings) *Row
```

```go
var tmp0 bool = true
if tmp1 := in.Share; tmp1 != nil {
	tmp2 := *tmp1
	tmp0 = tmp2
}
...
var tmp6 *time.Time
if !in.At.IsZero() {
	tmp7 := in.At
	tmp6 = &tmp7
}
var tmp8 *time.Time
if in.Created != 0 {
	tmp9 := mapping.Unix(in.Created)
	tmp8 = &tmp9
}
```

They apply to a single source field, not element by element, and not together with `map:`, whose `*=X` entry plays the part of a default.

## Toggles

Toggles stand on a line of their own, on a method or on the interface to cover all of its methods.

`// mapper:skip-unmatched` leaves every target field without a directive and without a source of its name unset, rather than reporting it. That is useful for ent's `Edges`, IDs the database assigns and similar. A field whose source exists but cannot be converted, and a directive naming a missing source, still fail. A field that has a source of its name but must stay unset still needs its `from:"-"`.

`// mapper:inline Settings` reads a field without a source of its name from `Settings.<Name>`, as if `from:"Settings.<Name>"` were written for it. Several names are tried in order: `// mapper:inline Report Host`.

```go
// mapper:inline Settings
// field:Created from:"-"
Profile(in *models.UserMessage, owner string) *models.Profile
```

```go
func (m RowMapperImpl) Profile(in *models.UserMessage, owner string) *models.Profile {
	var tmp0 string
	if in.Settings != nil {
		tmp0 = in.Settings.Theme
	}

	var tmp1 string
	if in.Settings != nil {
		tmp1 = in.Settings.Language
	}

	return &models.Profile{
		Name:     in.Name,
		Theme:    tmp0,
		Language: tmp1,
		Owner:    owner,
	}
}
```

`// mapper:skip-optional` applies to [builders](#builders) only.

## Enums

Two enum types with the same underlying type and the same values convert directly. When the values differ, the constants are paired by name: case, underscores, the type name and a prefix shared by all constants are ignored, so ent's `StateActive = "active"` meets protoc's `SetState_SET_STATE_ACTIVE` or `SetState_SET_ACTIVE` as `ACTIVE`. A string constant also matches by its value. The generated switch has an error branch for values outside the declared constants, so the method has to return an error:

```go
var tmp6 models.StatusView
switch in.Status {
case models.StatusActive:
	tmp6 = models.StatusView_STATUS_ACTIVE
case models.StatusBlocked:
	tmp6 = models.StatusView_STATUS_BLOCKED
default:
	return nil, fmt.Errorf("unmapped value for StatusView: %v", in.Status)
}
```

Source constants without a counterpart are reported as warnings, or as errors with `-strict-enums`.

Where names don't line up, list the pairs with `map:`. Either side is a constant name or a literal value (strings, integers or `true`/`false`), so an enum also maps onto a plain string or bool. `*=X` replaces the error branch with a default value, and on its own pairs everything else by name:

```go
// field:Role map:"admin=RoleView_ROLE_ADMIN,member=RoleView_ROLE_MEMBER,*=RoleView_ROLE_UNKNOWN"
// field:State map:"*=SetState_SET_UNSPECIFIED"
```

`map:` works on values, on `*T` → `*U` and on `[]T` → `[]U`.

## Custom converters

`using:"Func"` converts a field with a function of your own, `func(T) U` or `func(T) (U, error)`:

```go
// Unix turns a time into seconds.
func Unix(t time.Time) int64 { return t.Unix() }

// mapper:generate
type RowMapper interface {
	// field:Created from:"CreatedAt" using:"Unix"
	Stamp(in *models.User) *models.Profile
}
```

The function is named as:

- `"Unix"`, a function of the interface's own package
- `"conv.Unix"`, through an import of the file that declares the interface
- `"example.com/app/conv.Unix"`, by its full import path
- `"example.com/app/conv.(*Clock).Unix"`, a method without arguments on the source type

A pointer source is nil-checked before the call. For slices and maps, a function of the element type is applied to every element or value: with `func Hex(int) string`, `using:"Hex"` maps `[]int` onto `[]string`.

A converter needed in many places is better declared once as a method of the interface. Every method with a single input is a converter for the rest of the interface, and the generator chains them.

`mapper.RegisterCustomMapper(fn)` registers a converter for every field of its types. It only takes effect in a generator binary of your own that imports `github.com/ligustah/mapper/mapper`, since the `mapper` command cannot run code from the package it generates for.

## Builders

A method whose output is also one of its inputs fills a builder instead of returning a struct, the shape of ent's `Create` and `UpdateOne` builders. Every method `SetX(v T)` returning the builder is a target field `X`:

```go
// field:CreatedAt from:"-"
UserRow(in *models.UserMessage, b *models.UserCreate) *models.UserCreate
```

```go
func (m RowMapperImpl) UserRow(in *models.UserMessage, b *models.UserCreate) *models.UserCreate {
	if in == nil {
		return b
	}
	b.SetID(in.Id)
	b.SetName(in.Name)
	b.SetNillableNote(in.Note)
	b.SetTeamID(in.TeamId)

	return b
}
```

The builder comes back for chaining, so the caller adds what the source doesn't carry: `m.UserRow(msg, client.User.Create()).SetCreatedAt(now).Exec(ctx)`. A nil source leaves the builder untouched. Two ent conventions are built in:

- `SetNillableX(*T)` is used instead of `SetX(T)` when the source hands over a pointer, so nil never becomes a zero value.
- For an edge with both `SetX(*Entity)` and `SetXID(id)`, the id form is the field, unless a directive addresses `X`.

Variadic adders like `AddTagIDs(...)` are not fields. `// mapper:skip-optional` leaves optional setters, the ones with a `SetNillableX` twin, unset when nothing provides them; required setters still need a source.

## Nested paths

`from:"Edges.Owner.Name"` never dereferences a nil pointer on the way: the field keeps its zero value instead. When the method returns an error and the type holding the pointer has a `<Name>OrErr() (*T, error)` accessor, as ent generates for every edge, the path is read through it, so an edge that wasn't loaded is reported as an error.

## Errors

A method may return `(T, error)`. It has to exactly when one of its conversions can fail (parsing, an enum switch without a default, a fallible converter, an `OrErr` edge), and the generator reports the mismatch either way. Errors from converters are returned unchanged.

## Running the generator

```
mapper [-strict-enums] <package>
```

The generator loads `<package>`, finds every interface annotated with `// mapper:generate` and writes their implementations, `<Interface>Impl`, to `<package>/gen/gen.go` in package `gen`.

| Flag | Effect |
| --- | --- |
| `-strict-enums` | fail instead of warning when a generated enum conversion can produce a value the target does not declare |

The `example/` directory has a larger set of mapping interfaces with their generated code and tests.

## License

Apache License 2.0, see [LICENSE](LICENSE).
