-- Remove message column from sensor_readings table
ALTER TABLE sensor_readings DROP COLUMN IF EXISTS message;