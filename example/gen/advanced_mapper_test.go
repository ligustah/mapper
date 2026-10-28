package gen

import (
	"testing"
	"time"

	"github.com/ligustah/mapper/example/converters"
	"github.com/ligustah/mapper/example/models"
	"github.com/ligustah/mapper/example/protos"
)

func TestAdvancedMapper_MapUser(t *testing.T) {
	m := AdvancedMapperImpl{}
	now := time.Now()

	user := &models.User{
		ID:        "user-1",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: now,
	}

	proto := m.MapUser(user)

	if proto.ID != user.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, user.ID)
	}
	if proto.Name != user.Name {
		t.Errorf("Name mismatch: got %v, want %v", proto.Name, user.Name)
	}
	if proto.Email != user.Email {
		t.Errorf("Email mismatch: got %v, want %v", proto.Email, user.Email)
	}
	if proto.Age != int32(user.Age) {
		t.Errorf("Age mismatch: got %v, want %v", proto.Age, user.Age)
	}
	if proto.CreatedAt != now.Format(time.RFC3339) {
		t.Errorf("CreatedAt mismatch: got %v, want %v", proto.CreatedAt, now.Format(time.RFC3339))
	}
}

// TestAdvancedMapper_MapOptionalExample covers optional nested values: a nil field
// stays nil whether the pair is mapped through a value method or a pointer method
func TestAdvancedMapper_MapOptionalExample(t *testing.T) {
	m := AdvancedMapperImpl{}

	empty := m.MapOptionalExample(&models.OptionalExample{ID: "opt-1"})
	if empty.ID != "opt-1" {
		t.Errorf("ID mismatch: got %v, want %v", empty.ID, "opt-1")
	}
	if empty.Detail != nil {
		t.Errorf("Detail should stay nil, got %+v", empty.Detail)
	}
	if empty.Item != nil {
		t.Errorf("Item should stay nil, got %+v", empty.Item)
	}

	full := m.MapOptionalExample(&models.OptionalExample{
		ID:     "opt-2",
		Detail: &models.Detail{Note: "note", Weight: 7},
		Item:   &models.Item{Name: "item", Count: 3},
	})
	if full.Detail == nil || full.Detail.Note != "note" || full.Detail.Weight != 7 {
		t.Errorf("Detail mismatch: got %+v", full.Detail)
	}
	if full.Item == nil || full.Item.Name != "item" || full.Item.Count != 3 {
		t.Errorf("Item mismatch: got %+v", full.Item)
	}
}

// TestAdvancedMapper_NilInput covers the guard every pointer-input method opens
// with, for a pointer result and for one paired with an error
func TestAdvancedMapper_NilInput(t *testing.T) {
	m := AdvancedMapperImpl{}

	if got := m.MapItem(nil); got != nil {
		t.Errorf("MapItem(nil) = %+v, want nil", got)
	}

	got, err := m.MapUserReverse(nil)
	if err != nil {
		t.Errorf("MapUserReverse(nil) error = %v, want nil", err)
	}
	if got != nil {
		t.Errorf("MapUserReverse(nil) = %+v, want nil", got)
	}
}

func TestAdvancedMapper_MapUserReverse(t *testing.T) {
	m := AdvancedMapperImpl{}
	now := time.Now()

	proto := &protos.UserProto{
		ID:        "user-1",
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: now.Format(time.RFC3339),
	}

	user, err := m.MapUserReverse(proto)
	if err != nil {
		t.Fatalf("MapUserReverse failed: %v", err)
	}

	if user.ID != proto.ID {
		t.Errorf("ID mismatch: got %v, want %v", user.ID, proto.ID)
	}
	if user.Name != proto.Name {
		t.Errorf("Name mismatch: got %v, want %v", user.Name, proto.Name)
	}
	if user.Email != proto.Email {
		t.Errorf("Email mismatch: got %v, want %v", user.Email, proto.Email)
	}
	if user.Age != int(proto.Age) {
		t.Errorf("Age mismatch: got %v, want %v", user.Age, proto.Age)
	}
	if user.CreatedAt.Unix() != now.Unix() {
		t.Errorf("CreatedAt mismatch: got %v, want %v", user.CreatedAt, now)
	}
}

func TestAdvancedMapper_MapUserWithRoles(t *testing.T) {
	m := AdvancedMapperImpl{}

	user := &models.UserWithRoles{
		ID:    "user-1",
		Name:  "John Doe",
		Roles: []string{"admin", "user"},
		Tags: map[string]string{
			"dept": "engineering",
			"loc":  "NYC",
		},
		Scores: []int{100, 200, 300},
		Metadata: map[string]int{
			"level": 5,
			"xp":    1000,
		},
	}

	proto := m.MapUserWithRoles(user)

	if proto.ID != user.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, user.ID)
	}
	if len(proto.Roles) != len(user.Roles) {
		t.Errorf("Roles length mismatch: got %v, want %v", len(proto.Roles), len(user.Roles))
	}
	if len(proto.Scores) != len(user.Scores) {
		t.Errorf("Scores length mismatch: got %v, want %v", len(proto.Scores), len(user.Scores))
	}
	for i, score := range proto.Scores {
		if score != int32(user.Scores[i]) {
			t.Errorf("Score[%d] mismatch: got %v, want %v", i, score, user.Scores[i])
		}
	}
	if len(proto.Metadata) != len(user.Metadata) {
		t.Errorf("Metadata length mismatch: got %v, want %v", len(proto.Metadata), len(user.Metadata))
	}
	for k, v := range user.Metadata {
		if proto.Metadata[k] != int64(v) {
			t.Errorf("Metadata[%s] mismatch: got %v, want %v", k, proto.Metadata[k], v)
		}
	}
}

func TestAdvancedMapper_MapArrayExample(t *testing.T) {
	m := AdvancedMapperImpl{}

	src := &models.ArrayExample{
		ID:   "array-1",
		Data: [4]byte{1, 2, 3, 4},
		Name: "Test Array",
	}

	proto := m.MapArrayExample(src)

	if proto.ID != src.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, src.ID)
	}
	if proto.Data != src.Data {
		t.Errorf("Data mismatch: got %v, want %v", proto.Data, src.Data)
	}
	if proto.Name != src.Name {
		t.Errorf("Name mismatch: got %v, want %v", proto.Name, src.Name)
	}
}

func TestAdvancedMapper_MapCustomConvert(t *testing.T) {
	m := AdvancedMapperImpl{}
	now := time.Now()

	src := &models.CustomConvertExample{
		ID:        "custom-1",
		Value:     "test value",
		Timestamp: now.Unix(),
	}

	proto := m.MapCustomConvert(src)

	if proto.ID != src.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, src.ID)
	}
	if proto.Value != src.Value {
		t.Errorf("Value mismatch: got %v, want %v", proto.Value, src.Value)
	}
	expectedTime := now.Format(time.RFC3339)
	if proto.Timestamp != expectedTime {
		t.Errorf("Timestamp mismatch: got %v, want %v", proto.Timestamp, expectedTime)
	}
}

func TestAdvancedMapper_MapCustomConvertReverse(t *testing.T) {
	m := AdvancedMapperImpl{}
	now := time.Now()

	proto := &protos.CustomConvertExampleProto{
		ID:        "custom-1",
		Value:     "test value",
		Timestamp: now.Format(time.RFC3339),
	}

	src, err := m.MapCustomConvertReverse(proto)
	if err != nil {
		t.Fatalf("MapCustomConvertReverse failed: %v", err)
	}

	if src.ID != proto.ID {
		t.Errorf("ID mismatch: got %v, want %v", src.ID, proto.ID)
	}
	if src.Value != proto.Value {
		t.Errorf("Value mismatch: got %v, want %v", src.Value, proto.Value)
	}
	if src.Timestamp != now.Unix() {
		t.Errorf("Timestamp mismatch: got %v, want %v", src.Timestamp, now.Unix())
	}
}

func TestAdvancedMapper_MapNamedSlice(t *testing.T) {
	m := AdvancedMapperImpl{}

	src := &models.NamedSliceExample{
		ID:    "named-1",
		Items: models.StringSlice{"item1", "item2", "item3"},
	}

	proto := m.MapNamedSlice(src)

	if proto.ID != src.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, src.ID)
	}
	if len(proto.Items) != len(src.Items) {
		t.Errorf("Items length mismatch: got %v, want %v", len(proto.Items), len(src.Items))
	}
	for i, item := range proto.Items {
		if item != string(src.Items[i]) {
			t.Errorf("Items[%d] mismatch: got %v, want %v", i, item, src.Items[i])
		}
	}
}

// TestAdvancedMapper_MapPointerSlice covers a repeated message field, where the
// elements stay pointers on both sides
func TestAdvancedMapper_MapPointerSlice(t *testing.T) {
	m := AdvancedMapperImpl{}

	src := &models.PointerSliceExample{
		ID: "list-1",
		Items: []*models.Item{
			{Name: "first", Count: 1},
			{Name: "second", Count: 2},
		},
	}

	proto := m.MapPointerSlice(src)

	if proto.ID != src.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, src.ID)
	}
	if len(proto.Items) != len(src.Items) {
		t.Fatalf("Items length mismatch: got %v, want %v", len(proto.Items), len(src.Items))
	}

	for i, item := range proto.Items {
		if item == nil {
			t.Fatalf("Items[%d] is nil", i)
		}
		if item.Name != src.Items[i].Name {
			t.Errorf("Items[%d].Name mismatch: got %v, want %v", i, item.Name, src.Items[i].Name)
		}
		if item.Count != int32(src.Items[i].Count) {
			t.Errorf("Items[%d].Count mismatch: got %v, want %v", i, item.Count, src.Items[i].Count)
		}
	}
}

// TestAdvancedMapper_MapSnapshotExample covers using: converters that take a named
// struct the generator has to resolve out of a separately loaded package
func TestAdvancedMapper_MapSnapshotExample(t *testing.T) {
	m := AdvancedMapperImpl{}

	src := &models.SnapshotExample{
		ID:       "snap-1",
		Snapshot: &models.Snapshot{ID: "inner", Tick: 42},
	}

	proto := m.MapSnapshotExample(src)

	if proto.ID != src.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, src.ID)
	}
	if want := converters.SummarizeSnapshot(src.Snapshot); proto.Summary != want {
		t.Errorf("Summary mismatch: got %v, want %v", proto.Summary, want)
	}
	if want := converters.DigestExample(src); proto.Digest != want {
		t.Errorf("Digest mismatch: got %v, want %v", proto.Digest, want)
	}
}

// TestAdvancedMapper_MapEvent covers conversions derived without a converter: a
// method call, string to []byte and an int onto a named integer type
func TestAdvancedMapper_MapEvent(t *testing.T) {
	m := AdvancedMapperImpl{}

	at := time.Date(2026, time.September, 17, 12, 0, 0, 500, time.UTC)
	src := &models.Event{ID: "ev-1", At: at, Owner: "root", OwnerID: 7, Level: 1}

	proto := m.MapEvent(src)

	if proto.ID != src.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, src.ID)
	}
	if proto.AtNs != at.UnixNano() {
		t.Errorf("AtNs mismatch: got %v, want %v", proto.AtNs, at.UnixNano())
	}
	if string(proto.Owner) != src.Owner {
		t.Errorf("Owner mismatch: got %q, want %q", proto.Owner, src.Owner)
	}
	if proto.OwnerId != src.OwnerID {
		t.Errorf("OwnerId mismatch: got %v, want %v", proto.OwnerId, src.OwnerID)
	}
	if proto.Level != protos.LevelWarn {
		t.Errorf("Level mismatch: got %v, want %v", proto.Level, protos.LevelWarn)
	}

	if proto.RetiredAtNs != 0 || proto.Deadline != nil || proto.Note != "" {
		t.Errorf("nil optional fields should stay zero, got %+v", proto)
	}

	note := "hello"
	full := m.MapEvent(&models.Event{ID: "ev-2", RetiredAt: &at, Deadline: &at, Note: &note})
	if full.RetiredAtNs != at.UnixNano() {
		t.Errorf("RetiredAtNs mismatch: got %v, want %v", full.RetiredAtNs, at.UnixNano())
	}
	if full.Deadline == nil || !full.Deadline.AsTime().Equal(at) {
		t.Errorf("Deadline mismatch: got %v, want %v", full.Deadline, at)
	}
	if full.Note != "HELLO" {
		t.Errorf("Note mismatch: got %q, want %q", full.Note, "HELLO")
	}
}

// TestAdvancedMapper_MapOrder covers a from: path through an edge pointer: with an
// error return the OrErr accessor reports an unloaded edge, without one the field
// stays zero
func TestAdvancedMapper_MapOrder(t *testing.T) {
	m := AdvancedMapperImpl{}

	unloaded := &models.Order{ID: "ord-1"}
	loaded := &models.Order{ID: "ord-2", Edges: models.OrderEdges{Customer: &models.Customer{Name: "acme"}}}

	if _, err := m.MapOrder(unloaded); err == nil {
		t.Error("MapOrder should report an unloaded edge")
	}

	proto, err := m.MapOrder(loaded)
	if err != nil {
		t.Fatalf("MapOrder() error = %v", err)
	}
	if proto.ID != "ord-2" || proto.CustomerName != "acme" {
		t.Errorf("MapOrder mismatch: got %+v", proto)
	}

	if loose := m.MapOrderLoose(unloaded); loose.ID != "ord-1" || loose.CustomerName != "" {
		t.Errorf("MapOrderLoose on an unloaded edge: got %+v", loose)
	}
	if loose := m.MapOrderLoose(loaded); loose.CustomerName != "acme" {
		t.Errorf("MapOrderLoose mismatch: got %+v", loose)
	}
}

// TestAdvancedMapper_FillEvent covers a builder target: one setter call per column,
// the nillable setter for a converter that hands back a pointer, the id form of an
// edge, and an untouched builder for a nil source
func TestAdvancedMapper_FillEvent(t *testing.T) {
	m := AdvancedMapperImpl{}

	at := time.Date(2026, time.September, 17, 12, 0, 0, 500, time.UTC)
	src := &protos.EventProto{ID: "ev-1", AtNs: at.UnixNano(), Owner: []byte("root"), OwnerId: 7, Level: protos.LevelWarn, Note: "hi"}

	got := m.FillEvent(src, models.NewEventCreate()).Values
	if got["id"] != "ev-1" || got["owner"] != "root" || got["level"] != 1 || got["note"] != "hi" || got["customer_id"] != int64(7) {
		t.Errorf("columns mismatch: got %v", got)
	}
	if set, ok := got["at"].(time.Time); !ok || !set.Equal(at) {
		t.Errorf("at mismatch: got %v, want %v", got["at"], at)
	}
	if _, ok := got["retired_at"]; ok {
		t.Errorf("retired_at should not be set for a zero source, got %v", got["retired_at"])
	}
	if _, ok := got["customer"]; ok {
		t.Errorf("the edge should be set by id only, got %v", got["customer"])
	}

	src.RetiredAtNs = at.UnixNano()
	if set, ok := m.FillEvent(src, models.NewEventCreate()).Values["retired_at"].(time.Time); !ok || !set.Equal(at) {
		t.Errorf("retired_at mismatch: got %v, want %v", set, at)
	}

	if untouched := m.FillEvent(nil, models.NewEventCreate()).Values; len(untouched) != 0 {
		t.Errorf("a nil source should leave the builder untouched, got %v", untouched)
	}
}

// TestAdvancedMapper_FillEventSparse covers mapper:skip-optional: optional setters
// without a source field of their name stay unset, the others are still called
func TestAdvancedMapper_FillEventSparse(t *testing.T) {
	m := AdvancedMapperImpl{}

	src := &protos.EventProto{ID: "ev-1", RetiredAtNs: 42, Note: "hi"}
	got := m.FillEventSparse(src, models.NewEventCreate()).Values

	if got["note"] != "hi" || got["id"] != "ev-1" {
		t.Errorf("columns mismatch: got %v", got)
	}
	for _, unset := range []string{"retired_at", "deleted_at"} {
		if v, ok := got[unset]; ok {
			t.Errorf("%s should stay unset, got %v", unset, v)
		}
	}
}

// TestAdvancedMapper_FillEventDraft covers a pointer source field of the setter's
// own name: it goes through the nillable setter, so nil leaves the column unset
func TestAdvancedMapper_FillEventDraft(t *testing.T) {
	m := AdvancedMapperImpl{}

	got := m.FillEventDraft(&protos.EventDraftProto{ID: "ev-1"}, models.NewEventCreate()).Values
	if v, ok := got["note"]; ok {
		t.Errorf("note should stay unset for a nil source, got %v", v)
	}

	note := "hi"
	got = m.FillEventDraft(&protos.EventDraftProto{ID: "ev-1", Note: &note}, models.NewEventCreate()).Values
	if got["note"] != "hi" || got["id"] != "ev-1" {
		t.Errorf("columns mismatch: got %v", got)
	}
}

// TestAdvancedMapper_MapTask covers enums paired by constant name across differing
// underlying types, in both directions
func TestAdvancedMapper_MapTask(t *testing.T) {
	m := AdvancedMapperImpl{}

	proto, err := m.MapTask(&models.Task{ID: "t-1", Phase: models.PhaseDone})
	if err != nil {
		t.Fatalf("MapTask() error = %v", err)
	}
	if proto.Phase != protos.TaskPhase_TASK_PHASE_DONE {
		t.Errorf("Phase mismatch: got %v, want %v", proto.Phase, protos.TaskPhase_TASK_PHASE_DONE)
	}

	if _, err := m.MapTask(&models.Task{Phase: "paused"}); err == nil {
		t.Error("MapTask should reject a value Phase does not declare")
	}

	model, err := m.MapTaskReverse(&protos.TaskProto{ID: "t-2", Phase: protos.TaskPhase_TASK_PHASE_RUNNING})
	if err != nil {
		t.Fatalf("MapTaskReverse() error = %v", err)
	}
	if model.Phase != models.PhaseRunning {
		t.Errorf("Phase mismatch: got %v, want %v", model.Phase, models.PhaseRunning)
	}

	if _, err := m.MapTaskReverse(&protos.TaskProto{Phase: protos.TaskPhase_TASK_PHASE_UNSPECIFIED}); err == nil {
		t.Error("MapTaskReverse should reject a value without a counterpart")
	}
}

// TestAdvancedMapper_MapArrayConvert covers a fixed-size array whose elements need
// a conversion, so the array is rebuilt element by element
func TestAdvancedMapper_MapArrayConvert(t *testing.T) {
	m := AdvancedMapperImpl{}

	src := &models.ArrayConvertExample{ID: "arr-1", Values: [3]int{1, 2, 3}}

	proto := m.MapArrayConvert(src)

	if proto.ID != src.ID {
		t.Errorf("ID mismatch: got %v, want %v", proto.ID, src.ID)
	}
	if want := [3]int32{1, 2, 3}; proto.Values != want {
		t.Errorf("Values mismatch: got %v, want %v", proto.Values, want)
	}
}
