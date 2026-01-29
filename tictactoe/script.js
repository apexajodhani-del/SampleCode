class TicTacToeClient {
    constructor() {
        this.ws = null;
        this.playerId = null;
        this.playerSymbol = null;
        this.roomId = null;
        this.currentTurn = null;
        this.gameStatus = 'connecting';
        this.board = [
            ['', '', ''],
            ['', '', ''],
            ['', '', '']
        ];
        
        this.init();
    }

    init() {
        this.connect();
        this.setupEventListeners();
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        console.log('Connecting to:', wsUrl);
        this.updateConnectionStatus('Connecting...', false);
        
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.updateConnectionStatus('Connected', true);
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateConnectionStatus('Connection error', false);
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.updateConnectionStatus('Disconnected', false);
            this.updateStatus('Connection lost. Reconnecting...');
            
            // Attempt to reconnect after 3 seconds
            setTimeout(() => {
                if (this.gameStatus !== 'finished') {
                    this.connect();
                }
            }, 3000);
        };
    }

    handleMessage(message) {
        console.log('Received message:', message);

        switch (message.type) {
            case 'connected':
                this.playerId = message.playerId;
                this.updateStatus(message.message || 'Connected. Waiting for opponent...');
                break;

            case 'waiting':
                this.updateStatus(message.message || 'Waiting for another player...');
                break;

            case 'matched':
                this.playerSymbol = message.symbol;
                this.roomId = message.roomId;
                this.currentTurn = message.turn;
                this.board = message.board;
                this.gameStatus = 'playing';
                
                this.updateStatus(message.message || 'Game started!');
                this.updatePlayerInfo();
                this.updateBoard();
                this.updateGameControls();
                break;

            case 'update':
                this.board = message.board;
                this.currentTurn = message.turn;
                this.updateStatus(message.message || message.status);
                this.updateBoard();
                
                if (message.winner) {
                    this.gameStatus = 'finished';
                    this.handleGameEnd(message.winner);
                }
                break;

            case 'error':
                this.updateStatus(`Error: ${message.error}`, 'error');
                break;

            case 'opponent_disconnected':
                this.updateStatus(message.message || 'Opponent disconnected');
                this.gameStatus = 'waiting';
                break;
        }
    }

    setupEventListeners() {
        const cells = document.querySelectorAll('.cell');
        cells.forEach(cell => {
            cell.addEventListener('click', () => {
                const row = parseInt(cell.dataset.row);
                const col = parseInt(cell.dataset.col);
                this.makeMove(row, col);
            });
        });

        const newGameBtn = document.getElementById('newGameBtn');
        newGameBtn.addEventListener('click', () => {
            this.startNewGame();
        });
    }

    makeMove(row, col) {
        if (this.gameStatus !== 'playing') {
            return;
        }

        if (this.currentTurn !== this.playerSymbol) {
            this.updateStatus('Not your turn!', 'error');
            return;
        }

        if (this.board[row][col] !== '') {
            return;
        }

        const message = {
            type: 'move',
            playerId: this.playerId,
            roomId: this.roomId,
            row: row,
            col: col
        };

        this.ws.send(JSON.stringify(message));
    }

    updateBoard() {
        const cells = document.querySelectorAll('.cell');
        
        cells.forEach(cell => {
            const row = parseInt(cell.dataset.row);
            const col = parseInt(cell.dataset.col);
            const value = this.board[row][col];
            
            cell.textContent = value;
            cell.className = 'cell';
            
            if (value === 'X') {
                cell.classList.add('x');
            } else if (value === 'O') {
                cell.classList.add('o');
            }

            // Disable cells if not player's turn or game finished
            if (this.gameStatus !== 'playing' || this.currentTurn !== this.playerSymbol || value !== '') {
                cell.classList.add('disabled');
            } else {
                cell.classList.remove('disabled');
            }
        });
    }

    updateStatus(text, type = 'info') {
        const statusText = document.getElementById('statusText');
        const statusCard = document.getElementById('statusCard');
        
        statusText.textContent = text;
        
        // Update status card color based on type
        statusCard.className = 'status-card';
        if (type === 'error') {
            statusCard.style.background = 'linear-gradient(135deg, #e74c3c 0%, #c0392b 100%)';
        } else if (type === 'success') {
            statusCard.style.background = 'linear-gradient(135deg, #2ecc71 0%, #27ae60 100%)';
        } else {
            statusCard.style.background = 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)';
        }
    }

    updatePlayerInfo() {
        const playerSymbolEl = document.getElementById('playerSymbol');
        const roomIdEl = document.getElementById('roomId');
        const roomInfo = document.getElementById('roomInfo');
        
        playerSymbolEl.textContent = this.playerSymbol;
        roomIdEl.textContent = this.roomId;
        roomInfo.style.display = 'block';
    }

    updateConnectionStatus(text, connected) {
        const connectionText = document.getElementById('connectionText');
        const statusIndicator = document.getElementById('statusIndicator');
        
        connectionText.textContent = text;
        statusIndicator.className = 'status-indicator';
        
        if (connected) {
            statusIndicator.classList.add('connected');
        } else {
            statusIndicator.classList.add('disconnected');
        }
    }

    updateGameControls() {
        const newGameBtn = document.getElementById('newGameBtn');
        if (this.gameStatus === 'finished') {
            newGameBtn.style.display = 'block';
        } else {
            newGameBtn.style.display = 'none';
        }
    }

    handleGameEnd(winner) {
        this.updateGameControls();
        
        if (winner === 'draw') {
            this.updateStatus('Game ended in a draw!', 'info');
        } else if (winner === this.playerSymbol) {
            this.updateStatus('🎉 You won!', 'success');
        } else {
            this.updateStatus('You lost. Better luck next time!', 'error');
        }
    }

    startNewGame() {
        // Reset game state
        this.playerSymbol = null;
        this.roomId = null;
        this.currentTurn = null;
        this.gameStatus = 'connecting';
        this.board = [
            ['', '', ''],
            ['', '', ''],
            ['', '', '']
        ];
        
        // Hide room info
        document.getElementById('roomInfo').style.display = 'none';
        document.getElementById('newGameBtn').style.display = 'none';
        
        // Reconnect to get matched with a new opponent
        if (this.ws) {
            this.ws.close();
        }
        this.connect();
    }
}

// Initialize the game when the page loads
window.addEventListener('DOMContentLoaded', () => {
    new TicTacToeClient();
});

