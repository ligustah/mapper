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
