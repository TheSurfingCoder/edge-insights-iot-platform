import { useEffect, useRef } from 'react'
import { Chart, registerables } from 'chart.js'

Chart.register(...registerables)

interface TimeSeriesChartProps {
  data: any[]
  title: string
  type: 'line' | 'bar'
  height?: number
}

export default function TimeSeriesChart({ data, title, type = 'line', height = 300 }: TimeSeriesChartProps) {
  const chartRef = useRef<HTMLCanvasElement>(null)
  const chartInstance = useRef<Chart | null>(null)

  // Smart chart type detection
  const detectChartType = (data: any[], numericColumns: string[]) => {
    if (!data || data.length === 0) return 'line'
    
    // Check if this looks like aggregated data (counts, totals)
    const hasCountData = numericColumns.some(col => 
      col.includes('count') || col.includes('Count') || col.includes('total') || col.includes('Total')
    )
    
    // Check if this looks like time-series data (averages, values over time)
    const hasTimeSeriesData = numericColumns.some(col => 
      col.includes('avg') || col.includes('value') || col.includes('Value') || 
      col.includes('min') || col.includes('max')
    )
    
    // Check data length - bar charts work better for fewer data points
    const isShortDataset = data.length <= 10
    
    if (hasCountData || isShortDataset) {
      return 'bar'
    } else if (hasTimeSeriesData) {
      return 'line'
    } else {
      return 'line' // default
    }
  }

  useEffect(() => {
    console.log('📊 Chart useEffect triggered:')
    console.log('   - chartRef.current:', !!chartRef.current)
    console.log('   - data:', data)
    console.log('   - data length:', data?.length)
    console.log('   - data type:', typeof data)
    
    if (!chartRef.current || !data || data.length === 0) {
      console.log('📊 Chart: No data or empty data array')
      return
    }

    console.log('📊 Chart Component Analysis:')
    console.log('   - Data length:', data.length)
    console.log('   - First row:', data[0])
    console.log('   - First row keys:', Object.keys(data[0] || {}))
    console.log('   - Data types:', Object.keys(data[0] || {}).map(key => ({
      key,
      type: typeof data[0][key],
      value: data[0][key]
    })))

    // Destroy existing chart
    if (chartInstance.current) {
      chartInstance.current.destroy()
    }

    const ctx = chartRef.current.getContext('2d')
    if (!ctx) return

    // Prepare chart data
    const labels = data.map(row => {
      // Find the time column - be more flexible
      let timeValue = null
      let timeColumnType = null
      
      // Check for known time column names first
      if (row.time) { timeValue = row.time; timeColumnType = 'time' }
      else if (row.five_min_bucket) { timeValue = row.five_min_bucket; timeColumnType = 'five_min_bucket' }
      else if (row.hour) { timeValue = row.hour; timeColumnType = 'hour' }
      else if (row.day) { timeValue = row.day; timeColumnType = 'day' }
      else {
        // Look for any column that might be a time column
        for (const key of Object.keys(row)) {
          const keyLower = key.toLowerCase()
          const value = row[key]
          
          if (keyLower.includes('time') || keyLower.includes('date') || 
              (typeof value === 'string' && (value.includes('T') || value.includes('-') || value.includes(':')))) {
            timeValue = value
            timeColumnType = key
            break
          }
        }
      }
      
      if (!timeValue) return 'Unknown'
      
      const date = new Date(timeValue)
      
      // Smart time formatting based on data granularity
      if (timeColumnType === 'five_min_bucket' || timeColumnType?.includes('five_min')) {
        // 5-minute data - show time only
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      } else if (timeColumnType === 'hour' || timeColumnType?.includes('hour')) {
        // Hourly data - show date and hour
        return date.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit' })
      } else if (timeColumnType === 'day' || timeColumnType?.includes('day')) {
        // Daily data - show date only
        return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
      } else {
        // Default - show full date/time
        return date.toLocaleString()
      }
    })

    const datasets: any[] = []
    
    // Enhanced numeric column detection for continuous aggregates
    const numericColumns = Object.keys(data[0]).filter(key => {
      const value = data[0][key]
      const isNumeric = typeof value === 'number' && !isNaN(value)
      const isTimeColumn = ['time', 'five_min_bucket', 'hour', 'day'].includes(key)
      const isCountColumn = key.includes('count') || key.includes('Count')
      
      return isNumeric && !isTimeColumn && !isCountColumn
    })

    console.log('📊 Data Analysis:')
    console.log('   - Time columns found:', Object.keys(data[0]).filter(key => ['time', 'five_min_bucket', 'hour', 'day'].includes(key)))
    console.log('   - Numeric columns found:', numericColumns)
    console.log('   - Count columns found:', Object.keys(data[0]).filter(key => key.includes('count') || key.includes('Count')))
    
    // Determine if this is time-series data
    const hasTimeColumn = Object.keys(data[0]).some(key => ['time', 'five_min_bucket', 'hour', 'day'].includes(key))
    const hasNumericData = numericColumns.length > 0
    
    console.log('   - Has time column:', hasTimeColumn)
    console.log('   - Has numeric data:', hasNumericData)
    console.log('   - Chartable:', hasTimeColumn && hasNumericData)

    // Use smart chart type detection
    const detectedChartType = detectChartType(data, numericColumns)
    const finalChartType = type === 'line' ? detectedChartType : type // Allow override
    
    console.log('   - Detected chart type:', detectedChartType)
    console.log('   - Final chart type:', finalChartType)

    if (numericColumns.length > 0) {
      numericColumns.forEach((column, index) => {
        const colors = [
          'rgb(59, 130, 246)', // blue
          'rgb(16, 185, 129)', // green
          'rgb(245, 158, 11)', // yellow
          'rgb(239, 68, 68)',  // red
          'rgb(139, 92, 246)', // purple
        ]
        
        // Enhanced label formatting
        let label = column
          .replace(/_/g, ' ')
          .replace(/\b\w/g, l => l.toUpperCase())
          .replace('Avg Value', 'Average')
          .replace('Min Value', 'Minimum')
          .replace('Max Value', 'Maximum')
        
        datasets.push({
          label,
          data: data.map(row => row[column]),
          borderColor: colors[index % colors.length],
          backgroundColor: colors[index % colors.length] + '20',
          borderWidth: 2,
          fill: type === 'line',
          tension: 0.1,
        })
      })
    }

    // If no numeric columns found, try to create a chart from count data
    if (datasets.length === 0) {
      const countColumns = Object.keys(data[0]).filter(key => {
        const value = data[0][key]
        return typeof value === 'number' && (key.includes('count') || key.includes('Count'))
      })

      if (countColumns.length > 0) {
        countColumns.forEach((column, index) => {
          const colors = [
            'rgb(59, 130, 246)', // blue
            'rgb(16, 185, 129)', // green
            'rgb(245, 158, 11)', // yellow
            'rgb(239, 68, 68)',  // red
          ]
          
          let label = column
            .replace(/_/g, ' ')
            .replace(/\b\w/g, l => l.toUpperCase())
            .replace('Error Count', 'Errors')
            .replace('Warning Count', 'Warnings')
            .replace('Info Count', 'Info Logs')
            .replace('Total Readings', 'Total')
          
          datasets.push({
            label,
            data: data.map(row => row[column]),
            borderColor: colors[index % colors.length],
            backgroundColor: colors[index % colors.length] + '20',
            borderWidth: 2,
            fill: type === 'line',
            tension: 0.1,
          })
        })
      }
    }

    // Create new chart
    chartInstance.current = new Chart(ctx, {
      type: finalChartType,
      data: {
        labels,
        datasets,
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          title: {
            display: true,
            text: `${title} (${finalChartType} chart)`,
            font: {
              size: 16,
              weight: 'bold',
            },
          },
          legend: {
            display: datasets.length > 1,
            position: 'top',
          },
        },
        scales: {
          x: {
            display: true,
            title: {
              display: true,
              text: 'Time',
            },
          },
          y: {
            display: true,
            title: {
              display: true,
              text: 'Value',
            },
          },
        },
        interaction: {
          intersect: false,
          mode: 'index',
        },
      },
    })

    return () => {
      if (chartInstance.current) {
        chartInstance.current.destroy()
      }
    }
  }, [data, title, type])

  if (!data || data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 bg-gray-50 rounded-md">
        <div className="text-center">
          <p className="text-gray-500 mb-2">No data available for chart</p>
          <p className="text-sm text-gray-400">Try running a different query</p>
        </div>
      </div>
    )
  }

  // Check if data is chartable - be more flexible with time column detection
  const allKeys = Object.keys(data[0] || {})
  const timeColumnKeys = ['time', 'five_min_bucket', 'hour', 'day', 'timestamp', 'created_at', 'updated_at']
  
  // Look for any column that might be a time column (contains time-related keywords or is a string that looks like a date)
  const hasTimeColumn = allKeys.some(key => {
    const keyLower = key.toLowerCase()
    const value = data[0][key]
    
    // Check if it's a known time column name
    if (timeColumnKeys.includes(keyLower)) return true
    
    // Check if the key contains time-related words
    if (keyLower.includes('time') || keyLower.includes('date') || keyLower.includes('created') || keyLower.includes('updated')) return true
    
    // Check if the value looks like a date string
    if (typeof value === 'string' && (value.includes('T') || value.includes('-') || value.includes(':'))) {
      const date = new Date(value)
      return !isNaN(date.getTime())
    }
    
    return false
  })
  
  const numericColumns = allKeys.filter(key => {
    const value = data[0][key]
    const isNumeric = typeof value === 'number' && !isNaN(value)
    const isTimeColumn = timeColumnKeys.includes(key.toLowerCase()) || 
                        key.toLowerCase().includes('time') || 
                        key.toLowerCase().includes('date')
    return isNumeric && !isTimeColumn
  })
  
  console.log('📊 Chart Validation:')
  console.log('   - All keys:', allKeys)
  console.log('   - Has time column:', hasTimeColumn)
  console.log('   - Numeric columns:', numericColumns)
  console.log('   - Time column keys:', timeColumnKeys)
  
  if (!hasTimeColumn || numericColumns.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 bg-gray-50 rounded-md">
        <div className="text-center">
          <p className="text-gray-500 mb-2">Data not suitable for charting</p>
          <p className="text-sm text-gray-400">
            {!hasTimeColumn ? `Missing time column. Available columns: ${allKeys.join(', ')}` : 'No numeric data to plot'}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div style={{ height }}>
      <canvas ref={chartRef} />
    </div>
  )
}
