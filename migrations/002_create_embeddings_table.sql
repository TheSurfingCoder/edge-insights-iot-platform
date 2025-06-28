-- Create the embeddings table for vector search
CREATE TABLE IF NOT EXISTS device_logs_embedding_store (
    embedding_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    time TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL,
    chunk_seq INTEGER NOT NULL,
    chunk TEXT NOT NULL,
    embedding vector(1536) NOT NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_embeddings_device_id ON device_logs_embedding_store (device_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_embeddings_time ON device_logs_embedding_store (time DESC);

-- Enable vector similarity search
CREATE INDEX IF NOT EXISTS idx_embeddings_vector ON device_logs_embedding_store USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);