package predictions

import "math"

const (
	Margin  = 1.0
	MinOdds = 1.01
	MaxOdds = 50.0
)

func OddsFor(probs map[string]float64) map[string]float64 {
	odds := make(map[string]float64, len(probs))
	for option, p := range probs {
		if p <= 0 {
			continue
		}
		o := 1 / (p * Margin)
		o = math.Floor(o*100) / 100
		if o < MinOdds {
			o = MinOdds
		}
		if o > MaxOdds {
			o = MaxOdds
		}
		odds[option] = o
	}
	return odds
}