package handlers

import (
	"encoding/json"
	"memory-game/models"
	"net/http"
	"sort"
)

// GetLeaderboard handles GET /api/leaderboard?difficulty=all|easy|medium|hard
func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	difficulty := r.URL.Query().Get("difficulty")
	if difficulty == "" {
		difficulty = "all"
	}

	mu.Lock()
	var filtered []models.Score
	for _, score := range scores {
		if difficulty == "all" || score.Difficulty == difficulty {
			filtered = append(filtered, score)
		}
	}
	mu.Unlock()

	// Sort by score descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	// Return top 10
	if len(filtered) > 10 {
		filtered = filtered[:10]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}