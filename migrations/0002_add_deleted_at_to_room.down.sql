DROP INDEX IF EXISTS idx_rooms_active;
ALTER TABLE rooms DROP COLUMN deleted_at;
