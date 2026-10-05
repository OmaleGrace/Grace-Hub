from models import elo

OUTCOME_SCORE = {"home": 1.0, "draw": 0.5, "away": 0.0}


def normalize(name):
    return " ".join(name.lower().split())


def get_rating(cur, team):
    cur.execute("SELECT rating FROM team_ratings WHERE team = %s", (team,))
    row = cur.fetchone()
    return float(row[0]) if row else elo.DEFAULT_RATING


def save_rating(cur, team, rating):
    cur.execute(
        """
        INSERT INTO team_ratings (team, rating, matches_played)
        VALUES (%s, %s, 1)
        ON CONFLICT (team) DO UPDATE
        SET rating = EXCLUDED.rating,
            matches_played = team_ratings.matches_played + 1,
            updated_at = now()
        """,
        (team, round(rating, 2)),
    )


def apply_resolved_events(conn):
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT id, details->>'home', details->>'away', outcome
            FROM events
            WHERE category = 'sports' AND status = 'resolved'
              AND rated_at IS NULL AND outcome IN ('home', 'draw', 'away')
            ORDER BY resolved_at, id
            """
        )
        todo = cur.fetchall()
    conn.commit()

    applied = 0
    for event_id, home, away, outcome in todo:
        if not home or not away:
            continue
        home, away = normalize(home), normalize(away)

        with conn.transaction():
            with conn.cursor() as cur:
                cur.execute(
                    "UPDATE events SET rated_at = now() WHERE id = %s AND rated_at IS NULL",
                    (event_id,),
                )
                if cur.rowcount != 1:
                    continue

                home_rating = get_rating(cur, home)
                away_rating = get_rating(cur, away)
                new_home, new_away = elo.update(
                    home_rating, away_rating, OUTCOME_SCORE[outcome]
                )
                save_rating(cur, home, new_home)
                save_rating(cur, away, new_away)
        applied += 1
    return applied