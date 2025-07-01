-- Add primary key to sensor_readings table
-- Since this is a time-series table, we'll use a composite primary key
ALTER TABLE sensor_readings ADD CONSTRAINT sensor_readings_pkey 
PRIMARY KEY (time, device_id);

-- Note: If you prefer a UUID primary key instead, use this alternative:
-- ALTER TABLE sensor_readings ADD COLUMN id UUID DEFAULT gen_random_uuid();
-- ALTER TABLE sensor_readings ADD CONSTRAINT sensor_readings_pkey PRIMARY KEY (id);