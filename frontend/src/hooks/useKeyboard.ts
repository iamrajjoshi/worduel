import { useEffect } from 'react'

interface UseKeyboardOptions {
  onLetter: (letter: string) => void
  onEnter: () => void
  onBackspace: () => void
  enabled: boolean
}

export function useKeyboard({ onLetter, onEnter, onBackspace, enabled }: UseKeyboardOptions) {
  useEffect(() => {
    if (!enabled) return

    const handler = (e: KeyboardEvent) => {
      if (e.ctrlKey || e.metaKey || e.altKey) return
      if (e.key === 'Enter') {
        e.preventDefault()
        onEnter()
      } else if (e.key === 'Backspace') {
        e.preventDefault()
        onBackspace()
      } else if (/^[a-zA-Z]$/.test(e.key)) {
        onLetter(e.key)
      }
    }

    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [enabled, onLetter, onEnter, onBackspace])
}
