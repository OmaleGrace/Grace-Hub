DEFAULT_RATING = 1500
HOME_ADVANTAGE = 65
DRAW_MAX = 0.28
K = 20


def expected_home_score(home_rating, away_rating):
    diff = (home_rating + HOME_ADVANTAGE) - away_rating
    return 1 / (1 + 10 ** (-diff / 400))

def probabilities(home_rating, away_rating):
    e = expected_home_score(home_rating, away_rating)
    closeness = 1 - abs(2 * e - 1)
    p_draw = DRAW_MAX * closeness
    p_home = e - p_draw / 2
    p_away = (1 - e) - p_draw / 2
    return {"home": p_home, "draw": p_draw, "away": p_away}

def predict(home_rating, away_rating):
    probs = probabilities(home_rating, away_rating)
    pick = max(probs, key=probs.get)
    return pick, probs[pick]


def update(home_rating, away_rating, home_score):
    expected = expected_home_score(home_rating, away_rating)
    change = K * (home_score - expected)
    return home_rating + change, away_rating - change