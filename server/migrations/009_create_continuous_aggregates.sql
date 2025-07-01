-- Create hierarchical continuous aggregates for sensor data analytics
-- This creates a pyramid structure: 5min -> hourly -> daily for optimal performance

-- Level 1: 5-minute aggregates from raw sensor_readings (base level)
CREATE MATERIALIZED VIEW IF NOT EXISTS five_min_sensor_averages
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('5 minutes', time) AS five_min_bucket,
    device_type,
    location,
    avg(raw_value) as avg_value,
    min(raw_value) as min_value,
    max(raw_value) as max_value,
    count(*) as reading_count
FROM sensor_readings
WHERE raw_value IS NOT NULL
GROUP BY five_min_bucket, device_type, location;

-- Level 2: Hourly aggregates built on top of 5-minute aggregates
CREATE MATERIALIZED VIEW IF NOT EXISTS hourly_sensor_averages
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', five_min_bucket) AS hour,
    device_type,
    location,
    avg(avg_value) as avg_value,
    min(min_value) as min_value,
    max(max_value) as max_value,
    sum(reading_count) as reading_count
FROM five_min_sensor_averages
GROUP BY hour, device_type, location;

-- Level 3: Daily aggregates built on top of hourly aggregates
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_sensor_averages
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', hour) AS day,
    device_type,
    location,
    avg(avg_value) as avg_value,
    min(min_value) as min_value,
    max(max_value) as max_value,
    sum(reading_count) as reading_count
FROM hourly_sensor_averages
GROUP BY day, device_type, location;

-- Separate daily device activity summary (for log analysis)
-- This can be built directly from sensor_readings since it's counting log types
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_device_activity
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', time) AS day,
    device_type,
    location,
    count(*) as total_readings,
    count(CASE WHEN log_type IN ('ERROR', 'CRITICAL') THEN 1 END) as error_count,
    count(CASE WHEN log_type = 'WARNING' THEN 1 END) as warning_count,
    count(CASE WHEN log_type = 'INFO' THEN 1 END) as info_count
FROM sensor_readings
GROUP BY day, device_type, location;

-- Add refresh policies for automatic updates
-- Refresh 5-minute aggregates every 1 minute (base level)
SELECT add_continuous_aggregate_policy('five_min_sensor_averages',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

-- Refresh hourly aggregates every 5 minutes (depends on 5-min level)
SELECT add_continuous_aggregate_policy('hourly_sensor_averages',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '5 minutes');

-- Refresh daily aggregates every hour (depends on hourly level)
SELECT add_continuous_aggregate_policy('daily_sensor_averages',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 hour');

-- Refresh daily device activity every hour
SELECT add_continuous_aggregate_policy('daily_device_activity',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 hour');

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_five_min_sensor_averages_bucket 
ON five_min_sensor_averages (five_min_bucket DESC);

CREATE INDEX IF NOT EXISTS idx_hourly_sensor_averages_hour 
ON hourly_sensor_averages (hour DESC);

CREATE INDEX IF NOT EXISTS idx_daily_sensor_averages_day 
ON daily_sensor_averages (day DESC);

CREATE INDEX IF NOT EXISTS idx_daily_device_activity_day 
ON daily_device_activity (day DESC);

-- Add additional indexes for filtering
CREATE INDEX IF NOT EXISTS idx_five_min_sensor_averages_device_location 
ON five_min_sensor_averages (device_type, location, five_min_bucket DESC);

CREATE INDEX IF NOT EXISTS idx_hourly_sensor_averages_device_location 
ON hourly_sensor_averages (device_type, location, hour DESC);

CREATE INDEX IF NOT EXISTS idx_daily_sensor_averages_device_location 
ON daily_sensor_averages (device_type, location, day DESC); 