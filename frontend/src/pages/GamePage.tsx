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

  const playerName = localStorage.getItem('worduel_name')
  const game = useGame({ roomCode })

  useKeyboard({
    onLetter: game.addLetter,
    onEnter: game.submitGuess,
    onBackspace: game.removeLetter,
    enabled: game.gameStatus === 'active',
  })

  if (!playerName) {
    return <Navigate to="/" replace />
  }

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

  const opponents = Object.values(game.players).filter((p) => p.id !== game.clientId)
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
            maxPlayers={game.maxGuesses === 6 ? Math.max(playerCount, 2) : 2}
            players={game.players}
          />
        ) : (
          <div className="flex items-start gap-6 py-4">
            {/* Opponents sidebar (left) */}
            {opponents.length > 0 && (
              <div className="hidden sm:flex flex-col gap-4 w-28">
                {opponents.map((opp) => (
                  <OpponentProgress key={opp.id} opponent={opp} />
                ))}
              </div>
            )}

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

            {/* Spacer for symmetry when opponents are shown */}
            {opponents.length > 0 && <div className="hidden sm:block w-28" />}
          </div>
        )}

        {/* Mobile opponents */}
        {game.gameStatus === 'active' && opponents.length > 0 && (
          <div className="sm:hidden fixed bottom-2 right-2 flex flex-col gap-2">
            {opponents.slice(0, 3).map((opp) => (
              <OpponentProgress key={opp.id} opponent={opp} />
            ))}
            {opponents.length > 3 && (
              <p className="text-gray-500 text-xs text-center">+{opponents.length - 3} more</p>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
