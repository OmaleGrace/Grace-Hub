DEFAULT_RATING = 1500
HOME_ADVANTAGE = 65
K = 20


def expected_home_score(home_rating, away_rating):
    diff = (home_rating + HOME_ADVANTAGE) - away_rating
    return 1 / (1 + 10 ** (-diff / 400))


def predict(home_rating, away_rating):
    p_home = expected_home_score(home_rating, away_rating)
    if p_home >= 0.5:
        return "home", p_home
    return "away", 1 - p_home


def update(home_rating, away_rating, home_score):
    expected = expected_home_score(home_rating, away_rating)
    change = K * (home_score - expected)
    return home_rating + change, away_rating - change