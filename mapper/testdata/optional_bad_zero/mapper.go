package optional_bad_zero

type In struct{ Count int }

type Out struct{ Count int }

// mapper:generate
type Mapper interface {
	// field:Count zero:"nil"
	Map(in *In) *Out
}
