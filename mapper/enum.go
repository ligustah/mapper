package mapper

import (
	"errors"
	"fmt"
	"go/constant"
	"go/types"
	"strings"
)

// enumConstant is a single constant declared for a named type.
type enumConstant struct {
	name  string
	value constant.Value
}

// enumConstants collects the constants declared for a named type, looking in the
// package that declares the type. It returns nil for anything that does not look
// like an enum, i.e. a named type with no constants of its own.
func enumConstants(typ types.Type) []enumConstant {
	named, ok := typ.(*types.Named)
	if !ok {
		return nil
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return nil
	}

	scope := obj.Pkg().Scope()

	var out []enumConstant
	// Names() is sorted, so the reported problems are stable across runs
	for _, name := range scope.Names() {
		c, ok := scope.Lookup(name).(*types.Const)
		if !ok {
			continue
		}

		if !types.Identical(c.Type(), named) {
			continue
		}

		out = append(out, enumConstant{name: c.Name(), value: c.Val()})
	}

	return out
}

// normalizeEnumName reduces a constant name to a form that can be compared across
// two enum types. Constants commonly carry their type name as a prefix, so
// galaxy.StateActive and store.GalaxyStateActive both normalize to "ACTIVE", as
// does the protobuf style Transition_State_ACTIVE for type Transition_State.
func normalizeEnumName(constName, typeName string) string {
	name := constName

	if typeName != "" && len(name) > len(typeName) && strings.HasPrefix(name, typeName) {
		name = name[len(typeName):]
	}

	name = strings.ReplaceAll(name, "_", "")

	return strings.ToUpper(name)
}

// checkEnumConversion compares the constants of two enum-like named types that a
// plain type conversion is about to be generated for. Such a conversion keeps the
// underlying value as-is, which is only correct if the target declares that value
// with the same meaning. Two problems are reported:
//
//   - a source constant whose value the target does not declare at all, which
//     yields a target value outside its own value set
//   - a source constant whose name matches a target constant carrying a different
//     value, which silently changes what the value means
//
// Constants of the target that the source does not have are fine, a wider target
// can represent everything the source can. Types without constants are not enums
// and are not reported on.
func checkEnumConversion(from, to types.Type) []string {
	fromConsts := enumConstants(from)
	toConsts := enumConstants(to)

	if len(fromConsts) == 0 || len(toConsts) == 0 {
		return nil
	}

	fromName := typeObjName(from)
	toName := typeObjName(to)

	byValue := make(map[string]enumConstant, len(toConsts))
	byName := make(map[string]enumConstant, len(toConsts))

	for _, c := range toConsts {
		key := c.value.ExactString()
		if _, ok := byValue[key]; !ok {
			byValue[key] = c
		}

		normalized := normalizeEnumName(c.name, toName)
		if _, ok := byName[normalized]; !ok {
			byName[normalized] = c
		}
	}

	var problems []string
	for _, c := range fromConsts {
		if _, ok := byValue[c.value.ExactString()]; !ok {
			problems = append(problems, fmt.Sprintf(
				"%s (%s) has no counterpart in %s", c.name, c.value.ExactString(), to))
			continue
		}

		match, ok := byName[normalizeEnumName(c.name, fromName)]
		if ok && match.value.ExactString() != c.value.ExactString() {
			problems = append(problems, fmt.Sprintf(
				"%s (%s) maps onto %s, which is %s", c.name, c.value.ExactString(), match.name, match.value.ExactString()))
		}
	}

	return problems
}

// enumMapEntry is one `source=target` pair of a `map:` directive.
type enumMapEntry struct {
	from string
	to   string
}

// parseEnumMap parses the value of a `map:"a=X,b=Y"` directive. Both sides of a
// pair may be a constant name or the literal value of a constant.
func parseEnumMap(directive string) ([]enumMapEntry, error) {
	var entries []enumMapEntry

	seen := make(map[string]bool)
	for _, pair := range strings.Split(directive, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		from, to, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, fmt.Errorf("expected pairs of the form source=target, got %q", pair)
		}

		from = strings.TrimSpace(from)
		to = strings.TrimSpace(to)

		if from == "" || to == "" {
			return nil, fmt.Errorf("both sides of %q must be non-empty", pair)
		}

		if seen[from] {
			return nil, fmt.Errorf("duplicate source value %q", from)
		}
		seen[from] = true

		entries = append(entries, enumMapEntry{from: from, to: to})
	}

	if len(entries) == 0 {
		return nil, errors.New("no mappings given")
	}

	return entries, nil
}

// lookupEnumConstant resolves a reference from a `map:` directive against the
// constants of a type. A reference matches either the name of a constant or its
// literal value, so both `StateActive` and `active` find the same constant. The
// second return value reports whether the type declares constants at all.
func lookupEnumConstant(typ types.Type, ref string) (enumConstant, bool, bool) {
	consts := enumConstants(typ)
	if len(consts) == 0 {
		return enumConstant{}, false, false
	}

	for _, c := range consts {
		if c.name == ref {
			return c, true, true
		}
	}

	for _, c := range consts {
		if constantMatchesLiteral(c.value, ref) {
			return c, true, true
		}
	}

	return enumConstant{}, false, true
}

// constantMatchesLiteral compares a constant against the literal spelling used in
// a directive, e.g. the value `"active"` against the reference `active`.
func constantMatchesLiteral(value constant.Value, ref string) bool {
	switch value.Kind() {
	case constant.String:
		return constant.StringVal(value) == ref
	case constant.Int, constant.Float:
		return value.ExactString() == ref
	default:
		return false
	}
}

// enumConstantNames returns the names of the constants declared for a type.
func enumConstantNames(typ types.Type) []string {
	consts := enumConstants(typ)
	if len(consts) == 0 {
		return nil
	}

	names := make([]string, 0, len(consts))
	for _, c := range consts {
		names = append(names, c.name)
	}

	return names
}

// typeObjName returns the declared name of a named type, or an empty string.
func typeObjName(typ types.Type) string {
	named, ok := typ.(*types.Named)
	if !ok {
		return ""
	}

	return named.Obj().Name()
}
