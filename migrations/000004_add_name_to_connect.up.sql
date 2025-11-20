-- Add name column to connect table
ALTER TABLE connect ADD COLUMN name TEXT NOT NULL DEFAULT '';

-- Create index for name search
CREATE INDEX idx_connect_name ON connect(name);
