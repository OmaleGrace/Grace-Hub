CREATE TABLE bets (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    event_id    BIGINT NOT NULL REFERENCES events(id),
    choice      TEXT NOT NULL,
    stake       BIGINT NOT NULL CHECK (stake > 0),
    odds        NUMERIC(6,2) NOT NULL CHECK (odds >= 1.01),
    status      TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'won', 'lost')),
    payout      BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    settled_at  TIMESTAMPTZ,
    UNIQUE (user_id, event_id)
);

CREATE INDEX idx_bets_event ON bets(event_id);

ALTER TABLE wallet_entries ADD COLUMN bet_id BIGINT REFERENCES bets(id);