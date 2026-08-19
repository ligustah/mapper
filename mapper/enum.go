package mapper

import (
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

// typeObjName returns the declared name of a named type, or an empty string.
func typeObjName(typ types.Type) string {
	named, ok := typ.(*types.Named)
	if !ok {
		return ""
	}

	return named.Obj().Name()
}
