import { motion } from 'framer-motion'
import { GameTile } from './GameTile'
import type { LetterResult } from '../../types/game'
import { WORD_LENGTH } from '../../utils/constants'

interface GameRowProps {
  word: string
  results?: LetterResult[]
  isRevealing?: boolean
  shake?: boolean
}

export function GameRow({ word, results, isRevealing, shake }: GameRowProps) {
  const tiles = Array.from({ length: WORD_LENGTH }, (_, i) => ({
    letter: word[i] || '',
    result: results?.[i],
  }))

  return (
    <motion.div
      className="flex gap-[5px]"
      animate={shake ? { x: [0, -10, 10, -10, 10, 0] } : {}}
      transition={{ duration: 0.3 }}
    >
      {tiles.map((tile, i) => (
        <GameTile
          key={i}
          letter={tile.letter}
          result={tile.result}
          isRevealing={isRevealing}
          revealDelay={i * 0.15}
          isCurrent={!results && !!tile.letter}
        />
      ))}
    </motion.div>
  )
}
