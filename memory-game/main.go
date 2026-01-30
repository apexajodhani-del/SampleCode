package main

import (
	"fmt"
	"memory-game/handlers"
	"net/http"
)

func main() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// API routes
	http.HandleFunc("/api/game/result", handlers.SubmitResult)
	http.HandleFunc("/api/leaderboard", handlers.GetLeaderboard)
	http.HandleFunc("/api/health", handlers.HealthCheck)

	// Serve index.html for root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	fmt.Println("Server starting on :8081")
	http.ListenAndServe(":8081", nil)
}