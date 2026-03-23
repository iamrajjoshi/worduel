import { useParams, useNavigate, Navigate } from 'react-router-dom'
import { useGame } from '../hooks/useGame'
import { useKeyboard } from '../hooks/useKeyboard'
import { GameBoard } from '../components/game/GameBoard'
import { Keyboard } from '../components/game/Keyboard'
import { OpponentProgress } from '../components/game/OpponentProgress'
import { LobbyView } from '../components/lobby/LobbyView'
import { ConnectionStatus } from '../components/common/ConnectionStatus'
import { ErrorBanner } from '../components/common/ErrorBanner'

export function GamePage() {
  const { roomCode } = useParams<{ roomCode: string }>()
  const navigate = useNavigate()

  // If no name in localStorage, redirect to landing
  const playerName = localStorage.getItem('worduel_name')
  const game = useGame({ roomCode })

  // Physical keyboard
  useKeyboard({
    onLetter: game.addLetter,
    onEnter: game.submitGuess,
    onBackspace: game.removeLetter,
    enabled: game.gameStatus === 'active',
  })

  // Redirect to landing if no name
  if (!playerName) {
    return <Navigate to="/" replace />
  }

  // Navigate to results when game finishes
  if (game.gameStatus === 'finished' && game.targetWord) {
    return (
      <Navigate
        to={`/results/${roomCode}`}
        replace
        state={{
          winner: game.winner,
          targetWord: game.targetWord,
          guesses: game.guesses,
          players: game.players,
          playerId: game.clientId,
          playerName: game.playerName,
        }}
      />
    )
  }

  // Find opponent
  const opponent = Object.values(game.players).find((p) => p.id !== game.clientId) ?? null
  const playerCount = Object.keys(game.players).length

  return (
    <div className="min-h-screen flex flex-col" style={{ backgroundColor: '#121213' }}>
      {/* Header */}
      <header className="flex items-center justify-between px-4 py-3 border-b border-[#3a3a3c]">
        <button
          onClick={() => {
            game.leaveRoom()
            navigate('/')
          }}
          className="text-gray-400 hover:text-white text-sm cursor-pointer"
        >
          Leave
        </button>
        <h1 className="text-xl font-bold text-white tracking-wider">WORDUEL</h1>
        <ConnectionStatus status={game.connectionStatus} />
      </header>

      <ErrorBanner message={game.error} />

      {/* Main content */}
      <div className="flex-1 flex flex-col items-center justify-center relative">
        {game.gameStatus === 'waiting' ? (
          <LobbyView
            roomCode={roomCode || ''}
            playerName={playerName}
            playerCount={playerCount}
          />
        ) : (
          <div className="flex items-start gap-8 py-4">
            {/* Opponent sidebar (left) */}
            <div className="hidden sm:block w-24">
              <OpponentProgress opponent={opponent} />
            </div>

            {/* Game board (center) */}
            <div className="flex flex-col items-center gap-6">
              <GameBoard
                guesses={game.guesses}
                currentGuess={game.currentGuess}
                shakeRow={game.shakeRow}
              />
              <Keyboard
                keyboardState={game.keyboardState}
                onLetter={game.addLetter}
                onEnter={game.submitGuess}
                onBackspace={game.removeLetter}
              />
            </div>

            {/* Spacer for symmetry on desktop */}
            <div className="hidden sm:block w-24" />
          </div>
        )}

        {/* Mobile opponent indicator */}
        {game.gameStatus === 'active' && opponent && (
          <div className="sm:hidden fixed bottom-2 right-2">
            <OpponentProgress opponent={opponent} />
          </div>
        )}
      </div>
    </div>
  )
}
