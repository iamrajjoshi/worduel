import { useLocation, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import type { Guess, Player } from '../types/game'
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

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4" style={{ backgroundColor: '#121213' }}>
      {/* Result announcement */}
      <motion.div
        initial={{ scale: 0.8, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ duration: 0.5, type: 'spring' }}
        className="text-center mb-8"
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
        className="mb-8"
      >
        <p className="text-gray-400 text-sm text-center mb-2">The word was</p>
        <div className="flex gap-[5px]">
          {targetWord.split('').map((letter, i) => (
            <div
              key={i}
              className="w-[52px] h-[52px] flex items-center justify-center text-2xl font-bold text-white rounded"
              style={{ backgroundColor: COLORS.correct }}
            >
              {letter.toUpperCase()}
            </div>
          ))}
        </div>
      </motion.div>

      {/* Your board summary */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5 }}
        className="mb-8"
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
        <p className="text-gray-500 text-xs text-center mt-2">
          {guesses.length} guess{guesses.length !== 1 ? 'es' : ''}
        </p>
      </motion.div>

      {/* Actions */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.7 }}
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
