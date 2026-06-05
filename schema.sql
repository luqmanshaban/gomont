CREATE TABLE IF NOT EXISTS users(
     id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
     email TEXT NULL,
     names TEXT NULL,
     verification_code INT NULL,
     verification_expiry TIMESTAMPTZ NULL,
     created_at TIMESTAMPTZ NOT NULL default NOW(),
     updated_at TIMESTAMPTZ NOT NULL default NOW()
);
