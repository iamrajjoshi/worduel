import { GameRow } from './GameRow'
import type { Guess } from '../../types/game'
import { MAX_GUESSES } from '../../utils/constants'

interface GameBoardProps {
  guesses: Guess[]
  currentGuess: string
  shakeRow: boolean
}

export function GameBoard({ guesses, currentGuess, shakeRow }: GameBoardProps) {
  const rows = Array.from({ length: MAX_GUESSES }, (_, i) => {
    if (i < guesses.length) {
      return {
        word: guesses[i].word.toUpperCase(),
        results: guesses[i].results,
        isRevealing: false,
      }
    }
    if (i === guesses.length) {
      return { word: currentGuess, results: undefined, isRevealing: false }
    }
    return { word: '', results: undefined, isRevealing: false }
  })

  return (
    <div className="flex flex-col gap-[5px] items-center">
      {rows.map((row, i) => (
        <GameRow
          key={i}
          word={row.word}
          results={row.results}
          isRevealing={row.isRevealing}
          shake={shakeRow && i === guesses.length}
        />
      ))}
    </div>
  )
}
