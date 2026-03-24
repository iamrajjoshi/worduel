import { useLocation, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import type { Guess, LetterResult, Player } from '../types/game'
import { COLORS, WORD_LENGTH } from '../utils/constants'

interface ResultsState {
  winner: string | null
  targetWord: string
  guesses: Guess[]
  players: Record<string, Player>
  playerId: string | null
  playerName: string | null
}

const resultBg: Record<string, string> = {
  correct: COLORS.correct,
  present: COLORS.present,
  absent: COLORS.absent,
}

export function ResultsPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const state = location.state as ResultsState | null

  if (!state) {
    return (
      <div className="min-h-screen flex items-center justify-center" style={{ backgroundColor: '#121213' }}>
        <div className="text-center">
          <p className="text-gray-400 mb-4">No game data found</p>
          <button
            onClick={() => navigate('/')}
            className="bg-[#538d4e] text-white font-bold py-2 px-6 rounded-lg cursor-pointer"
          >
            Go Home
          </button>
        </div>
      </div>
    )
  }

  const { winner, targetWord, guesses, players, playerId } = state
  const isWinner = winner === playerId
  const winnerPlayer = winner ? Object.values(players).find((p) => p.id === winner) : null
  const allPlayers = Object.values(players).sort((a, b) => b.score - a.score)

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4 py-8" style={{ backgroundColor: '#121213' }}>
      {/* Result announcement */}
      <motion.div
        initial={{ scale: 0.8, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ duration: 0.5, type: 'spring' }}
        className="text-center mb-6"
      >
        <h1 className="text-4xl font-bold mb-2" style={{ color: isWinner ? COLORS.correct : '#e5484d' }}>
          {winner ? (isWinner ? 'You Won!' : 'You Lost') : "Time's Up!"}
        </h1>
        {winnerPlayer && !isWinner && (
          <p className="text-gray-400">
            <span className="text-white font-semibold">{winnerPlayer.name}</span> won
          </p>
        )}
      </motion.div>

      {/* Word reveal */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3 }}
        className="mb-6"
      >
        <p className="text-gray-400 text-sm text-center mb-2">The word was</p>
        <div className="flex gap-[5px] justify-center">
          {targetWord.split('').map((letter, i) => (
            <div
              key={i}
              className="w-[48px] h-[48px] flex items-center justify-center text-xl font-bold text-white rounded"
              style={{ backgroundColor: COLORS.correct }}
            >
              {letter.toUpperCase()}
            </div>
          ))}
        </div>
      </motion.div>

      {/* Scoreboard */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5 }}
        className="mb-6 w-full max-w-sm"
      >
        <p className="text-gray-400 text-sm text-center mb-3">Scoreboard</p>
        <div className="flex flex-col gap-2">
          {allPlayers.map((player, idx) => (
            <div
              key={player.id}
              className="flex items-center justify-between px-4 py-2 rounded-lg"
              style={{
                backgroundColor: player.id === winner ? '#538d4e22' : '#1a1a1b',
                border: player.id === winner ? '1px solid #538d4e' : '1px solid #3a3a3c',
              }}
            >
              <div className="flex items-center gap-3">
                <span className="text-gray-500 text-sm w-5">{idx + 1}.</span>
                <span className="text-white font-medium text-sm">
                  {player.name}
                  {player.id === playerId && <span className="text-gray-500 ml-1">(you)</span>}
                </span>
              </div>
              <div className="flex items-center gap-3">
                <span className="text-gray-400 text-xs">{player.guess_count} guess{player.guess_count !== 1 ? 'es' : ''}</span>
                {player.score > 0 && <span className="text-white font-bold text-sm">{player.score}pts</span>}
              </div>
            </div>
          ))}
        </div>
      </motion.div>

      {/* Your board summary */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.6 }}
        className="mb-6"
      >
        <p className="text-gray-400 text-sm text-center mb-3">Your guesses</p>
        <div className="flex flex-col gap-1 items-center">
          {guesses.map((guess, rowIdx) => (
            <div key={rowIdx} className="flex gap-1">
              {Array.from({ length: WORD_LENGTH }, (_, colIdx) => (
                <div
                  key={colIdx}
                  className="w-8 h-8 flex items-center justify-center text-xs font-bold text-white rounded-sm"
                  style={{ backgroundColor: resultBg[guess.results[colIdx]] }}
                >
                  {guess.word[colIdx]?.toUpperCase()}
                </div>
              ))}
            </div>
          ))}
        </div>
      </motion.div>

      {/* Opponent patterns */}
      {allPlayers.filter((p) => p.id !== playerId).length > 0 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.7 }}
          className="mb-6"
        >
          <p className="text-gray-400 text-sm text-center mb-3">Opponents</p>
          <div className="flex gap-6 justify-center flex-wrap">
            {allPlayers
              .filter((p) => p.id !== playerId)
              .map((opponent) => {
                const patterns = opponent.guess_patterns || []
                return (
                  <div key={opponent.id} className="flex flex-col items-center gap-1">
                    <p className="text-gray-400 text-xs">{opponent.name}</p>
                    <div className="flex flex-col gap-[2px]">
                      {patterns.map((row: LetterResult[], rowIdx: number) => (
                        <div key={rowIdx} className="flex gap-[2px]">
                          {row.map((result: LetterResult, colIdx: number) => (
                            <div
                              key={colIdx}
                              className="w-3 h-3 rounded-sm"
                              style={{ backgroundColor: resultBg[result] }}
                            />
                          ))}
                        </div>
                      ))}
                    </div>
                  </div>
                )
              })}
          </div>
        </motion.div>
      )}

      {/* Actions */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.8 }}
        className="flex gap-4"
      >
        <button
          onClick={() => navigate('/')}
          className="bg-[#538d4e] hover:bg-[#6aaf5e] text-white font-bold py-3 px-8 rounded-lg text-lg transition-colors cursor-pointer"
        >
          Play Again
        </button>
        <button
          onClick={() => navigate('/')}
          className="bg-[#3a3a3c] hover:bg-[#4a4a4c] text-white font-bold py-3 px-8 rounded-lg text-lg transition-colors cursor-pointer"
        >
          Home
        </button>
      </motion.div>
    </div>
  )
}
