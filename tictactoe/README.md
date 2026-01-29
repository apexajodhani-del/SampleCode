# Real-Time Multiplayer Tic-Tac-Toe

A web-based real-time multiplayer Tic-Tac-Toe game built with Golang and WebSockets.

## Features

- Real-time multiplayer gameplay via WebSockets
- Automatic player matching into private rooms
- Turn-based game with validation
- Win/draw detection
- Responsive, modern UI
- Automatic reconnection on disconnect

## Requirements

- Go 1.21 or higher
- Google Chrome (or any modern browser)

## Setup

1. Install dependencies:
```bash
go mod tidy
```

2. Run the server:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

## Usage

1. Open your browser and navigate to `http://localhost:8080`
2. The app will automatically connect to the WebSocket server
3. Wait for another player to join (open a second browser tab/window)
4. Once matched, players are assigned X and O automatically
5. Take turns clicking on the board to make moves
6. The game detects wins and draws automatically

## Testing

To test with two players:
1. Open `http://localhost:8080` in one browser tab
2. Open `http://localhost:8080` in another browser tab (or incognito window)
3. Both players will be automatically matched and can play!

## Architecture

- **Backend**: Go server with WebSocket support using `gorilla/websocket`
- **Frontend**: Vanilla JavaScript with WebSocket API
- **Communication**: JSON messages over WebSocket protocol

