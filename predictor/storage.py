import os

import psycopg
from psycopg.types.json import Jsonb

DEFAULT_URL = "postgresql://predict:predict_dev_password@localhost:5432/predictions"


def connect():
    return psycopg.connect(os.environ.get("DATABASE_URL", DEFAULT_URL))

def save_prediction(conn, event_id, probabilities, model_version):
    cleaned = {}
    for option, p in probabilities.items():
        if not 0 <= p <= 1:
            raise ValueError("probabilities must be between 0 and 1")
        cleaned[option.strip().lower()] = round(p, 4)
    if abs(sum(cleaned.values()) - 1) > 0.001:
        raise ValueError("probabilities must add up to 1")

    with conn.cursor() as cur:
        cur.execute("SELECT options FROM events WHERE id = %s", (event_id,))
        row = cur.fetchone()
    conn.commit()
    if row is None:
        return "event_not_found"
    if set(cleaned) != set(row[0]):
        return "invalid_option"

    predicted = max(cleaned, key=cleaned.get)
    confidence = round(cleaned[predicted], 3)

    sql = """
        INSERT INTO system_predictions
            (event_id, predicted, confidence, probabilities, model_version)
        SELECT id, %s, %s, %s, %s FROM events
        WHERE id = %s AND status = 'open' AND locks_at > now()
        ON CONFLICT (event_id, model_version) DO NOTHING
    """
    with conn.cursor() as cur:
        cur.execute(sql, (predicted, confidence, Jsonb(cleaned), model_version, event_id))
        inserted = cur.rowcount == 1
    conn.commit()
    if inserted:
        return "saved"

    with conn.cursor() as cur:
        cur.execute(
            "SELECT status = 'open' AND locks_at > now() FROM events WHERE id = %s",
            (event_id,),
        )
        is_open = cur.fetchone()[0]
    conn.commit()
    return "already_exists" if is_open else "event_closed"