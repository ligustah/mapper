package models

import (
	"time"

	"github.com/volatiletech/null"
)

// Job is an object representing the database table.
type Job struct {
	ID         string      `boil:"id" json:"id" toml:"id" yaml:"id"`
	CreatedAt  time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt  null.Time   `boil:"updated_at" json:"updated_at,omitempty" toml:"updated_at" yaml:"updated_at,omitempty"`
	RunAt      time.Time   `boil:"run_at" json:"run_at" toml:"run_at" yaml:"run_at"`
	Group      null.String `boil:"group" json:"group,omitempty" toml:"group" yaml:"group,omitempty"`
	Queue      string      `boil:"queue" json:"queue" toml:"queue" yaml:"queue"`
	DataType   string      `boil:"data_type" json:"data_type" toml:"data_type" yaml:"data_type"`
	Data       []byte      `boil:"data" json:"data" toml:"data" yaml:"data"`
	ExternalID null.String `boil:"external_id" json:"external_id,omitempty" toml:"external_id" yaml:"external_id,omitempty"`
	CronID     null.String `boil:"cron_id" json:"cron_id,omitempty" toml:"cron_id" yaml:"cron_id,omitempty"`

	R *jobR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L jobL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// jobR is where relationships are stored.
type jobR struct {
	Cron        *Cron
	Transitions TransitionSlice
}

type jobL struct{}

// Cron is an object representing the database table.
type Cron struct {
	ID            string    `boil:"id" json:"id" toml:"id" yaml:"id"`
	CreatedAt     time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt     null.Time `boil:"updated_at" json:"updated_at,omitempty" toml:"updated_at" yaml:"updated_at,omitempty"`
	Schedule      string    `boil:"schedule" json:"schedule" toml:"schedule" yaml:"schedule"`
	DataType      string    `boil:"data_type" json:"data_type" toml:"data_type" yaml:"data_type"`
	Data          []byte    `boil:"data" json:"data" toml:"data" yaml:"data"`
	Queue         string    `boil:"queue" json:"queue" toml:"queue" yaml:"queue"`
	SkipExecution bool      `boil:"skip_execution" json:"skip_execution" toml:"skip_execution" yaml:"skip_execution"`
	StartTime     time.Time `boil:"start_time" json:"start_time" toml:"start_time" yaml:"start_time"`

	R *cronR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L cronL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// cronL is where Load methods for each relationship are stored.
type cronL struct{}

type JobSlice []*Job
type TransitionSlice []*Transition

// cronR is where relationships are stored.
type cronR struct {
	Jobs JobSlice
}

// Transition is an object representing the database table.
type Transition struct {
	ID        int64     `boil:"id" json:"id" toml:"id" yaml:"id"`
	JobID     string    `boil:"job_id" json:"job_id" toml:"job_id" yaml:"job_id"`
	CreatedAt time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	State     string    `boil:"state" json:"state" toml:"state" yaml:"state"`
	Attempts  int       `boil:"attempts" json:"attempts" toml:"attempts" yaml:"attempts"`
	RetryAt   null.Time `boil:"retry_at" json:"retry_at,omitempty" toml:"retry_at" yaml:"retry_at,omitempty"`

	R *transitionR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L transitionL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

// transitionR is where relationships are stored.
type transitionR struct {
	Job *Job
}

// transitionL is where Load methods for each relationship are stored.
type transitionL struct{}
