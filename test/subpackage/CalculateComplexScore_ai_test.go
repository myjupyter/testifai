// ✅ Tests for CalculateComplexScore method of HistoricalEvaluator

package subpackage

import (
	"testing"
)

func TestHistoricalEvaluator_CalculateComplexScore(t *testing.T) {
	tests := []struct {
		name            string
		initialModifier float64
		inputs          []float64
		expectedScores  []float64
		expectedCalls   []int
		// ⚠️ !!!ATTENTION!!! Test case where input triggers secretModifier change
	}{
		{
			name:            "Single call with non-trigger value",
			initialModifier: 0.0,
			inputs:          []float64{5.0},
			expectedScores:  []float64{5.25},
			expectedCalls:   []int{1},
		},
		{
			name:            "Multiple calls with non-trigger values",
			initialModifier: 0.0,
			inputs:          []float64{1.0, 2.0, 3.0},
			expectedScores:  []float64{1.05, 2.2, 3.45},
			expectedCalls:   []int{1, 2, 3},
		},
		{
			name:            "Trigger value changes modifier",
			initialModifier: 0.0,
			inputs:          []float64{7.0},
			expectedScores:  []float64{-8.4}, // (7 + (-15)) * 1.05
			expectedCalls:   []int{1},
		},
		{
			name:            "Trigger followed by normal calls",
			initialModifier: 0.0,
			inputs:          []float64{7.0, 10.0, 20.0},
			expectedScores:  []float64{-8.4, -5.6, 5.25}, // Second uses updated modifier and call count
			expectedCalls:   []int{1, 2, 3},
		},
		{
			name:            "Start with negative modifier",
			initialModifier: -2.0,
			inputs:          []float64{5.0, 7.0, 3.0},
			expectedScores:  []float64{2.9, -17.85, -12.6}, // Third uses new modifier and updated call count
			expectedCalls:   []int{1, 2, 3},
		},
		{
			name:            "Zero input with zero modifier",
			initialModifier: 0.0,
			inputs:          []float64{0.0},
			expectedScores:  []float64{0.0},
			expectedCalls:   []int{1},
		},
		{
			name:            "Large input value",
			initialModifier: 0.0,
			inputs:          []float64{100.0},
			expectedScores:  []float64{105.0},
			expectedCalls:   []int{1},
		},
		{
			name:            "Negative input value",
			initialModifier: 0.0,
			inputs:          []float64{-5.0},
			expectedScores:  []float64{-5.25},
			expectedCalls:   []int{1},
		},
		// ⚠️ !!!ATTENTION!!! Edge case with multiple triggers in sequence
		{
			name:            "Multiple trigger inputs",
			initialModifier: 0.0,
			inputs:          []float64{7.0, 7.0, 7.0},
			expectedScores:  []float64{-8.4, -16.8, -16.8}, // Modifier changes only on first trigger
			expectedCalls:   []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluator := &HistoricalEvaluator{
				secretModifier: tt.initialModifier,
				callCount:      0,
			}

			for i, input := range tt.inputs {
				score := evaluator.CalculateComplexScore(input)
				expected := tt.expectedScores[i]
				callCount := evaluator.callCount

				if score != expected {
					t.Errorf("CalculateComplexScore(%v) = %v, want %v", input, score, expected)
				}

				if callCount != tt.expectedCalls[i] {
					t.Errorf("callCount after input %v = %v, want %v", input, callCount, tt.expectedCalls[i])
				}
			}
		})
	}
}
