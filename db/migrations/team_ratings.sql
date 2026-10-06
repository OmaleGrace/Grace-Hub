CREATE TABLE team_ratings (
    team            TEXT PRIMARY KEY,
    rating          NUMERIC(8,2) NOT NULL DEFAULT 1500,
    matches_played  INTEGER NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE events ADD COLUMN rated_at TIMESTAMPTZ;