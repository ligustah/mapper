package optional

import (
	"strconv"
	"time"
)

type Kind string

const (
	KindFile Kind = "file"
	KindDir  Kind = "dir"
)

type Clock struct{ seconds int64 }

func (c *Clock) Seconds() int64 { return c.seconds }

type Point struct{ X, Y int }

type Settings struct {
	On       *bool
	Share    *bool
	Name     *string
	Kind     *Kind
	Count    *int
	Limit    *string
	Clock    *Clock
	At       time.Time
	Created  int64
	Max      int
	Origin   Point
	Revision int
}

type Row struct {
	On       bool
	Share    bool
	Name     string
	Kind     Kind
	Count    int32
	Limit    int
	Clock    int64
	At       *time.Time
	Created  *time.Time
	Max      *int
	Origin   *Point
	Revision *int
}

func Limit(s string) (int, error) { return strconv.Atoi(s) }

func Unix(s int64) time.Time { return time.Unix(s, 0) }

// mapper:generate
type Mapper interface {
	// field:On default:"true"
	// field:Share default:"false"
	// field:Name default:"anonymous"
	// field:Kind default:"KindDir"
	// field:Count default:"-1"
	// field:Limit using:"Limit" default:"100"
	// field:Clock method:"Seconds" default:"0"
	// field:At zero:"nil"
	// field:Created using:"Unix" zero:"nil"
	// field:Max zero:"nil"
	// field:Origin zero:"nil"
	ToRow(in *Settings, revision int) (*Row, error)
}

type RowCreate struct{}

func (c *RowCreate) SetOn(v bool) *RowCreate          { return c }
func (c *RowCreate) SetNillableOn(v *bool) *RowCreate { return c }
func (c *RowCreate) SetMax(v int) *RowCreate          { return c }
func (c *RowCreate) SetNillableMax(v *int) *RowCreate { return c }

// mapper:generate
type Builder interface {
	// field:On default:"true"
	// field:Max zero:"nil"
	Fill(in *Settings, b *RowCreate) *RowCreate
}
