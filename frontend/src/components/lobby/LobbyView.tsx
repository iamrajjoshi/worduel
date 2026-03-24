import { useState } from 'react'
import type { Player } from '../../types/game'

interface LobbyViewProps {
  roomCode: string
  playerName: string
  playerCount: number
  maxPlayers: number
  players: Record<string, Player>
}

export function LobbyView({ roomCode, playerName, playerCount, maxPlayers, players }: LobbyViewProps) {
  const [copied, setCopied] = useState(false)

  const copyCode = async () => {
    await navigator.clipboard.writeText(roomCode)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const playerList = Object.values(players)

  return (
    <div className="flex flex-col items-center gap-8 py-12">
      <div className="text-center">
        <p className="text-gray-400 text-sm mb-2">Share this code with your friends</p>
        <button
          onClick={copyCode}
          className="flex items-center gap-3 bg-[#1a1a1b] border-2 border-[#3a3a3c] rounded-lg px-6 py-4 hover:border-[#565758] transition-colors cursor-pointer"
        >
          <span className="text-4xl font-mono font-bold tracking-[0.3em] text-white">
            {roomCode}
          </span>
          <span className="text-gray-400 text-sm">
            {copied ? 'Copied!' : 'Copy'}
          </span>
        </button>
      </div>

      {/* Player list */}
      <div className="flex flex-col items-center gap-3">
        <p className="text-gray-400 text-sm">
          {playerCount}/{maxPlayers} players joined
        </p>
        <div className="flex flex-col gap-2 min-w-[160px]">
          {playerList.map((p) => (
            <div key={p.id} className="flex items-center gap-2 bg-[#1a1a1b] rounded-lg px-4 py-2">
              <span className="w-2 h-2 rounded-full bg-green-500" />
              <span className="text-white text-sm font-medium">
                {p.name}
                {p.name === playerName && <span className="text-gray-500 ml-1">(you)</span>}
              </span>
            </div>
          ))}
          {Array.from({ length: maxPlayers - playerCount }, (_, i) => (
            <div key={`empty-${i}`} className="flex items-center gap-2 bg-[#1a1a1b] rounded-lg px-4 py-2 opacity-40">
              <span className="w-2 h-2 rounded-full bg-gray-600" />
              <span className="text-gray-600 text-sm">Waiting...</span>
            </div>
          ))}
        </div>
      </div>

      <div className="flex items-center gap-2">
        <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
        <p className="text-gray-400 text-sm">
          {playerCount < 2
            ? 'Waiting for players to join...'
            : 'Starting game...'}
        </p>
      </div>
    </div>
  )
}
