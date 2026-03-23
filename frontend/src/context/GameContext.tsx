import { createContext, useContext, useReducer, type Dispatch, type ReactNode } from 'react'
import type { GameStatus, Guess, LetterResult, Player } from '../types/game'
import { WORD_LENGTH } from '../utils/constants'

export interface GameState {
  roomCode: string | null
  playerName: string | null
  playerId: string | null
  gameStatus: GameStatus
  wordLength: number
  maxGuesses: number
  players: Record<string, Player>
  winner: string | null
  targetWord: string | null
  currentGuess: string
  guesses: Guess[]
  keyboardState: Record<string, LetterResult>
  error: string | null
  shakeRow: boolean
}

const initialState: GameState = {
  roomCode: null,
  playerName: null,
  playerId: null,
  gameStatus: 'waiting',
  wordLength: WORD_LENGTH,
  maxGuesses: 6,
  players: {},
  winner: null,
  targetWord: null,
  currentGuess: '',
  guesses: [],
  keyboardState: {},
  error: null,
  shakeRow: false,
}

export type GameAction =
  | { type: 'SET_ROOM'; roomCode: string }
  | { type: 'SET_PLAYER'; playerId: string; playerName: string }
  | { type: 'JOIN_SUCCESS'; players: Record<string, Player>; gameStatus: GameStatus }
  | { type: 'GAME_STARTED'; maxGuesses: number }
  | { type: 'ADD_LETTER'; letter: string }
  | { type: 'REMOVE_LETTER' }
  | { type: 'GUESS_RESULT'; guess: Guess }
  | { type: 'GAME_UPDATE'; players: Record<string, Player>; gameStatus: GameStatus; winner: string | null }
  | { type: 'GAME_COMPLETED'; winner: string; targetWord: string }
  | { type: 'PLAYER_UPDATE'; players: Record<string, Player> }
  | { type: 'SET_ERROR'; error: string }
  | { type: 'CLEAR_ERROR' }
  | { type: 'SHAKE_ROW' }
  | { type: 'RESET' }

function updateKeyboardState(
  current: Record<string, LetterResult>,
  word: string,
  results: LetterResult[],
): Record<string, LetterResult> {
  const next = { ...current }
  const priority: Record<string, number> = { correct: 3, present: 2, absent: 1 }
  for (let i = 0; i < word.length; i++) {
    const letter = word[i].toUpperCase()
    const result = results[i]
    if (!next[letter] || priority[result] > priority[next[letter]]) {
      next[letter] = result
    }
  }
  return next
}

function gameReducer(state: GameState, action: GameAction): GameState {
  switch (action.type) {
    case 'SET_ROOM':
      return { ...state, roomCode: action.roomCode }
    case 'SET_PLAYER':
      return { ...state, playerId: action.playerId, playerName: action.playerName }
    case 'JOIN_SUCCESS':
      return {
        ...state,
        players: action.players,
        gameStatus: action.gameStatus,
      }
    case 'GAME_STARTED':
      return {
        ...state,
        gameStatus: 'active',
        maxGuesses: action.maxGuesses,
        guesses: [],
        currentGuess: '',
        keyboardState: {},
        winner: null,
        targetWord: null,
      }
    case 'ADD_LETTER':
      if (state.currentGuess.length >= state.wordLength) return state
      return { ...state, currentGuess: state.currentGuess + action.letter.toUpperCase() }
    case 'REMOVE_LETTER':
      return { ...state, currentGuess: state.currentGuess.slice(0, -1) }
    case 'GUESS_RESULT':
      return {
        ...state,
        guesses: [...state.guesses, action.guess],
        currentGuess: '',
        keyboardState: updateKeyboardState(state.keyboardState, action.guess.word, action.guess.results),
      }
    case 'GAME_UPDATE':
      return {
        ...state,
        players: action.players,
        gameStatus: action.gameStatus,
        winner: action.winner,
      }
    case 'GAME_COMPLETED':
      return {
        ...state,
        gameStatus: 'finished',
        winner: action.winner,
        targetWord: action.targetWord,
      }
    case 'PLAYER_UPDATE':
      return { ...state, players: action.players }
    case 'SET_ERROR':
      return { ...state, error: action.error, shakeRow: false }
    case 'CLEAR_ERROR':
      return { ...state, error: null }
    case 'SHAKE_ROW':
      return { ...state, shakeRow: true }
    case 'RESET':
      return { ...initialState, playerName: state.playerName }
    default:
      return state
  }
}

const GameStateContext = createContext<GameState>(initialState)
const GameDispatchContext = createContext<Dispatch<GameAction>>(() => {})

export function GameProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(gameReducer, initialState)
  return (
    <GameStateContext.Provider value={state}>
      <GameDispatchContext.Provider value={dispatch}>{children}</GameDispatchContext.Provider>
    </GameStateContext.Provider>
  )
}

export function useGameState() {
  return useContext(GameStateContext)
}

export function useGameDispatch() {
  return useContext(GameDispatchContext)
}
