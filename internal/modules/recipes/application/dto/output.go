package dto

import "time"

type Recipe struct {
	ID          string
	Title       string
	Description string
	Ingredients []Ingredient
	Steps       []Step
	CookingTime int64
	Portions    int
	Tags        []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Ingredient struct {
	Name   string
	Amount float64
	Unit   string
}

type Step struct {
	Order       int
	Description string
	Duration    time.Duration
}
