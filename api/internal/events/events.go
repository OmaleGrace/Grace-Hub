package events

import "time"

type Event struct {
	ID         int64
	Category   string
	Title      string
	Details    []byte
	OpensAt    time.Time
	LocksAt    time.Time
	ResolvesAt time.Time
	Status     string
	Outcome    *string
	ResolvedAt *time.Time
	CreatedAt  time.Time
}    