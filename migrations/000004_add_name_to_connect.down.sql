-- Drop index
DROP INDEX IF EXISTS idx_connect_name;

-- Remove name column from connect table
ALTER TABLE connect DROP COLUMN IF EXISTS name;
