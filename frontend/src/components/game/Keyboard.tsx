import type { LetterResult } from '../../types/game'
import { COLORS, KEYBOARD_ROWS } from '../../utils/constants'

interface KeyboardProps {
  keyboardState: Record<string, LetterResult>
  onLetter: (letter: string) => void
  onEnter: () => void
  onBackspace: () => void
}

const resultBg: Record<string, string> = {
  correct: COLORS.correct,
  present: COLORS.present,
  absent: COLORS.absent,
}

export function Keyboard({ keyboardState, onLetter, onEnter, onBackspace }: KeyboardProps) {
  const handleClick = (key: string) => {
    if (key === 'ENTER') onEnter()
    else if (key === 'BACK') onBackspace()
    else onLetter(key)
  }

  return (
    <div className="flex flex-col gap-2 items-center w-full max-w-[500px] mx-auto">
      {KEYBOARD_ROWS.map((row, rowIdx) => (
        <div key={rowIdx} className="flex gap-[6px] justify-center">
          {row.map((key) => {
            const isWide = key === 'ENTER' || key === 'BACK'
            const bg = keyboardState[key] ? resultBg[keyboardState[key]] : COLORS.keyDefault
            return (
              <button
                key={key}
                onClick={() => handleClick(key)}
                className="flex items-center justify-center rounded font-bold text-sm uppercase select-none cursor-pointer active:opacity-70"
                style={{
                  backgroundColor: bg,
                  color: COLORS.text,
                  width: isWide ? 65 : 43,
                  height: 58,
                  fontSize: isWide ? 12 : 14,
                }}
              >
                {key === 'BACK' ? (
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                    <path d="M22 3H7c-.69 0-1.23.35-1.59.88L0 12l5.41 8.11c.36.53.9.89 1.59.89h15c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H7.07L2.4 12l4.66-7H22v14zm-11.59-2L14 13.41 17.59 17 19 15.59 15.41 12 19 8.41 17.59 7 14 10.59 10.41 7 9 8.41 12.59 12 9 15.59z" fill="currentColor"/>
                  </svg>
                ) : (
                  key
                )}
              </button>
            )
          })}
        </div>
      ))}
    </div>
  )
}
