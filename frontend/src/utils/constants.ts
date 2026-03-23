export const MAX_GUESSES = 6
export const WORD_LENGTH = 5

export const COLORS = {
  correct: '#538d4e',
  present: '#b59f3b',
  absent: '#3a3a3c',
  border: '#3a3a3c',
  borderFilled: '#565758',
  bg: '#121213',
  keyDefault: '#818384',
  text: '#ffffff',
} as const

export const KEYBOARD_ROWS = [
  ['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
  ['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'],
  ['ENTER', 'Z', 'X', 'C', 'V', 'B', 'N', 'M', 'BACK'],
] as const

export const ERROR_MESSAGES: Record<string, string> = {
  INVALID_WORD: 'Not a valid word',
  ROOM_NOT_FOUND: 'Room not found',
  ROOM_FULL: 'Room is full',
  PLAYER_EXISTS: 'Already in this room',
  GAME_NOT_ACTIVE: 'Game is not active',
  TOO_MANY_GUESSES: 'No more guesses',
  GAME_ALREADY_WON: 'Game already ended',
  INVALID_WORD_LENGTH: 'Word must be 5 letters',
  NOT_IN_ROOM: 'Not in a room',
  RATE_LIMIT_EXCEEDED: 'Slow down!',
}
