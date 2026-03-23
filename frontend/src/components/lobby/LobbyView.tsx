import { useState } from 'react'

interface LobbyViewProps {
  roomCode: string
  playerName: string
  playerCount: number
}

export function LobbyView({ roomCode, playerName, playerCount }: LobbyViewProps) {
  const [copied, setCopied] = useState(false)

  const copyCode = async () => {
    await navigator.clipboard.writeText(roomCode)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="flex flex-col items-center gap-8 py-12">
      <div className="text-center">
        <p className="text-gray-400 text-sm mb-2">Share this code with your friend</p>
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

      <div className="flex flex-col items-center gap-2">
        <p className="text-gray-300">
          Playing as <span className="font-semibold text-white">{playerName}</span>
        </p>
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
          <p className="text-gray-400 text-sm">
            {playerCount < 2
              ? 'Waiting for opponent to join...'
              : 'Starting game...'}
          </p>
        </div>
      </div>
    </div>
  )
}
