package handlers

import (
	"encoding/json"
	"memory-game/models"
	"net/http"
	"sync"
)

var scores []models.Score
var mu sync.Mutex

// SubmitResult handles POST /api/game/result
func SubmitResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var score models.Score
	if err := json.NewDecoder(r.Body).Decode(&score); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mu.Lock()
	scores = append(scores, score)
	mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

// HealthCheck handles GET /api/health
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}