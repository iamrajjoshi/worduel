import type { ConnectionStatus as Status } from '../../hooks/useWebSocket'

const statusConfig: Record<Status, { color: string; label: string }> = {
  connected: { color: 'bg-green-500', label: 'Connected' },
  connecting: { color: 'bg-yellow-500 animate-pulse', label: 'Connecting...' },
  reconnecting: { color: 'bg-yellow-500 animate-pulse', label: 'Reconnecting...' },
  disconnected: { color: 'bg-red-500', label: 'Disconnected' },
  failed: { color: 'bg-red-500', label: 'Connection failed' },
}

interface ConnectionStatusProps {
  status: Status
}

export function ConnectionStatus({ status }: ConnectionStatusProps) {
  const config = statusConfig[status]

  if (status === 'connected') {
    return null // Don't show when connected
  }

  return (
    <div className="flex items-center gap-2 text-xs text-gray-400">
      <span className={`w-2 h-2 rounded-full ${config.color}`} />
      {config.label}
    </div>
  )
}
