package rates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAverageCalculatorCalculate(t *testing.T) {
	tests := []struct {
		name    string
		levels  []SpotLevel
		n       int
		m       int
		want    float64
		wantErr bool
	}{
		{
			name: "returns average for range",
			levels: []SpotLevel{
				{Price: 100},
				{Price: 110},
				{Price: 120},
				{Price: 130},
			},
			n:    1,
			m:    3,
			want: 120,
		},
		{
			name: "fails when range exceeds bounds",
			levels: []SpotLevel{
				{Price: 100},
				{Price: 110},
			},
			n:       1,
			m:       2,
			wantErr: true,
		},
		{
			name: "fails when range is reversed",
			levels: []SpotLevel{
				{Price: 100},
				{Price: 110},
			},
			n:       1,
			m:       0,
			wantErr: true,
		},
		{
			name: "fails when range starts below zero",
			levels: []SpotLevel{
				{Price: 100},
				{Price: 110},
			},
			n:       -1,
			m:       0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := NewAverageCalculator(tt.n, tt.m).Calculate(tt.levels)
			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, value)
		})
	}
}
