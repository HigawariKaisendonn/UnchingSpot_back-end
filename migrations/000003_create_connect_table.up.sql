-- Create connect table
CREATE TABLE connect (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pins_id_1 UUID NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
    pins_id_2 UUID NOT NULL REFERENCES pins(id) ON DELETE CASCADE,
    show BOOLEAN NOT NULL DEFAULT true
);

-- Create indexes
CREATE INDEX idx_connect_user_id ON connect(user_id);
CREATE INDEX idx_connect_pins ON connect(pins_id_1, pins_id_2);
