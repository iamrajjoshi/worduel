import type { LetterResult, Player } from '../../types/game'
import { COLORS, MAX_GUESSES, WORD_LENGTH } from '../../utils/constants'

interface OpponentProgressProps {
  opponent: Player | null
}

const resultBg: Record<LetterResult, string> = {
  correct: COLORS.correct,
  present: COLORS.present,
  absent: COLORS.absent,
}

export function OpponentProgress({ opponent }: OpponentProgressProps) {
  if (!opponent) {
    return (
      <div className="text-center text-gray-500 text-sm p-4">
        Waiting for opponent...
      </div>
    )
  }

  const patterns = opponent.guess_patterns || []

  return (
    <div className="flex flex-col items-center gap-2">
      <p className="text-sm font-medium text-gray-300 truncate max-w-[120px]">{opponent.name}</p>
      <div className="flex flex-col gap-1">
        {Array.from({ length: MAX_GUESSES }, (_, rowIdx) => (
          <div key={rowIdx} className="flex gap-1">
            {Array.from({ length: WORD_LENGTH }, (_, colIdx) => {
              const result = patterns[rowIdx]?.[colIdx]
              return (
                <div
                  key={colIdx}
                  className="w-3 h-3 rounded-sm"
                  style={{
                    backgroundColor: result ? resultBg[result] : rowIdx < patterns.length ? COLORS.absent : '#2a2a2a',
                  }}
                />
              )
            })}
          </div>
        ))}
      </div>
      <p className="text-xs text-gray-500">
        {patterns.length}/{MAX_GUESSES} guesses
      </p>
    </div>
  )
}
