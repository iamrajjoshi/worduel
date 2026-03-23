import { motion } from 'framer-motion'
import type { LetterResult } from '../../types/game'
import { COLORS } from '../../utils/constants'

interface GameTileProps {
  letter: string
  result?: LetterResult
  isRevealing?: boolean
  revealDelay?: number
  isCurrent?: boolean
}

const resultBg: Record<LetterResult, string> = {
  correct: COLORS.correct,
  present: COLORS.present,
  absent: COLORS.absent,
}

export function GameTile({ letter, result, isRevealing, revealDelay = 0, isCurrent }: GameTileProps) {
  const hasResult = result !== undefined
  const bg = hasResult ? resultBg[result] : 'transparent'
  const border = hasResult ? bg : letter ? COLORS.borderFilled : COLORS.border

  return (
    <motion.div
      className="w-[62px] h-[62px] flex items-center justify-center text-3xl font-bold uppercase select-none"
      style={{
        backgroundColor: bg,
        border: `2px solid ${border}`,
        color: COLORS.text,
      }}
      initial={false}
      animate={
        isRevealing
          ? {
              rotateX: [0, 90, 0],
              transition: { duration: 0.5, delay: revealDelay, times: [0, 0.5, 1] },
            }
          : isCurrent && letter
            ? { scale: [1, 1.1, 1], transition: { duration: 0.1 } }
            : {}
      }
    >
      {letter}
    </motion.div>
  )
}
