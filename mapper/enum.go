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

// parseEnumMap parses the value of a `map:"a=X,b=Y,*=Z"` directive. Both sides of a
// pair may be a constant name or the literal value of a constant. The target of a
// `*` source is returned as fallback, it stands for every value not listed.
func parseEnumMap(directive string) (entries []enumMapEntry, fallback string, err error) {

	seen := make(map[string]bool)
	for _, pair := range strings.Split(directive, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		from, to, ok := strings.Cut(pair, "=")
		if !ok {
			return nil, "", fmt.Errorf("expected pairs of the form source=target, got %q", pair)
		}

		from = strings.TrimSpace(from)
		to = strings.TrimSpace(to)

		if from == "" || to == "" {
			return nil, "", fmt.Errorf("both sides of %q must be non-empty", pair)
		}

		if seen[from] {
			return nil, "", fmt.Errorf("duplicate source value %q", from)
		}
		seen[from] = true

		if from == "*" {
			fallback = to
			continue
		}

		entries = append(entries, enumMapEntry{from: from, to: to})
	}

	if len(entries) == 0 && fallback == "" {
		return nil, "", errors.New("no mappings given")
	}

	return entries, fallback, nil
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

// enumKey reduces a constant to the word that names its value, for matching the
// constants of two enums that spell their names differently. Case and underscores
// are ignored, and the type name is stripped as often as it leads, along with the
// outer message of a nested protoc enum. ent's StateActive of type State, protoc's
// SetState_SET_STATE_ACTIVE of type SetState and Commit_ACTIVE of type Commit_State
// all come down to ACTIVE.
func enumKey(constName, typeName string) string {
	key := squashName(constName)

	prefixes := []string{squashName(typeName)}
	if i := strings.LastIndex(typeName, "_"); i > 0 {
		prefixes = append(prefixes, squashName(typeName[:i]))
	}

	for stripped := true; stripped; {
		stripped = false
		for _, prefix := range prefixes {
			if prefix != "" && len(key) > len(prefix) && strings.HasPrefix(key, prefix) {
				key = key[len(prefix):]
				stripped = true
			}
		}
	}

	return key
}

func squashName(name string) string {
	return strings.ToUpper(strings.ReplaceAll(name, "_", ""))
}

// enumPair is a source constant and the target constant it maps onto.
type enumPair struct {
	from enumConstant
	to   enumConstant
}

// matchEnumConstants pairs the constants of two enum types by name, see enumKey. A
// string constant is also matched by its value, so ent's StateActive = "active"
// finds ACTIVE either way. A source constant with no counterpart, or with more than
// one, is returned in unmatched. Both are nil when either type declares no
// constants.
func matchEnumConstants(from, to types.Type) (pairs []enumPair, unmatched []string) {
	fromConsts, toConsts := enumConstants(from), enumConstants(to)
	if len(fromConsts) == 0 || len(toConsts) == 0 {
		return nil, nil
	}

	fromName, toName := typeObjName(from), typeObjName(to)

	byKey := map[string][]enumConstant{}
	for _, c := range toConsts {
		for _, key := range enumKeys(c, toName) {
			byKey[key] = appendConstant(byKey[key], c)
		}
	}

	// a second chance for constants left over: protoc's SetState_SET_ACTIVE carries
	// a prefix beyond its type name, but one all constants of the type share
	fromPrefix, toPrefix := sharedPrefix(fromConsts), sharedPrefix(toConsts)
	byShortKey := map[string][]enumConstant{}
	for key, consts := range byKey {
		byShortKey[key] = append(byShortKey[key], consts...)
	}
	for _, c := range toConsts {
		if key := shortKey(c, toPrefix); key != "" {
			byShortKey[key] = appendConstant(byShortKey[key], c)
		}
	}

	for _, c := range fromConsts {
		candidates := lookupCandidates(byKey, enumKeys(c, fromName))

		if len(candidates) == 0 {
			keys := enumKeys(c, fromName)
			if key := shortKey(c, fromPrefix); key != "" {
				keys = append(keys, key)
			}
			candidates = lookupCandidates(byShortKey, keys)
		}

		if len(candidates) != 1 {
			unmatched = append(unmatched, c.name)
			continue
		}

		pairs = append(pairs, enumPair{from: c, to: candidates[0]})
	}

	return pairs, unmatched
}

func lookupCandidates(byKey map[string][]enumConstant, keys []string) []enumConstant {
	var candidates []enumConstant
	for _, key := range keys {
		for _, target := range byKey[key] {
			candidates = appendConstant(candidates, target)
		}
	}

	return candidates
}

// sharedPrefix is the longest prefix ending in an underscore that the names of all
// consts share, or an empty string for fewer than two constants.
func sharedPrefix(consts []enumConstant) string {
	if len(consts) < 2 {
		return ""
	}

	prefix := consts[0].name
	for _, c := range consts[1:] {
		for !strings.HasPrefix(c.name, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}

	return prefix[:strings.LastIndex(prefix, "_")+1]
}

// shortKey is the key of a constant with prefix cut off its name, or an empty
// string without a prefix.
func shortKey(c enumConstant, prefix string) string {
	if prefix == "" {
		return ""
	}

	return squashName(strings.TrimPrefix(c.name, prefix))
}

// enumKeys are the keys a constant is matched by: its name and, for a string
// constant, its value.
func enumKeys(c enumConstant, typeName string) []string {
	keys := []string{enumKey(c.name, typeName)}
	if c.value.Kind() == constant.String {
		if value := squashName(constant.StringVal(c.value)); value != "" && value != keys[0] {
			keys = append(keys, value)
		}
	}

	return keys
}

func appendConstant(consts []enumConstant, c enumConstant) []enumConstant {
	for _, existing := range consts {
		if existing.name == c.name {
			return consts
		}
	}

	return append(consts, c)
}

// typeObjName returns the declared name of a named type, or an empty string.
func typeObjName(typ types.Type) string {
	named, ok := typ.(*types.Named)
	if !ok {
		return ""
	}

	return named.Obj().Name()
}
