import type { NextApiRequest, NextApiResponse } from 'next'

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' })
  }

  // Check if API_URL is configured
  if (!process.env.API_URL) {
    console.error('API_URL environment variable is not set')
    return res.status(500).json({ 
      success: false,
      error: 'API_URL not configured',
      query: req.body?.query || '',
      time: new Date().toISOString()
    })
  }

  try {
    // Ensure API_URL has proper protocol
    let apiUrl = process.env.API_URL
    if (!apiUrl.startsWith('http://') && !apiUrl.startsWith('https://')) {
      apiUrl = `https://${apiUrl}`
    }
    
    const fullUrl = `${apiUrl}/api/ai/query`
    console.log('Proxying request to:', fullUrl)
    console.log('Request body:', req.body)
    
    const response = await fetch(fullUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(req.body),
    })

    console.log('Railway response status:', response.status)
    
    if (!response.ok) {
      console.error('Railway API error:', response.status, response.statusText)
      return res.status(response.status).json({
        success: false,
        error: `Railway API error: ${response.status} ${response.statusText}`,
        query: req.body?.query || '',
        time: new Date().toISOString()
      })
    }

    const data = await response.json()
    console.log('Railway response data:', data)
    res.status(response.status).json(data)
  } catch (error) {
    console.error('API proxy error:', error)
    res.status(500).json({ 
      success: false,
      error: error instanceof Error ? error.message : 'Internal server error',
      query: req.body?.query || '',
      time: new Date().toISOString()
    })
  }
} 