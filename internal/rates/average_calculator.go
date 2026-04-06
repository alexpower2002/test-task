package rates

import "fmt"

type AverageCalculator struct {
	n int
	m int
}

func NewAverageCalculator(n, m int) Calculator {
	return &AverageCalculator{
		n: n,
		m: m,
	}
}

func (c *AverageCalculator) Calculate(levels []SpotLevel) (float64, error) {
	if c.n < 0 || c.m >= len(levels) || c.n > c.m {
		return 0, fmt.Errorf("avgNM range [%d;%d] out of range", c.n, c.m)
	}

	var sum float64

	for _, level := range levels[c.n : c.m+1] {
		sum += level.Price
	}

	return sum / float64(c.m-c.n+1), nil
}
