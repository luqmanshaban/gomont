CREATE TABLE IF NOT EXISTS users(
     id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
     email TEXT NULL UNIQUE,
     names TEXT NULL,
     verification_code INT NULL,
     verification_expiry TIMESTAMPTZ NULL,
     created_at TIMESTAMPTZ NOT NULL default NOW(),
     updated_at TIMESTAMPTZ NOT NULL default NOW()
);

CREATE TABLE IF NOT EXISTS notification_channels(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emails TEXT[] NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL default NOW()
);


CREATE TABLE IF NOT EXISTS urls(
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    is_healthy BOOLEAN NOT NULL default false,
    notification_sent BOOLEAN NOT NULL default false,
    max_retries INT NOT NULL default 4,
    interval INT NOT NULL, 
    retry_at TIMESTAMPTZ NULL,
    last_manual_retry_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL default NOW(),
    updated_at TIMESTAMPTZ NOT NULL default NOW()
);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    notification_channels_id  BIGINT NOT NULL REFERENCES notification_channels(id),
    url_id BIGINT NOT NULL REFERENCES urls(id),
    error_mesasge TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL default NOW()
);

-- Create the Trigger Function for auto-populating `notification_channels`
CREATE OR REPLACE FUNCTION initialize_user_notification_channels()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO notification_channels (user_id, emails)
    VALUES (NEW.id, ARRAY[NEW.email]);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 4. Bind the Trigger Function to the Users Table
-- Using a conditional check so running your migration repeatedly won't fail
DROP TRIGGER IF EXISTS after_user_created ON users;

CREATE TRIGGER after_user_created
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION initialize_user_notification_channels();