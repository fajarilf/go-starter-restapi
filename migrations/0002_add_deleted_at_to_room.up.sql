ALTER TABLE rooms ADD COLUMN deleted_at TIMESTAMPTZ NULL;
CREATE INDEX idx_rooms_active ON rooms (id) WHERE deleted_at IS NULL;
