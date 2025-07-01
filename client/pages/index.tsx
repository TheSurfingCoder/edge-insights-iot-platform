import { useState, useEffect } from 'react'
import LogFeed from '@/components/LogFeed'
import DeviceStatus from '@/components/DeviceStatus'
import AIQueryInterface from '@/components/AIQueryInterface'
import TimeSeriesChart from '@/components/TimeSeriesChart'

export default function Dashboard() {
  const [chartData, setChartData] = useState<any[]>([])
  const [chartTitle, setChartTitle] = useState('')
  const [showChart, setShowChart] = useState(false)
  const [chartLoading, setChartLoading] = useState(false)

  // Listen for chart data from AI query results
  useEffect(() => {
    const handleChartData = (event: CustomEvent) => {
      console.log('📊 Chart event received:', event.detail)
      if (event.detail && event.detail.type === 'chart_data') {
        console.log('📊 Chart data received:', event.detail.data)
        console.log('📊 Chart title:', event.detail.title)
        console.log('📊 Data length:', event.detail.data?.length)
        console.log('📊 First row:', event.detail.data?.[0])
        setChartData(event.detail.data)
        setChartTitle(event.detail.title)
        setShowChart(true)
        setChartLoading(false)
        console.log('📊 Chart state updated - should be visible now')
      } else {
        console.log('⚠️  Invalid chart event received:', event.detail)
      }
    }

    const handleChartLoading = (event: CustomEvent) => {
      if (event.detail && event.detail.type === 'chart_loading') {
        setChartLoading(true)
        setShowChart(false)
      }
    }

    window.addEventListener('chart-data', handleChartData as EventListener)
    window.addEventListener('chart-loading', handleChartLoading as EventListener)
    return () => {
      window.removeEventListener('chart-data', handleChartData as EventListener)
      window.removeEventListener('chart-loading', handleChartLoading as EventListener)
    }
  }, [])

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">
          Edge Insights Dashboard
        </h1>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Real-time Log Feed */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">Live Log Feed</h2>
            </div>
            <LogFeed />
          </div>

          {/* Device Status */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-xl font-semibold text-gray-900">Device Status</h2>
            </div>
            <DeviceStatus />
          </div>
        </div>

        {/* Dynamic Chart Display */}
        {(showChart || chartLoading) && (
          <div className="mt-8 bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <div className="flex justify-between items-center">
                <h2 className="text-xl font-semibold text-gray-900">
                  {chartLoading ? 'Generating Chart...' : 'Query Results Chart'}
                </h2>
                {showChart && (
                  <button
                    onClick={() => setShowChart(false)}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    ✕
                  </button>
                )}
              </div>
            </div>
            <div className="p-6">
              {chartLoading ? (
                <div className="flex items-center justify-center h-64">
                  <div className="text-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
                    <p className="text-gray-600">Processing query and generating chart...</p>
                  </div>
                </div>
              ) : (
                <TimeSeriesChart 
                  data={chartData} 
                  title={chartTitle}
                  type="line"
                  height={400}
                />
              )}
            </div>
          </div>
        )}

        {/* AI Query Interface */}
        <div className="mt-8 bg-white rounded-lg shadow">
          <div className="p-6 border-b border-gray-200">
            <h2 className="text-xl font-semibold text-gray-900">AI Query Interface</h2>
          </div>
          <AIQueryInterface />
        </div>
      </div>
    </div>
  )
}
