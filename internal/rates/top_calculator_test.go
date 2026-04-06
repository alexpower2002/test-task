package rates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTopCalculatorCalculate(t *testing.T) {
	tests := []struct {
		name    string
		levels  []SpotLevel
		n       int
		want    float64
		wantErr bool
	}{
		{
			name: "returns nth price",
			levels: []SpotLevel{
				{Price: 101},
				{Price: 102},
				{Price: 103},
			},
			n:    1,
			want: 102,
		},
		{
			name: "fails when out of bounds",
			levels: []SpotLevel{
				{Price: 100},
			},
			n:       1,
			wantErr: true,
		},
		{
			name: "fails when n is invalid",
			levels: []SpotLevel{
				{Price: 100},
			},
			n:       -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := NewTopCalculator(tt.n).Calculate(tt.levels)

			if tt.wantErr {
				assert.Error(t, err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, value)
		})
	}
}
