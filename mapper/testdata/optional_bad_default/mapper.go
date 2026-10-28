package optional_bad_default

type In struct{ Count int }

type Out struct{ Count int }

// mapper:generate
type Mapper interface {
	// field:Count default:"1"
	Map(in *In) *Out
}
