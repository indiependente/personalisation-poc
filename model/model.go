package model

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	Tags      []string  `json:"tags"`
	Segments  []Segment `json:"segments"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Segment struct {
	Type          string     `json:"type"`
	Categories    []Category `json:"categories"`
	TopCategories []string   `json:"top_categories"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ExpiresAt     time.Time  `json:"expires_at"`
}

type Category struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"`
}
