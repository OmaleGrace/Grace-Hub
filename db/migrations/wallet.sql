CREATE TABLE wallet_entries (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    currency    TEXT NOT NULL DEFAULT 'points' CHECK (currency IN ('points')),
    amount      BIGINT NOT NULL CHECK (amount <> 0),
    reason      TEXT NOT NULL CHECK (reason IN
                ('signup_bonus', 'stake', 'stake_refund', 'payout', 'correction')),
    event_id    BIGINT REFERENCES events(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_wallet_user ON wallet_entries(user_id);
CREATE UNIQUE INDEX one_signup_bonus ON wallet_entries(user_id) WHERE reason = 'signup_bonus';

CREATE FUNCTION wallet_entries_immutable() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'wallet_entries rows cannot be changed or deleted';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER wallet_entries_no_change
  BEFORE UPDATE OR DELETE ON wallet_entries
  FOR EACH ROW EXECUTE FUNCTION wallet_entries_immutable();

INSERT INTO wallet_entries (user_id, amount, reason)
SELECT id, 100000, 'signup_bonus' FROM users
ON CONFLICT DO NOTHING;