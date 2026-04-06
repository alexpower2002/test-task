package rates

import "fmt"

type TopCalculator struct {
	n int
}

func NewTopCalculator(n int) Calculator {
	return &TopCalculator{n: n}
}

func (c *TopCalculator) Calculate(levels []SpotLevel) (float64, error) {
	if c.n < 0 || c.n >= len(levels) {
		return 0, fmt.Errorf("topN index %d out of range", c.n)
	}

	return levels[c.n].Price, nil
}
