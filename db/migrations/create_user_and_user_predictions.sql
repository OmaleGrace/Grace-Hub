CREATE TABLE users (
    id             BIGSERIAL PRIMARY KEY,
    username       TEXT NOT NULL UNIQUE,
    email          TEXT NOT NULL UNIQUE,
    password_hash  TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_predictions (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id),
    event_id       BIGINT NOT NULL REFERENCES events(id),
    predicted      TEXT NOT NULL,
    confidence     NUMERIC(4,3) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    score          NUMERIC,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, event_id)
);

CREATE INDEX idx_user_predictions_event ON user_predictions(event_id);