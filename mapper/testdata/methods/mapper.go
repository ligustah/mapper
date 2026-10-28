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
}

func (p *Probe) Passed() bool { return p.Status == "ok" }

type Result struct {
	Passed bool
	Body   string
	Blobs  []string
	Sums   []string
	Totals map[string]string
}

func Hex(v int) string { return strconv.FormatInt(int64(v), 16) }

// mapper:generate
type Mapper interface {
	// field:Passed from:"." method:"Passed"
	// field:Body method:"Decode"
	// field:Blobs method:"Decode"
	// field:Sums using:"Hex"
	// field:Totals using:"Hex"
	ToResult(in *Probe) (*Result, error)
}
