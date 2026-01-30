package models

// Score represents a game score entry
type Score struct {
	Name       string `json:"name"`
	Difficulty string `json:"difficulty"` // "easy", "medium", "hard"
	Time       int    `json:"time"`       // in seconds
	Moves      int    `json:"moves"`
	Score      int    `json:"score"`
}