// Package converters holds converters referenced from using: directives. It sits
// beside the mapper package rather than inside it, so the generator has to resolve
// a type it already knows from a package it loads separately.
package converters

import (
	"fmt"

	"github.com/ligustah/mapper/example/models"
)

// SummarizeSnapshot renders a snapshot as a single line
func SummarizeSnapshot(s *models.Snapshot) string {
	if s == nil {
		return ""
	}

	return fmt.Sprintf("%s@%d", s.ID, s.Tick)
}

// DigestExample renders a whole example as a single line
func DigestExample(e *models.SnapshotExample) string {
	if e == nil {
		return ""
	}

	return e.ID + ":" + SummarizeSnapshot(e.Snapshot)
}
