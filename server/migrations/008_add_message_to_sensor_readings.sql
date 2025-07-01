-- Add message column to sensor_readings table
ALTER TABLE sensor_readings ADD COLUMN IF NOT EXISTS message TEXT;

