package enumfallback

type State string

const (
	StateActive  State = "active"
	StateClosing State = "closing"
	StateLost    State = "lost"
)

type SetState int32

const (
	SetState_SET_UNSPECIFIED SetState = 0
	SetState_SET_ACTIVE      SetState = 1
	SetState_SET_CLOSING     SetState = 2
)

type Kind int

const (
	KindFile Kind = iota
	KindDir
)

type Row struct {
	State   State
	Kind    Kind
	Prepaid bool
}

type Message struct {
	State   SetState
	Kind    string
	Prepaid Kind
}

// mapper:generate
type Mapper interface {
	// field:State map:"*=SetState_SET_UNSPECIFIED"
	// field:Kind map:"KindFile=file,KindDir=dir,*=unknown"
	// field:Prepaid map:"true=KindDir,*=KindFile"
	ToMessage(in *Row) *Message
}
