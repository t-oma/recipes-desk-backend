package dto

import "time"

// Tag represents a tag output DTO.
type Tag struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
