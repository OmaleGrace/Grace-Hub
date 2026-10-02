CREATE TABLE events (
    id             BIGSERIAL PRIMARY KEY,
    category       TEXT NOT NULL CHECK (category IN ('sports', 'stocks')),
    title          TEXT NOT NULL,
    details        JSONB NOT NULL DEFAULT '{}',
    opens_at       TIMESTAMPTZ NOT NULL,
    locks_at       TIMESTAMPTZ NOT NULL,
    resolves_at    TIMESTAMPTZ NOT NULL,
    status         TEXT NOT NULL DEFAULT 'open'
                   CHECK (status IN ('open', 'locked', 'awaiting_resolution', 'resolved')),
    outcome        TEXT,
    resolved_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (opens_at < locks_at AND locks_at <= resolves_at)
);

CREATE TABLE system_predictions (
    id             BIGSERIAL PRIMARY KEY,
    event_id       BIGINT NOT NULL REFERENCES events(id),
    predicted      TEXT NOT NULL,
    confidence     NUMERIC(4,3) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    model_version  TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, model_version)
);