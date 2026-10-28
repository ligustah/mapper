package nocopy

import "sync"

// Bracket stands in for a protobuf message, which embeds a [0]sync.Mutex so that
// go vet reports every copy.
type Bracket struct {
	_     [0]sync.Mutex
	Every int64
	Keep  int32
}

type Policy struct {
	_        [0]sync.Mutex
	Name     string
	Brackets []*Bracket
	Primary  *Bracket
}

type StoreBracket struct {
	Every int64
	Keep  int32
}

type StorePolicy struct {
	Name     string
	Brackets []StoreBracket
	Primary  *StoreBracket
}

// mapper:generate
type Mapper interface {
	ToStore(in *Policy) *StorePolicy
	FromStore(in *StorePolicy) *Policy
}
