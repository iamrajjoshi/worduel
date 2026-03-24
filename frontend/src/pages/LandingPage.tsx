import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { createRoom } from '../utils/api'

const TITLE_LETTERS = ['W', 'O', 'R', 'D', 'U', 'E', 'L']
const TITLE_COLORS = ['#538d4e', '#b59f3b', '#538d4e', '#3a3a3c', '#b59f3b', '#538d4e', '#3a3a3c']
const PLAYER_COUNTS = [2, 3, 4, 5, 6, 7, 8] as const

export function LandingPage() {
  const navigate = useNavigate()
  const [name, setName] = useState(() => localStorage.getItem('worduel_name') || '')
  const [joinCode, setJoinCode] = useState('')
  const [maxPlayers, setMaxPlayers] = useState(2)
  const [error, setError] = useState('')
  const [creating, setCreating] = useState(false)

  const handleCreate = async () => {
    if (!name.trim()) {
      setError('Enter your name first')
      return
    }
    setCreating(true)
    setError('')
    try {
      localStorage.setItem('worduel_name', name.trim())
      const room = await createRoom(maxPlayers)
      navigate(`/room/${room.roomCode}`)
    } catch {
      setError('Failed to create room. Is the server running?')
    } finally {
      setCreating(false)
    }
  }

  const handleJoin = () => {
    if (!name.trim()) {
      setError('Enter your name first')
      return
    }
    const code = joinCode.trim().toUpperCase()
    if (code.length !== 6) {
      setError('Room code must be 6 characters')
      return
    }
    localStorage.setItem('worduel_name', name.trim())
    navigate(`/room/${code}`)
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4" style={{ backgroundColor: '#121213' }}>
      {/* Title */}
      <div className="flex gap-[6px] mb-2">
        {TITLE_LETTERS.map((letter, i) => (
          <div
            key={i}
            className="w-[52px] h-[52px] flex items-center justify-center text-2xl font-bold text-white rounded"
            style={{ backgroundColor: TITLE_COLORS[i] }}
          >
            {letter}
          </div>
        ))}
      </div>
      <p className="text-gray-500 text-sm mb-10">Competitive Wordle with friends</p>

      {/* Name input */}
      <div className="w-full max-w-xs mb-6">
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Your name"
          maxLength={20}
          className="w-full bg-[#1a1a1b] border-2 border-[#3a3a3c] rounded-lg px-4 py-3 text-white text-center text-lg placeholder-gray-600 focus:border-[#538d4e] focus:outline-none transition-colors"
        />
      </div>

      {/* Player count selector */}
      <div className="w-full max-w-xs mb-4">
        <p className="text-gray-400 text-xs text-center mb-2">Players</p>
        <div className="flex gap-1 justify-center">
          {PLAYER_COUNTS.map((count) => (
            <button
              key={count}
              onClick={() => setMaxPlayers(count)}
              className="w-9 h-9 rounded font-bold text-sm cursor-pointer transition-colors"
              style={{
                backgroundColor: maxPlayers === count ? '#538d4e' : '#1a1a1b',
                color: maxPlayers === count ? '#fff' : '#818384',
                border: `2px solid ${maxPlayers === count ? '#538d4e' : '#3a3a3c'}`,
              }}
            >
              {count}
            </button>
          ))}
        </div>
      </div>

      {/* Actions */}
      <div className="flex flex-col gap-4 w-full max-w-xs">
        <button
          onClick={handleCreate}
          disabled={creating}
          className="w-full bg-[#538d4e] hover:bg-[#6aaf5e] disabled:opacity-50 text-white font-bold py-3 px-6 rounded-lg text-lg transition-colors cursor-pointer"
        >
          {creating ? 'Creating...' : `Create Game (${maxPlayers} players)`}
        </button>

        <div className="flex items-center gap-3">
          <div className="flex-1 h-px bg-[#3a3a3c]" />
          <span className="text-gray-600 text-sm">or</span>
          <div className="flex-1 h-px bg-[#3a3a3c]" />
        </div>

        <div className="flex gap-2">
          <input
            type="text"
            value={joinCode}
            onChange={(e) => setJoinCode(e.target.value.toUpperCase().slice(0, 6))}
            onKeyDown={(e) => e.key === 'Enter' && handleJoin()}
            placeholder="Room code"
            maxLength={6}
            className="flex-1 bg-[#1a1a1b] border-2 border-[#3a3a3c] rounded-lg px-4 py-3 text-white text-center font-mono text-lg tracking-widest placeholder-gray-600 uppercase focus:border-[#b59f3b] focus:outline-none transition-colors"
          />
          <button
            onClick={handleJoin}
            className="bg-[#b59f3b] hover:bg-[#c9b04a] text-white font-bold py-3 px-6 rounded-lg transition-colors cursor-pointer"
          >
            Join
          </button>
        </div>
      </div>

      {/* Error */}
      {error && (
        <p className="text-red-400 text-sm mt-4">{error}</p>
      )}
    </div>
  )
}
