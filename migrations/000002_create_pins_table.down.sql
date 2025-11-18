-- Drop indexes
DROP INDEX IF EXISTS idx_pins_deleted_at;
DROP INDEX IF EXISTS idx_pins_location;
DROP INDEX IF EXISTS idx_pins_user_id;

-- Drop pins table
DROP TABLE IF EXISTS pins;
