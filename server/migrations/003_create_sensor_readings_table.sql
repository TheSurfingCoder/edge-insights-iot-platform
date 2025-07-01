-- Create the new sensor_readings table with proper structure
CREATE TABLE IF NOT EXISTS sensor_readings (
    time TIMESTAMPTZ NOT NULL,
    device_id TEXT NOT NULL,
    device_type TEXT NOT NULL,
    location TEXT,
    raw_value NUMERIC,
    unit TEXT,
    message TEXT
);

-- Convert to hypertable for time-series optimization
SELECT create_hypertable('sensor_readings', 'time', if_not_exists => TRUE);