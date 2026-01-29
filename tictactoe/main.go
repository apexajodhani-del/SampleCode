package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type Player struct {
	ID       string          `json:"id"`
	Conn     *websocket.Conn `json:"-"`
	Symbol   string          `json:"symbol"`
	RoomID   string          `json:"roomId"`
	IsReady  bool            `json:"isReady"`
}

type Room struct {
	ID      string    `json:"id"`
	Players []*Player `json:"players"`
	Board   [3][3]string `json:"board"`
	Turn    string    `json:"turn"`
	Status  string    `json:"status"` // "waiting", "playing", "finished"
	Winner  string    `json:"winner"`
	mutex   sync.Mutex
}

type Message struct {
	Type      string      `json:"type"`
	PlayerID  string      `json:"playerId,omitempty"`
	RoomID    string      `json:"roomId,omitempty"`
	Symbol    string      `json:"symbol,omitempty"`
	Row       int         `json:"row,omitempty"`
	Col       int         `json:"col,omitempty"`
	Board     [3][3]string `json:"board,omitempty"`
	Turn      string      `json:"turn,omitempty"`
	Status    string      `json:"status,omitempty"`
	Winner    string      `json:"winner,omitempty"`
	Error     string      `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
}

var (
	rooms    = make(map[string]*Room)
	players  = make(map[string]*Player)
	waitingQueue []*Player
	roomMutex sync.Mutex
	playerMutex sync.Mutex
	roomCounter int
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/", serveStatic)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "index.html")
	} else {
		http.ServeFile(w, r, r.URL.Path[1:])
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	playerID := generatePlayerID()
	player := &Player{
		ID:      playerID,
		Conn:    conn,
		Symbol:  "",
		RoomID:  "",
		IsReady: false,
	}

	playerMutex.Lock()
	players[playerID] = player
	playerMutex.Unlock()

	log.Printf("Player %s connected", playerID)

	// Send welcome message
	sendMessage(conn, Message{
		Type:    "connected",
		PlayerID: playerID,
		Message: "Connected to server. Waiting for opponent...",
	})

	// Try to match player
	matchPlayer(player)

	// Handle incoming messages
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read error for player %s: %v", playerID, err)
			handleDisconnect(player)
			break
		}

		handleMessage(player, msg)
	}
}

func matchPlayer(player *Player) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	if len(waitingQueue) == 0 {
		// No one waiting, add to queue
		waitingQueue = append(waitingQueue, player)
		sendMessage(player.Conn, Message{
			Type:    "waiting",
			Message: "Waiting for another player...",
		})
		log.Printf("Player %s added to waiting queue", player.ID)
		return
	}

	// Match with waiting player
	opponent := waitingQueue[0]
	waitingQueue = waitingQueue[1:]

	// Create new room
	roomCounter++
	roomID := generateRoomID(roomCounter)
	room := &Room{
		ID:      roomID,
		Players: []*Player{opponent, player},
		Board:   [3][3]string{{"", "", ""}, {"", "", ""}, {"", "", ""}},
		Turn:    "X",
		Status:  "playing",
		Winner:  "",
	}

	// Assign symbols
	opponent.Symbol = "X"
	opponent.RoomID = roomID
	opponent.IsReady = true
	player.Symbol = "O"
	player.RoomID = roomID
	player.IsReady = true

	rooms[roomID] = room

	log.Printf("Room %s created with players %s (X) and %s (O)", roomID, opponent.ID, player.ID)

	// Notify both players
	sendMessage(opponent.Conn, Message{
		Type:    "matched",
		PlayerID: opponent.ID,
		RoomID:  roomID,
		Symbol:  "X",
		Board:   room.Board,
		Turn:    room.Turn,
		Status:  room.Status,
		Message: "Game started! You are X. Your turn.",
	})

	sendMessage(player.Conn, Message{
		Type:    "matched",
		PlayerID: player.ID,
		RoomID:  roomID,
		Symbol:  "O",
		Board:   room.Board,
		Turn:    room.Turn,
		Status:  room.Status,
		Message: "Game started! You are O. Waiting for X...",
	})
}

func handleMessage(player *Player, msg Message) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := rooms[player.RoomID]
	if !exists {
		sendMessage(player.Conn, Message{
			Type:  "error",
			Error: "Room not found",
		})
		return
	}

	switch msg.Type {
	case "move":
		handleMove(player, room, msg)
	}
}

func handleMove(player *Player, room *Room, msg Message) {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	// Validate it's player's turn
	if room.Turn != player.Symbol {
		sendMessage(player.Conn, Message{
			Type:  "error",
			Error: "Not your turn",
		})
		return
	}

	// Validate move coordinates
	if msg.Row < 0 || msg.Row > 2 || msg.Col < 0 || msg.Col > 2 {
		sendMessage(player.Conn, Message{
			Type:  "error",
			Error: "Invalid coordinates",
		})
		return
	}

	// Validate cell is empty
	if room.Board[msg.Row][msg.Col] != "" {
		sendMessage(player.Conn, Message{
			Type:  "error",
			Error: "Cell already occupied",
		})
		return
	}

	// Make move
	room.Board[msg.Row][msg.Col] = player.Symbol

	// Check for win or draw
	winner := checkWinner(room.Board)
	if winner != "" {
		room.Status = "finished"
		room.Winner = winner
		room.Turn = ""
	} else if isBoardFull(room.Board) {
		room.Status = "finished"
		room.Winner = "draw"
		room.Turn = ""
	} else {
		// Switch turn
		if room.Turn == "X" {
			room.Turn = "O"
		} else {
			room.Turn = "X"
		}
	}

	// Broadcast update to both players
	for _, p := range room.Players {
		statusMsg := room.Status
		if room.Status == "finished" {
			if room.Winner == "draw" {
				statusMsg = "Game ended in a draw!"
			} else if room.Winner == p.Symbol {
				statusMsg = "You won!"
			} else {
				statusMsg = "You lost!"
			}
		} else if room.Turn == p.Symbol {
			statusMsg = "Your turn"
		} else {
			statusMsg = "Opponent's turn"
		}

		sendMessage(p.Conn, Message{
			Type:    "update",
			Board:   room.Board,
			Turn:    room.Turn,
			Status:  statusMsg,
			Winner:  room.Winner,
			Message: statusMsg,
		})
	}
}

func checkWinner(board [3][3]string) string {
	// Check rows
	for i := 0; i < 3; i++ {
		if board[i][0] != "" && board[i][0] == board[i][1] && board[i][1] == board[i][2] {
			return board[i][0]
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if board[0][i] != "" && board[0][i] == board[1][i] && board[1][i] == board[2][i] {
			return board[0][i]
		}
	}

	// Check diagonals
	if board[0][0] != "" && board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return board[0][0]
	}
	if board[0][2] != "" && board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return board[0][2]
	}

	return ""
}

func isBoardFull(board [3][3]string) bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == "" {
				return false
			}
		}
	}
	return true
}

func handleDisconnect(player *Player) {
	roomMutex.Lock()
	defer roomMutex.Unlock()

	playerMutex.Lock()
	delete(players, player.ID)
	playerMutex.Unlock()

	// Remove from waiting queue if present
	for i, p := range waitingQueue {
		if p.ID == player.ID {
			waitingQueue = append(waitingQueue[:i], waitingQueue[i+1:]...)
			break
		}
	}

	// Handle room cleanup
	if player.RoomID != "" {
		room, exists := rooms[player.RoomID]
		if exists {
			// Notify opponent
			for _, p := range room.Players {
				if p.ID != player.ID {
					sendMessage(p.Conn, Message{
						Type:    "opponent_disconnected",
						Message: "Opponent disconnected",
					})
				}
			}
			delete(rooms, player.RoomID)
		}
	}

	log.Printf("Player %s disconnected", player.ID)
}

func sendMessage(conn *websocket.Conn, msg Message) {
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Write error: %v", err)
	}
}

func generatePlayerID() string {
	return "player_" + randomString(8)
}

func generateRoomID(counter int) string {
	return "room_" + randomString(6)
}

func randomString(length int) string {
	// Generate enough bytes for base64 encoding
	bytesNeeded := (length * 3) / 4
	if bytesNeeded < 1 {
		bytesNeeded = 1
	}
	b := make([]byte, bytesNeeded)
	rand.Read(b)
	encoded := base64.URLEncoding.EncodeToString(b)
	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}

