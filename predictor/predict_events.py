import ratings
import storage
from models import elo

MODEL_VERSION = "elo-v2"


def predict_open_events(conn):
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT e.id, e.details->>'home', e.details->>'away'
            FROM events e
            WHERE e.category = 'sports' AND e.status = 'open' AND e.locks_at > now()
              AND NOT EXISTS (
                  SELECT 1 FROM system_predictions sp WHERE sp.event_id = e.id
              )
            ORDER BY e.opens_at, e.id
            """
        )
        todo = cur.fetchall()
    conn.commit()

    counts = {}
    for event_id, home, away in todo:
        if not home or not away:
            counts["missing_teams"] = counts.get("missing_teams", 0) + 1
            continue

        with conn.cursor() as cur:
            home_rating = ratings.get_rating(cur, ratings.normalize(home))
            away_rating = ratings.get_rating(cur, ratings.normalize(away))
        conn.commit()

        probs = elo.probabilities(home_rating, away_rating)
        result = storage.save_prediction(conn, event_id, probs, MODEL_VERSION)
        counts[result] = counts.get(result, 0) + 1
    return counts