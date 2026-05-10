ALTER TABLE pm_samples 
ALTER COLUMN collected_at TYPE timestamptz USING collected_at AT TIME ZONE 'UTC';