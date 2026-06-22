CREATE TABLE IF NOT EXISTS rooms (
    id          SERIAL      PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NULL
)