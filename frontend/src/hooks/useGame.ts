import { useCallback, useRef } from 'react'
import { useGameDispatch, useGameState } from '../context/GameContext'
import { useWebSocket, type SendFn } from './useWebSocket'
import { createRoom } from '../utils/api'
import { ERROR_MESSAGES, WORD_LENGTH } from '../utils/constants'
import type {
  ServerMessage,
  GameStartedData,
  GuessResultData,
  GameUpdateData,
  GameCompletedData,
  ErrorData,
  JoinSuccessData,
} from '../types/messages'
import type { GameStatus, Player } from '../types/game'

interface UseGameOptions {
  roomCode?: string
}

export function useGame({ roomCode }: UseGameOptions = {}) {
  const state = useGameState()
  const dispatch = useGameDispatch()
  const hasJoined = useRef(false)
  const roomCodeRef = useRef(roomCode)
  roomCodeRef.current = roomCode

  const handleMessage = useCallback(
    (message: ServerMessage) => {
      switch (message.type) {
        case 'connection_ack': {
          const data = message.data as { client_id: string }
          const name = localStorage.getItem('worduel_name') || 'Player'
          dispatch({ type: 'SET_PLAYER', playerId: data.client_id, playerName: name })
          break
        }
        case 'join_success': {
          const data = message.data as JoinSuccessData
          dispatch({
            type: 'JOIN_SUCCESS',
            players: data.players as unknown as Record<string, Player>,
            gameStatus: (data.game_state?.status ?? 'waiting') as GameStatus,
          })
          break
        }
        case 'game_started': {
          const data = message.data as GameStartedData
          dispatch({ type: 'GAME_STARTED', maxGuesses: data.max_guesses })
          break
        }
        case 'guess_result': {
          const data = message.data as GuessResultData
          dispatch({
            type: 'GUESS_RESULT',
            guess: {
              word: data.word,
              results: data.results,
              is_correct: data.is_correct,
              timestamp: data.timestamp,
            },
          })
          break
        }
        case 'game_update': {
          const data = message.data as GameUpdateData
          if (data.game_summary) {
            dispatch({
              type: 'GAME_UPDATE',
              players: data.game_summary.players,
              gameStatus: data.game_summary.status,
              winner: data.game_summary.winner ?? null,
            })
          }
          break
        }
        case 'game_completed': {
          const data = message.data as GameCompletedData
          dispatch({
            type: 'GAME_COMPLETED',
            winner: data.winner,
            targetWord: data.target_word,
          })
          break
        }
        case 'player_update': {
          // Server broadcasts these but we get full state via game_update
          break
        }
        case 'error': {
          const data = message.data as ErrorData
          const friendlyMessage = ERROR_MESSAGES[data.code] || data.message
          dispatch({ type: 'SET_ERROR', error: friendlyMessage })
          setTimeout(() => dispatch({ type: 'CLEAR_ERROR' }), 3000)
          break
        }
      }
    },
    [dispatch],
  )

  const { status: connectionStatus, clientId, send, disconnect } = useWebSocket({
    onMessage: handleMessage,
    onConnected: useCallback(
      (wsSend: SendFn) => {
        if (roomCodeRef.current && !hasJoined.current) {
          const name = localStorage.getItem('worduel_name') || 'Player'
          wsSend('join', { room_id: roomCodeRef.current, player_name: name })
          dispatch({ type: 'SET_ROOM', roomCode: roomCodeRef.current })
          hasJoined.current = true
        }
      },
      [dispatch],
    ),
  })

  const joinRoom = useCallback(
    (code: string, playerName: string) => {
      if (hasJoined.current) return
      localStorage.setItem('worduel_name', playerName)
      dispatch({ type: 'SET_PLAYER', playerId: clientId ?? '', playerName })
      dispatch({ type: 'SET_ROOM', roomCode: code })
      send('join', { room_id: code, player_name: playerName })
      hasJoined.current = true
    },
    [send, clientId, dispatch],
  )

  const handleCreateRoom = useCallback(async () => {
    const room = await createRoom()
    return room.roomCode
  }, [])

  const submitGuess = useCallback(() => {
    if (state.currentGuess.length !== WORD_LENGTH) {
      dispatch({ type: 'SHAKE_ROW' })
      dispatch({ type: 'SET_ERROR', error: 'Not enough letters' })
      setTimeout(() => dispatch({ type: 'CLEAR_ERROR' }), 2000)
      return
    }
    send('guess', { word: state.currentGuess.toLowerCase() })
  }, [state.currentGuess, send, dispatch])

  const addLetter = useCallback(
    (letter: string) => {
      if (state.gameStatus !== 'active') return
      if (state.guesses.length >= state.maxGuesses) return
      dispatch({ type: 'ADD_LETTER', letter })
    },
    [state.gameStatus, state.guesses.length, state.maxGuesses, dispatch],
  )

  const removeLetter = useCallback(() => {
    dispatch({ type: 'REMOVE_LETTER' })
  }, [dispatch])

  const leaveRoom = useCallback(() => {
    send('leave', {})
    hasJoined.current = false
    dispatch({ type: 'RESET' })
    disconnect()
  }, [send, dispatch, disconnect])

  const resetForNewGame = useCallback(() => {
    hasJoined.current = false
    dispatch({ type: 'RESET' })
  }, [dispatch])

  return {
    ...state,
    connectionStatus,
    clientId,
    joinRoom,
    createRoom: handleCreateRoom,
    submitGuess,
    addLetter,
    removeLetter,
    leaveRoom,
    resetForNewGame,
  }
}
