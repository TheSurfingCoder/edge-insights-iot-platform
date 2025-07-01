import type { NextApiRequest, NextApiResponse } from 'next'

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' })
  }

  try {
    const response = await fetch(`${process.env.API_URL}/api/ai/query`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(req.body),
    })

    const data = await response.json()
    res.status(response.status).json(data)
  } catch (error) {
    console.error('API proxy error:', error)
    res.status(500).json({ 
      success: false,
      error: 'Internal server error',
      query: req.body?.query || '',
      time: new Date().toISOString()
    })
  }
} 