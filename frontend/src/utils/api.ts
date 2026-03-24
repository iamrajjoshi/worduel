interface CreateRoomResponse {
  roomId: string
  roomCode: string
  name: string
  createdAt: string
}

export async function createRoom(maxPlayers = 2): Promise<CreateRoomResponse> {
  const res = await fetch('/api/rooms', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name: 'Game Room', maxPlayers }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Failed to create room' }))
    throw new Error(err.message || 'Failed to create room')
  }
  return res.json()
}
