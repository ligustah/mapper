package models

import "time"

// EventCreate mimics the builder ent generates for an entity: a setter per column,
// a nillable twin for each optional column, both forms of an edge setter and a
// variadic adder for a to-many edge. Values records what was set.
type EventCreate struct {
	Values map[string]any
}

// NewEventCreate starts an empty builder
func NewEventCreate() *EventCreate {
	return &EventCreate{Values: map[string]any{}}
}

func (c *EventCreate) set(name string, v any) *EventCreate {
	c.Values[name] = v
	return c
}

// SetID sets the "id" column
func (c *EventCreate) SetID(v string) *EventCreate { return c.set("id", v) }

// SetAt sets the "at" column
func (c *EventCreate) SetAt(v time.Time) *EventCreate { return c.set("at", v) }

// SetOwner sets the "owner" column
func (c *EventCreate) SetOwner(v string) *EventCreate { return c.set("owner", v) }

// SetLevel sets the "level" column
func (c *EventCreate) SetLevel(v int) *EventCreate { return c.set("level", v) }

// SetNote sets the optional "note" column
func (c *EventCreate) SetNote(v string) *EventCreate { return c.set("note", v) }

// SetNillableNote sets the "note" column if v is not nil
func (c *EventCreate) SetNillableNote(v *string) *EventCreate {
	if v != nil {
		c.SetNote(*v)
	}
	return c
}

// SetRetiredAt sets the optional "retired_at" column
func (c *EventCreate) SetRetiredAt(v time.Time) *EventCreate { return c.set("retired_at", v) }

// SetNillableRetiredAt sets the "retired_at" column if v is not nil
func (c *EventCreate) SetNillableRetiredAt(v *time.Time) *EventCreate {
	if v != nil {
		c.SetRetiredAt(*v)
	}
	return c
}

// SetDeletedAt sets the optional "deleted_at" column
func (c *EventCreate) SetDeletedAt(v time.Time) *EventCreate { return c.set("deleted_at", v) }

// SetNillableDeletedAt sets the "deleted_at" column if v is not nil
func (c *EventCreate) SetNillableDeletedAt(v *time.Time) *EventCreate {
	if v != nil {
		c.SetDeletedAt(*v)
	}
	return c
}

// SetCustomerID sets the "customer" edge by id
func (c *EventCreate) SetCustomerID(v int64) *EventCreate { return c.set("customer_id", v) }

// SetCustomer sets the "customer" edge to an entity
func (c *EventCreate) SetCustomer(v *Customer) *EventCreate { return c.set("customer", v) }

// AddTagIDs adds to the "tags" edge by ids
func (c *EventCreate) AddTagIDs(ids ...int) *EventCreate { return c.set("tag_ids", ids) }
