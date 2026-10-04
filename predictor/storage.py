import os

import psycopg

DEFAULT_URL = "postgresql://predict:predict_dev_password@localhost:5432/predictions"


def connect():
    return psycopg.connect(os.environ.get("DATABASE_URL", DEFAULT_URL))


def save_prediction(conn, event_id, predicted, confidence, model_version):
    if not 0 <= confidence <= 1:
        raise ValueError("confidence must be between 0 and 1")

    sql = """
        INSERT INTO system_predictions (event_id, predicted, confidence, model_version)
        SELECT id, %s, %s, %s FROM events
        WHERE id = %s AND status = 'open' AND locks_at > now()
        ON CONFLICT (event_id, model_version) DO NOTHING
    """
    with conn.cursor() as cur:
        cur.execute(sql, (predicted, confidence, model_version, event_id))
        inserted = cur.rowcount == 1
    conn.commit()
    if inserted:
        return "saved"

    with conn.cursor() as cur:
        cur.execute(
            "SELECT status = 'open' AND locks_at > now() FROM events WHERE id = %s",
            (event_id,),
        )
        row = cur.fetchone()
    if row is None:
        return "event_not_found"
    if not row[0]:
        return "event_closed"
    return "already_exists"