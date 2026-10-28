package inline

type Report struct {
	Total int
	Count int
}

type Host struct {
	Name  string
	Count int
}

type Row struct {
	ID     string
	Report *Report
	Host   Host
}

type Revenue struct {
	ID    string
	Total int
	Count int
	Name  string
}

// mapper:generate
type Mapper interface {
	// mapper:inline Report Host
	ToRevenue(in *Row) *Revenue
}

// mapper:generate
// mapper:inline Host
type Hosts interface {
	// field:Total from:"-"
	// field:Count from:"Report.Count"
	ToHostRevenue(in *Row) *Revenue
}
