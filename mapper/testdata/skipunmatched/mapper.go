package skipunmatched

type Row struct {
	ID    string
	Name  string
	Edges RowEdges
	Extra int
}

type RowEdges struct{}

type Message struct {
	ID   string
	Name string
}

// mapper:generate
// mapper:skip-unmatched
type Lenient interface {
	ToRow(in *Message) *Row
}

// mapper:generate
type PerMethod interface {
	// mapper:skip-unmatched
	ToRowSparse(in *Message) *Row
}

type Builder struct{ values map[string]any }

func (b *Builder) SetID(v string) *Builder    { b.values["id"] = v; return b }
func (b *Builder) SetOwner(v string) *Builder { b.values["owner"] = v; return b }

// mapper:generate
// mapper:skip-unmatched
type Fill interface {
	Fill(in *Message, b *Builder) *Builder
}
