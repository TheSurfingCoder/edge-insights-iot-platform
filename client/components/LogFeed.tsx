import { useState, useEffect } from 'react'
import { format } from 'date-fns'

interface Log {
  time: string
  device_id: string
  log_type: string
  message: string
}

export default function LogFeed() {
  const [logs, setLogs] = useState<Log[]>([])
  const [connected, setConnected] = useState(false)

  useEffect(() => {
    const ws = new WebSocket('wss://edge-insights-iot-platform-production.up.railway.app/ws')
    
    ws.onopen = () => {
      setConnected(true)
      console.log('Connected to WebSocket')
    }
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.success) {
          setLogs(prev => [data, ...prev.slice(0, 49)]) // Keep last 50 logs
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error)
      }
    }
    
    ws.onclose = () => {
      setConnected(false)
      console.log('Disconnected from WebSocket')
    }
    
    return () => ws.close()
  }, [])

  const getLogTypeColor = (type: string) => {
    switch (type) {
      case 'ERROR': return 'text-red-600 bg-red-100'
      case 'WARN': return 'text-yellow-600 bg-yellow-100'
      case 'INFO': return 'text-blue-600 bg-blue-100'
      case 'DEBUG': return 'text-gray-600 bg-gray-100'
      default: return 'text-gray-600 bg-gray-100'
    }
  }

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-4">
        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
          connected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
        }`}>
          {connected ? 'Connected' : 'Disconnected'}
        </span>
        <span className="text-sm text-gray-500">{logs.length} logs</span>
      </div>
      
      <div className="space-y-3 max-h-96 overflow-y-auto">
        {logs.map((log, index) => (
          <div key={index} className="border-l-4 border-gray-200 pl-4 py-2">
            <div className="flex items-center justify-between">
              <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${getLogTypeColor(log.log_type)}`}>
                {log.log_type}
              </span>
              <span className="text-xs text-gray-500">
                {format(new Date(log.time), 'HH:mm:ss')}
              </span>
            </div>
            <p className="text-sm font-medium text-gray-900 mt-1">{log.device_id}</p>
            <p className="text-sm text-gray-600 mt-1">{log.message}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
