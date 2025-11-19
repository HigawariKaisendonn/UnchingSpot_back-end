-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create pins table
CREATE TABLE pins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location GEOMETRY(Point, 4326) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    edit_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_pins_user_id ON pins(user_id);
CREATE INDEX idx_pins_location ON pins USING GIST(location);
CREATE INDEX idx_pins_deleted_at ON pins(deleted_at);
