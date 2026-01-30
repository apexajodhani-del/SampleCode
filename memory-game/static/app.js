// Emojis for cards
const emojis = ['🎮', '🎵', '🎨', '🚀', '🐶', '🍎', '⚽', '🌟', '🍕', '🎂', '🌈', '🐱', '🎸', '🚲', '🍦', '🦄', '🍔', '🎃', '🌺', '🐸', '🍇', '⚡', '🔥', '🌙', '💎', '🎈', '🔔', '🎁', '🎊', '🍭', '🍪', '🥤'];

const difficulties = {
    easy: { rows: 4, cols: 4, pairs: 8, timeLimit: 300 }, // 5 min
    medium: { rows: 6, cols: 6, pairs: 18, timeLimit: 240 }, // 4 min
    hard: { rows: 8, cols: 8, pairs: 32, timeLimit: 180 } // 3 min
};

let currentDifficulty = null;
let board = [];
let flippedCards = [];
let matchedPairs = 0;
let moves = 0;
let score = 0;
let timer = 0;
let interval = null;
let startTime = null;

document.getElementById('easy-btn').addEventListener('click', () => selectDifficulty('easy'));
document.getElementById('medium-btn').addEventListener('click', () => selectDifficulty('medium'));
document.getElementById('hard-btn').addEventListener('click', () => selectDifficulty('hard'));
document.getElementById('new-game').addEventListener('click', () => {
    document.getElementById('game-area').style.display = 'none';
    document.getElementById('difficulty-selection').style.display = 'flex';
    document.getElementById('win-modal').style.display = 'none';
});
document.getElementById('submit-score').addEventListener('click', submitScore);
document.getElementById('play-again').addEventListener('click', () => location.reload());
document.querySelectorAll('.filter').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.filter').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        loadLeaderboard(btn.dataset.difficulty);
    });
});

// Load leaderboard on start
loadLeaderboard('all');

function selectDifficulty(diff) {
    document.getElementById('difficulty-selection').style.display = 'none';
    document.getElementById('game-area').style.display = 'flex';
    startGame(diff);
}

function startGame(diff) {
    currentDifficulty = diff;
    const { rows, cols, pairs } = difficulties[diff];
    board = generateBoard(pairs);
    renderBoard(rows, cols);
    flippedCards = [];
    matchedPairs = 0;
    moves = 0;
    score = 0;
    timer = 0;
    startTime = null;
    updateStats();
    clearInterval(interval);
    interval = setInterval(() => {
        if (startTime) {
            timer++;
            updateStats();
            if (timer >= difficulties[currentDifficulty].timeLimit) {
                // Time out, but for simplicity, just continue
            }
        }
    }, 1000);
    loadLeaderboard('all');
}

function generateBoard(pairs) {
    const symbols = [];
    for (let i = 0; i < pairs; i++) {
        symbols.push(emojis[i], emojis[i]);
    }
    return shuffle(symbols);
}

function shuffle(array) {
    for (let i = array.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [array[i], array[j]] = [array[j], array[i]];
    }
    return array;
}

function renderBoard(rows, cols) {
    const boardEl = document.getElementById('board');
    boardEl.style.gridTemplateColumns = `repeat(${cols}, 1fr)`;
    boardEl.innerHTML = '';
    board.forEach((symbol, index) => {
        const card = document.createElement('div');
        card.className = 'card';
        card.dataset.index = index;
        card.innerHTML = `
            <div class="card-back">?</div>
            <div class="card-front">${symbol}</div>
        `;
        card.addEventListener('click', () => flipCard(card, index));
        boardEl.appendChild(card);
    });
}

function flipCard(card, index) {
    if (card.classList.contains('flipped') || card.classList.contains('matched') || flippedCards.length >= 2) return;
    if (!startTime) startTime = Date.now();
    card.classList.add('flipped');
    flippedCards.push({ card, index });
    if (flippedCards.length === 2) {
        moves++;
        setTimeout(checkMatch, 1000);
    }
}

function checkMatch() {
    const [card1, card2] = flippedCards;
    if (board[card1.index] === board[card2.index]) {
        card1.card.classList.add('matched');
        card2.card.classList.add('matched');
        matchedPairs++;
        score += 10; // base score
        if (matchedPairs === difficulties[currentDifficulty].pairs) {
            score += Math.max(0, 100 - timer); // bonus for speed
            win();
        }
    } else {
        card1.card.classList.remove('flipped');
        card2.card.classList.remove('flipped');
        score = Math.max(0, score - 2); // penalty
    }
    flippedCards = [];
    updateStats();
}

function win() {
    clearInterval(interval);
    document.getElementById('win-time').textContent = formatTime(timer);
    document.getElementById('win-moves').textContent = moves;
    document.getElementById('win-score').textContent = score;
    document.getElementById('win-modal').style.display = 'flex';
}

function submitScore() {
    const name = document.getElementById('player-name').value;
    if (!name) return;
    fetch('/api/game/result', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            name,
            difficulty: currentDifficulty,
            time: timer,
            moves,
            score
        })
    }).then(() => {
        loadLeaderboard(document.querySelector('.filter.active').dataset.difficulty);
        document.getElementById('win-modal').style.display = 'none';
    });
}

function loadLeaderboard(difficulty) {
    fetch(`/api/leaderboard?difficulty=${difficulty}`)
        .then(res => res.json())
        .then(data => {
            const content = document.getElementById('leaderboard-content');
            content.innerHTML = data.map((s, i) =>
                `<p>${i+1}. ${s.name}: ${s.score} pts (${formatTime(s.time)}, ${s.moves} moves) - ${s.difficulty}</p>`
            ).join('');
        });
}

function updateStats() {
    document.getElementById('timer').textContent = formatTime(timer);
    document.getElementById('moves').textContent = moves;
    document.getElementById('score').textContent = score;
}

function formatTime(seconds) {
    const min = Math.floor(seconds / 60);
    const sec = seconds % 60;
    return `${min.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`;
}