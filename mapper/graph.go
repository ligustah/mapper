package mapper

import (
	"github.com/yourbasic/graph"
)

type TypeGraph struct {
	graph *graph.Mutable

	typeIndex map[string]int
	types     []*GoType
	mappers   map[int]map[int]*TypeMapping
}

func (t *TypeGraph) storeTypIndex(typ *GoType) int {
	//if typ.Kind == PointerKind {
	//	typ = typ.Element
	//}

	if idx, ok := t.typeIndex[typ.String()]; ok {
		return idx
	}

	t.types = append(t.types, typ)
	idx := len(t.types) - 1

	t.typeIndex[typ.String()] = idx
	return idx
}

func (t *TypeGraph) getTypeIndex(typ *GoType) (int, bool) {
	//if typ.Kind == PointerKind {
	//	typ = typ.Element
	//}

	// look for a direct match
	if idx, ok := t.typeIndex[typ.String()]; ok {
		return idx, ok
	}

	if typ.Kind == PointerKind {
		// look for a type match
		if idx, ok := t.typeIndex[typ.Element.String()]; ok {
			return idx, ok
		}
	} else {
		// look for a pointer match
		if idx, ok := t.typeIndex[typ.pointer().String()]; ok {
			return idx, ok
		}
	}

	return -1, false
}

func (t *TypeGraph) getMapping(from, to *GoType) []*TypeMapping {
	fromIdx, ok := t.getTypeIndex(from)
	if !ok {
		return nil
	}

	toIdx, ok := t.getTypeIndex(to)
	if !ok {
		return nil
	}

	path, length := graph.ShortestPath(t.graph, fromIdx, toIdx)
	if length == -1 {
		return nil
	}

	var mappings []*TypeMapping

	for i := 0; i < len(path)-1; i++ {
		from := path[i]
		to := path[i+1]

		mappings = append(mappings, t.mappers[from][to])
	}

	return mappings
}

func (t *TypeGraph) add(mapping *TypeMapping) {
	fromIdx := t.storeTypIndex(mapping.types.from)
	toIdx := t.storeTypIndex(mapping.types.to)

	t.graph.Add(fromIdx, toIdx)

	if _, ok := t.mappers[fromIdx]; !ok {
		t.mappers[fromIdx] = make(map[int]*TypeMapping)
	}

	t.mappers[fromIdx][toIdx] = mapping
}

func NewTypeGraph() *TypeGraph {
	return &TypeGraph{
		graph:     graph.New(1000),
		typeIndex: make(map[string]int),
		mappers:   map[int]map[int]*TypeMapping{},
	}
}
