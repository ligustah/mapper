package skipunmatched_directive

type Row struct {
	ID   string
	Name string
}

type Message struct {
	ID string
}

// mapper:generate
// mapper:skip-unmatched
type Mapper interface {
	// field:Name from:"Title"
	ToRow(in *Message) *Row
}
