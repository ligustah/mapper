package pointermethods

type Ref struct{ Sum string }

type Blob struct{ data []byte }

// Encoded hands back a pointer of its own, as a codec does.
func (b *Blob) Encoded() (*Ref, error) { return &Ref{Sum: string(b.data)}, nil }

type Money struct{ cents int64 }

func (m Money) String() string { return "" }

func Wrap(b Blob) *Ref { return &Ref{Sum: string(b.data)} }

type Source struct {
	Body    *Blob
	Bodies  []*Blob
	Raw     *Blob
	Amounts map[string]Money
}

type Target struct {
	Body    *Ref
	Bodies  []*Ref
	Raw     *Ref
	Amounts map[string]string
}

// mapper:generate
type Mapper interface {
	// field:Body method:"Encoded"
	// field:Bodies method:"Encoded"
	// field:Raw using:"Wrap"
	// field:Amounts method:"String"
	ToTarget(in *Source) (*Target, error)
}
