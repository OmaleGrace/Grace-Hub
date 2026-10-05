package events

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID         int64           `json:"id"`
	Category   string          `json:"category"`
	Title      string          `json:"title"`
	Details    json.RawMessage `json:"details"`
	Options    []string        `json:"options"`
	OpensAt    time.Time       `json:"opens_at"`
	LocksAt    time.Time       `json:"locks_at"`
	ResolvesAt time.Time       `json:"resolves_at"`
	Status     string          `json:"status"`
	Outcome    *string         `json:"outcome"`
	ResolvedAt *time.Time      `json:"resolved_at"`
	CreatedAt  time.Time       `json:"created_at"`
}