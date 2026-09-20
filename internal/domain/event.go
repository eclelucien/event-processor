package domain

import "time"

type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
