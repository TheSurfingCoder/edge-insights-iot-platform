import { useState } from 'react'
import { ChartBarIcon, MagnifyingGlassIcon, LightBulbIcon } from '@heroicons/react/24/outline'

interface SQLResult {
  sql: string
  explanation: string
  result: Record<string, unknown>[]
  row_count: number
}

interface LogEntry {
  time: string
  device_id: string
  device_type: string
  location: string
  log_type: string
  raw_value: string
  unit: string
  distance: number
}

interface SemanticResult {
  relevant_logs: LogEntry[]
}

interface QueryResult {
  success: boolean
  result: SQLResult | SemanticResult | Record<string, unknown>
  query: string
  time: string
  error?: string
}

interface QuerySuggestion {
  text: string
  type: 'data' | 'pattern'
  description: string
}

export default function AIQueryInterface() {
  const [query, setQuery] = useState('')
  const [result, setResult] = useState<QueryResult | null>(null)
  const [loading, setLoading] = useState(false)

  // Query suggestions for users
  const querySuggestions: QuerySuggestion[] = [
    {
      text: "Show me temperature readings from warehouse_a in the last hour",
      type: 'data',
      description: "Get specific sensor data"
    },
    {
      text: "What's the average humidity over the last 24 hours?",
      type: 'data', 
      description: "Calculate aggregated statistics"
    },
    {
      text: "Find logs about temperature problems",
      type: 'pattern',
      description: "Discover patterns and issues"
    },
    {
      text: "Show me unusual security events",
      type: 'pattern',
      description: "Find anomalies and alerts"
    },
    {
      text: "Which devices had readings above 30°C yesterday?",
      type: 'data',
      description: "Filter data by conditions"
    },
    {
      text: "Find logs similar to HVAC failures",
      type: 'pattern',
      description: "Semantic search for related issues"
    }
  ]

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!query.trim()) return

    setLoading(true)
    
    // Dispatch chart loading event
    const loadingEvent = new CustomEvent('chart-loading', {
      detail: {
        type: 'chart_loading'
      }
    })
    window.dispatchEvent(loadingEvent)
    
    try {
      const response = await fetch('/api/ai/query', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ query }),
      })
      
      const data = await response.json()
      setResult(data)

      // Automatically show chart for text-to-SQL results
      console.log('📊 FULL API RESPONSE ANALYSIS:')
      console.log('   - Success:', data.success)
      console.log('   - Has result:', !!data.result)
      console.log('   - Has sql:', !!data.result?.sql)
      console.log('   - Has result.result:', !!data.result?.result)
      console.log('   - Result type:', typeof data.result?.result)
      console.log('   - Result length:', data.result?.result?.length)
      
      if (data.success && data.result && 'sql' in data.result && data.result.sql && 'result' in data.result && data.result.result) {
        console.log('📊 Chart Data Analysis:')
        console.log('   - Data type: text-to-SQL')
        console.log('   - Row count:', data.result.result.length)
        console.log('   - First row:', data.result.result[0])
        console.log('   - Column names:', Object.keys(data.result.result[0] || {}))
        console.log('   - Is chartable:', data.result.result.length > 0 && data.result.result[0])
        
        // Only dispatch chart if we have valid data
        if (data.result.result.length > 0 && data.result.result[0]) {
          console.log('📊 Dispatching chart data:', data.result.result)
          // Dispatch custom event to show chart
          const chartEvent = new CustomEvent('chart-data', {
            detail: {
              type: 'chart_data',
              data: data.result.result,
              title: `Query Results: ${query}`
            }
          })
          window.dispatchEvent(chartEvent)
          console.log('📊 Chart event dispatched successfully')
        } else {
          console.log('⚠️  No chartable data found - skipping chart dispatch')
        }
      } else {
        console.log('📊 No text-to-SQL result - skipping chart dispatch')
        console.log('   - Reason: Missing required fields')
      }

      console.log('🔍 Full API response:', data);
    } catch (error) {
      setResult({
        success: false,
        result: {},
        query: query,
        time: new Date().toISOString(),
        error: (error as Error).message
      })
    } finally {
      setLoading(false)
    }
  }

  const handleSuggestionClick = (suggestion: QuerySuggestion) => {
    setQuery(suggestion.text)
  }

  const renderResult = () => {
    if (!result) return null

    if (!result.success) {
      return (
        <div className="mt-6 p-4 bg-red-50 border border-red-200 rounded-md">
          <h3 className="text-lg font-medium text-red-900 mb-2">Error:</h3>
          <p className="text-red-700">{result.error}</p>
        </div>
      )
    }

    // Check if it's a text-to-SQL result
    if ('sql' in result.result && result.result.sql) {
      const sqlResult = result.result as SQLResult
      return (
        <div className="mt-6 space-y-4">
          <div className="p-4 bg-blue-50 border border-blue-200 rounded-md">
            <h3 className="text-lg font-medium text-blue-900 mb-2">
              <ChartBarIcon className="inline w-5 h-5 mr-2" />
              Data Query Results
            </h3>
            <p className="text-blue-700 mb-3">{sqlResult.explanation}</p>
            <details className="text-sm">
              <summary className="cursor-pointer text-blue-600 hover:text-blue-800">
                View Generated SQL
              </summary>
              <pre className="mt-2 p-2 bg-blue-100 rounded text-xs overflow-x-auto">
                {sqlResult.sql}
              </pre>
            </details>
          </div>
          
          <div className="bg-white border border-gray-200 rounded-md overflow-hidden">
            <div className="px-4 py-2 bg-gray-50 border-b border-gray-200">
              <span className="text-sm font-medium text-gray-700">
                Results ({sqlResult.row_count || 0} rows)
              </span>
            </div>
            <div className="max-h-96 overflow-y-auto">
              {sqlResult.result && sqlResult.result.length > 0 ? (
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      {Object.keys(sqlResult.result[0]).map(key => (
                        <th key={key} className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          {key}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {sqlResult.result.map((row: Record<string, unknown>, index: number) => (
                      <tr key={index}>
                        {Object.values(row).map((value: unknown, valueIndex: number) => (
                          <td key={valueIndex} className="px-4 py-2 text-sm text-gray-900">
                            {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              ) : (
                <div className="p-8 text-center text-gray-500">
                  <p>No data found for this query.</p>
                  <p className="text-sm mt-2">Try adjusting your search criteria or time range.</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )
    }

    // Check if it's a semantic search result
    if ('relevant_logs' in result.result && result.result.relevant_logs) {
      const semanticResult = result.result as SemanticResult
      const logs = semanticResult.relevant_logs;
      console.log('🔍 Semantic search results:', logs);
      console.log('🔍 First log entry:', logs[0]);
      console.log('🔍 First log keys:', Object.keys(logs[0]));
      console.log('🔍 Log type value:', logs[0]?.log_type);
      console.log('🔍 All field values for first log:', {
        time: logs[0]?.time,
        device_id: logs[0]?.device_id,
        device_type: logs[0]?.device_type,
        location: logs[0]?.location,
        log_type: logs[0]?.log_type,
        raw_value: logs[0]?.raw_value,
        unit: logs[0]?.unit,
        distance: logs[0]?.distance
      });
      const columns = ['time', 'device_id', 'device_type', 'location', 'log_type', 'raw_value', 'unit', 'distance'];

      return (
        <div className="bg-white border border-gray-200 rounded-md overflow-hidden mt-6">
          <div className="px-4 py-2 bg-gray-50 border-b border-gray-200">
            <span className="text-sm font-medium text-gray-700">
              Relevant Logs ({logs.length} found)
            </span>
          </div>
          <div className="max-h-96 overflow-y-auto">
            {logs.length > 0 ? (
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    {columns.map(key => (
                      <th key={key} className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        {key === 'distance' ? 'Similarity Score' : key.replace(/_/g, ' ')}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {logs.map((row: LogEntry, index: number) => (
                    <tr key={index}>
                      {columns.map((key, valueIndex) => (
                        <td key={valueIndex} className={`px-4 py-2 text-sm ${
                          key === 'distance' ? 'font-medium' : 'text-gray-900'
                        } ${
                          key === 'distance' ? 
                            (1 - row.distance) * 100 > 80 ? 'text-green-600' : 
                            (1 - row.distance) * 100 > 60 ? 'text-yellow-600' : 'text-red-600' 
                          : ''
                        }`}>
                          {(() => {
                            const value = row[key as keyof LogEntry];
                            if (value === null || value === undefined) {
                              return 'N/A';
                            }
                            
                            // Format similarity score (distance) as percentage
                            if (key === 'distance' && typeof value === 'number') {
                              const similarity = Math.max(0, (1 - value) * 100);
                              return `${similarity.toFixed(1)}%`;
                            }
                            
                            // Format time if it's a timestamp
                            if (key === 'time' && typeof value === 'string') {
                              try {
                                const date = new Date(value);
                                return date.toLocaleString();
                              } catch {
                                return value;
                              }
                            }
                            
                            // Handle objects and other types
                            return typeof value === 'object' ? JSON.stringify(value) : String(value);
                          })()}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div className="p-8 text-center text-gray-500">
                <p>No relevant logs found.</p>
                <p className="text-sm mt-2">Try adjusting your search criteria.</p>
              </div>
            )}
          </div>
        </div>
      )
    }

    // Fallback for other result types
    return (
      <div className="mt-6 p-4 bg-gray-50 border border-gray-200 rounded-md">
        <h3 className="text-lg font-medium text-gray-900 mb-2">Result:</h3>
        <pre className="text-sm overflow-x-auto">
          {JSON.stringify(result.result, null, 2)}
        </pre>
      </div>
    )
  }

  return (
    <div className="p-6">
      {/* Query Input */}
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label htmlFor="query" className="block text-sm font-medium text-gray-700 mb-2">
            Ask about your IoT data
          </label>
          <div className="flex space-x-2">
            <input
              type="text"
              id="query"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="e.g., Show me temperature readings from warehouse_a in the last hour"
              className="flex-1 border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 p-3"
            />
            <button
              type="submit"
              disabled={loading}
              className="bg-blue-600 text-white px-6 py-3 rounded-md hover:bg-blue-700 disabled:opacity-50 flex items-center"
            >
              {loading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Querying...
                </>
              ) : (
                <>
                  <MagnifyingGlassIcon className="w-4 h-4 mr-2" />
                  Ask AI
                </>
              )}
            </button>
          </div>
        </div>
      </form>

      {/* Query Suggestions */}
      <div className="mt-6">
        <h3 className="text-sm font-medium text-gray-700 mb-3">Try these queries:</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {querySuggestions.map((suggestion, index) => (
            <button
              key={index}
              onClick={() => handleSuggestionClick(suggestion)}
              className="text-left p-3 border border-gray-200 rounded-md hover:border-blue-300 hover:bg-blue-50 transition-colors"
            >
              <div className="flex items-center mb-1">
                {suggestion.type === 'data' ? (
                  <ChartBarIcon className="w-4 h-4 text-blue-600 mr-2" />
                ) : (
                  <LightBulbIcon className="w-4 h-4 text-green-600 mr-2" />
                )}
                <span className="text-sm font-medium text-gray-900">
                  {suggestion.type === 'data' ? 'Data Query' : 'Pattern Search'}
                </span>
              </div>
              <p className="text-sm text-gray-700 mb-1">{suggestion.text}</p>
              <p className="text-xs text-gray-500">{suggestion.description}</p>
            </button>
          ))}
        </div>
      </div>

      {/* Results */}
      {renderResult()}
    </div>
  )
}
