import type { GameSummary, LetterResult } from './game'

export interface ServerMessage {
  type: string
  player_id?: string
  room_id?: string
  data: unknown
  timestamp: string
}

export interface ConnectionAckData {
  client_id: string
  connected_at: string
}

export interface JoinSuccessData {
  game_state: {
    status: string
    word_length: number
    max_guesses: number
    current_round: number
    started_at?: string
    finished_at?: string
    winner?: string
  }
  players: Record<string, { id: string; name: string; status: string }>
}

export interface GameStartedData {
  target_word_length: number
  max_guesses: number
  game_status: string
}

export interface GuessResultData {
  word: string
  results: LetterResult[]
  is_correct: boolean
  timestamp: string
}

export interface GameUpdateData {
  game_summary: GameSummary
  triggering_player: string
  update_reason: string
}

export interface GameCompletedData {
  winner: string
  game_status: string
  completed_at: string
  target_word: string
}

export interface PlayerUpdateData {
  event: 'player_joined' | 'player_left' | 'player_disconnected'
  player_id: string
  player_name?: string
  player_count?: number
}

export interface ErrorData {
  code: string
  message: string
}
