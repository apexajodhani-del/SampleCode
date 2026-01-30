# Memory Matching Game

A complete memory matching game built with Go backend and HTML/CSS/JS frontend, featuring smooth UI, animations, scoring, and leaderboard.

## Features

- Three difficulty levels: Easy (4x4), Medium (4x5), Hard (5x6)
- Timer (MM:SS format), moves counter, score counter
- Gradient UI with card flip animations
- Leaderboard with filters (All, Easy, Medium, Hard)
- Scoring system with bonuses and penalties

## Folder Structure

```
memory-game/
├── main.go
├── go.mod
├── handlers/
│   ├── leaderboard.go
│   └── game.go
├── models/
│   └── score.go
├── static/
│   ├── index.html
│   ├── style.css
│   ├── app.js
│   └── assets/
│       └── icons/
└── README.md
```

## Run Instructions

1. Navigate to the memory-game directory:
   ```
   cd memory-game
   ```

2. Run the Go server:
   ```
   go run main.go
   ```

3. Open your browser and navigate to `http://localhost:8080`

4. Click "NEW GAME" to start (defaults to Easy).

## Game Rules

- Click on cards (showing ?) to flip them.
- Only 2 cards open at a time.
- Find matching emoji pairs.
- Matched cards stay open with glow.
- Complete all pairs to win.
- Score: +10 per match, bonus for speed, penalty for wrong moves.
- Submit score to leaderboard after winning.