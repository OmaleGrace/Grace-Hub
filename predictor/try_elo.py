from models import elo

print(elo.predict(1500, 1500))
print(elo.predict(1600, 1500))
print(elo.predict(1400, 1600))

new_home, new_away = elo.update(1600, 1500, 1)
print(round(new_home, 1), round(new_away, 1))