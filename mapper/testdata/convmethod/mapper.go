package convmethod

type Item struct{ Name string }

type ItemView struct{ Name string }

type State string

const (
	StateOpen   State = "open"
	StateClosed State = "closed"
)

type StateView int32

const (
	StateView_UNKNOWN StateView = 0
	StateView_OPEN    StateView = 1
	StateView_CLOSED  StateView = 2
)

// mapper:generate
type Mapper interface {
	ToView(in *Item) *ItemView
	ToViews(in []*Item) []*ItemView
	ToState(in State) (StateView, error)
}
