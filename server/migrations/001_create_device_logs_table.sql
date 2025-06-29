-- Create the device_logs hypertable
CREATE TABLE IF NOT EXISTS device_logs (
    time TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL,
    log_type TEXT NOT NULL,
    message TEXT NOT NULL
);

-- Convert to hypertable
SELECT create_hypertable('device_logs', 'time', if_not_exists => TRUE);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_device_logs_device_id ON device_logs (device_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_logs_log_type ON device_logs (log_type, time DESC);

-- used for initial setup - it will build all my tables
-- used when i need to change schema. I would create a new migration file each time i need to make a change to the database
-- ensures everyone database stays in sync
-- deployments. when i run this file, it will stay consistent everwhere 