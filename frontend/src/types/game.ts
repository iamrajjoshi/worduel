export type GameStatus = 'waiting' | 'active' | 'finished'
export type PlayerStatus = 'active' | 'disconnected' | 'finished'
export type LetterResult = 'correct' | 'present' | 'absent'

export interface Guess {
  word: string
  results: LetterResult[]
  timestamp: string
  is_correct: boolean
}

export interface Player {
  id: string
  name: string
  status: PlayerStatus
  score: number
  guess_count: number
  guesses?: Guess[]
  guess_patterns?: LetterResult[][]
  last_activity: string
}

export interface GameSummary {
  status: GameStatus
  word_length: number
  max_guesses: number
  current_round: number
  started_at?: string
  finished_at?: string
  winner?: string
  players: Record<string, Player>
}
