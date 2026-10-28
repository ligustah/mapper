package methods

import "strconv"

type Blob struct{ data []byte }

func (b Blob) Decode() (string, error) { return string(b.data), nil }

type Probe struct {
	Status string
	Body   *Blob
	Blobs  []*Blob
	Sums   []int
	Totals map[string]int
	Level  *Blob
	Sizes  []Blob
	Size   *Blob
}

func (p *Probe) Passed() bool { return p.Status == "ok" }

func (p *Probe) Ceiling() int { return len(p.Status) }

type Level int

func (b Blob) Size() int { return len(b.data) }

type Result struct {
	Passed  bool
	Body    string
	Blobs   []string
	Sums    []string
	Totals  map[string]string
	StopsAt int32
	Level   Level
	Sizes   []int64
	Size    *int32
}

func Hex(v int) string { return strconv.FormatInt(int64(v), 16) }

// mapper:generate
type Mapper interface {
	// field:Passed from:"." method:"Passed"
	// field:Body method:"Decode"
	// field:Blobs method:"Decode"
	// field:Sums using:"Hex"
	// field:Totals using:"Hex"
	// field:StopsAt from:"." method:"Ceiling"
	// field:Level method:"Size"
	// field:Sizes method:"Size"
	// field:Size method:"Size"
	ToResult(in *Probe) (*Result, error)
}
