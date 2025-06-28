# Create a comprehensive README
cat > README.md << 'EOF'
# Edge Insights IoT Platform

A real-time IoT log intelligence system with AI-powered analytics built with Go, TimescaleDB, and OpenAI.

## �� Features

- **Real-time IoT Log Ingestion**: WebSocket server for live device data
- **Time-series Database**: TimescaleDB for efficient log storage and querying
- **AI-Powered Analytics**: OpenAI embeddings for semantic search
- **Vector Similarity Search**: pgvector integration for intelligent log discovery
- **REST API**: Comprehensive endpoints for log retrieval and AI analysis




## 🛠️ Tech Stack

- **Backend**: Go
- **Database**: TimescaleDB Cloud
- **Vector Search**: pgvector
- **AI**: OpenAI Embeddings API
- **Real-time**: WebSocket
- **API**: REST endpoints

## 📋 Prerequisites

- Go 1.24+
- TimescaleDB Cloud account
- OpenAI API key
- pgvector extension enabled

## �� Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/edge-insights-iot-platform.git
   cd edge-insights-iot-platform
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your credentials
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```

5. **Test the endpoints**
   ```bash
   # Health check
   curl http://localhost:8080/health
   
   # Semantic search
   curl -X POST http://localhost:8080/api/ai/search \
     -H "Content-Type: application/json" \
     -d '{"search_text": "temperature sensor readings", "limit": 5}'
   ```

## �� API Endpoints

### Core Endpoints
- `GET /health` - Health check
- `GET /api/logs` - Get recent logs
- `GET /api/logs/device/{id}` - Get device-specific logs

### AI Endpoints
- `POST /api/ai/query` - Natural language log queries
- `POST /api/ai/search` - Semantic search using embeddings
- `POST /api/ai/summarize` - AI-powered log summaries
- `GET /api/ai/anomalies` - Anomaly detection

### WebSocket
- `ws://localhost:8080/ws` - Real-time IoT log ingestion

## 🧪 Testing

Run the IoT simulator to generate test data:
```bash
go run scripts/main.go
```

## �� Database Schema

- `device_logs` - Time-series table for IoT logs
- `device_logs_embedding_store` - Vector embeddings for semantic search

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details

## 👨‍💻 Author

Edge Insights Project
EOF

# Add and commit the README
git add README.md
git commit -m "Add comprehensive README.md"
git push