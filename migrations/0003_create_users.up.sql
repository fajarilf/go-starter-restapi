CREATE TABLE IF NOT EXISTS users (
    id            SERIAL      PRIMARY KEY,
    username      TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ponytail: seed a default admin user so auth works out of the box.
-- Rotate the password in production via env/secret manager.
INSERT INTO users (username, password_hash)
VALUES ('admin', '$2a$10$3yQmNm0U93S.tWlm3nf7NuPki4JMX9ZN3zo.EE4hpbL.edqg57Cta')
ON CONFLICT (username) DO NOTHING;
