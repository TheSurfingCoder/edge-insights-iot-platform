import { useState, useEffect } from 'react'
import LogFeed from '@/components/LogFeed'
import DeviceStatus from '@/components/DeviceStatus'
import AIQueryInterface from '@/components/AIQueryInterface'

export default function Dashboard() {
  const [metrics, setMetrics] = useState({ recordsPerSecond: 0, totalLogs: 0 })

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">
          Edge Insights Dashboard
        </h1>
        
        {/* Metrics Overview */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Records/Second</h3>
            <p className="text-3xl font-bold text-blue-600">{metrics.recordsPerSecond}</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Total Logs</h3>
            <p className="text-3xl font-bold text-green-600">{metrics.totalLogs}</p>
          </div>
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-medium text-gray-900">Active Devices</h3>
            <p className="text-3xl font-bold text-purple-600">5</p>
          </div>
        </div>

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
