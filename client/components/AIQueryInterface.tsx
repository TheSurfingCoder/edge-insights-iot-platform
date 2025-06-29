import { useState } from 'react'

export default function AIQueryInterface() {
  const [query, setQuery] = useState('')
  const [result, setResult] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!query.trim()) return

    setLoading(true)
    try {
      const response = await fetch('https://edge-insights-iot-platform-production.up.railway.app/api/ai/query', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ query }),
      })
      
      const data = await response.json()
      if (data.success) {
        setResult(JSON.stringify(data.result, null, 2))
      } else {
        setResult('Error: ' + data.error)
      }
    } catch (error) {
      setResult('Error: ' + (error as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-6">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="query" className="block text-sm font-medium text-gray-700">
            Ask about your IoT logs
          </label>
          <input
            type="text"
            id="query"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="e.g., Show me temperature sensor readings from the last hour"
            className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 p-2"
          />
        </div>
        <button
          type="submit"
          disabled={loading}
          className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? 'Querying...' : 'Ask AI'}
        </button>
      </form>
      
      {result && (
        <div className="mt-6">
          <h3 className="text-lg font-medium text-gray-900 mb-2">Result:</h3>
          <pre className="bg-gray-100 p-4 rounded-md text-sm overflow-x-auto">
            {result}
          </pre>
        </div>
      )}
    </div>
  )
}
