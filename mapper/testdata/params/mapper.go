package params

type Policy struct {
	Name string
}

type Row struct {
	Name       string
	ClosedAtNs int64
	Owner      string
	Actor      string
}

// mapper:generate
type Mapper interface {
	ToRow(in *Policy, closedAtNs int64, actor string, owner string) *Row
}
