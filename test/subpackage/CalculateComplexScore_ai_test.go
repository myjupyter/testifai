// ✅ Tests for CalculateComplexScore method of HistoricalEvaluator

package subpackage

import (
	"fmt"
	"testing"
)

func TestHistoricalEvaluator_CalculateComplexScore(t *testing.T) {
	tests := []struct {
		name           string
		initialModifier float64
		input          float64
		expectedScore  float64
		expectedCalls  int
		expectedModifier float64
		triggerActivated bool // indicates if the special case for input=7.0 should activate
	}{
		{
			name:             "First call with zero input",
			initialModifier:  0.0,
			input:            0.0,
			expectedScore:    0.0,
			expectedCalls:    1,
			expectedModifier: 0.0,
			triggerActivated: false,
		},
		{
			name:             "First call with positive input",
			initialModifier:  0.0,
			input:            5.0,
			expectedScore:    5.25,
			expectedCalls:    1,
			expectedModifier: 0.0,
			triggerActivated: false,
		},
		{
			name:             "First call with negative input",
			initialModifier:  0.0,
			input:            -3.0,
			expectedScore:    -2.85,
			expectedCalls:    1,
			expectedModifier: 0.0,
			triggerActivated: false,
		},
		{
			name:             "Trigger input 7.0 changes modifier",
			initialModifier:  0.0,
			input:            7.0,
			expectedScore:    -8.05,
			expectedCalls:    1,
			expectedModifier: -15.0,
			triggerActivated: true,
		},
		// ⚠️ !!!ATTENTION!!! Test case where input is 7.0 but initial modifier is already -15.0
		{
			name:             "Input 7.0 with already changed modifier",
			initialModifier:  -15.0,
			input:            7.0,
			expectedScore:    -8.8,
			expectedCalls:    1,
			expectedModifier: -15.0,
			triggerActivated: true,
		},
		// ⚠️ Test multiple calls to check history factor accumulation
		{
			name:             "Multiple calls with same evaluator",
			initialModifier:  0.0,
			input:            10.0,
			expectedScore:    10.5,
			expectedCalls:    1,
			expectedModifier: 0.0,
			triggerActivated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			he := &HistoricalEvaluator{
				secretModifier: tt.initialModifier,
				callCount:      0,
			}

			score := he.CalculateComplexScore(tt.input)

			if fmt.Sprintf("%.2f", score) != fmt.Sprintf("%.2f", tt.expectedScore) {
				t.Errorf("CalculateComplexScore() = %.2f, want %.2f", score, tt.expectedScore)
			}

			if he.callCount != tt.expectedCalls {
				t.Errorf("callCount = %d, want %d", he.callCount, tt.expectedCalls)
			}

			if he.secretModifier != tt.expectedModifier {
				t.Errorf("secretModifier = %.1f, want %.1f", he.secretModifier, tt.expectedModifier)
			}
		})
	}
}

// ⚠️ Test sequence of calls to verify internal state changes correctly over time
func TestHistoricalEvaluator_SequenceOfCalls(t *testing.T) {
	he := &HistoricalEvaluator{
		secretModifier: 0.0,
		callCount:      0,
	}

	// First call
	score1 := he.CalculateComplexScore(5.0)
	expected1 := 5.25
	if fmt.Sprintf("%.2f", score1) != fmt.Sprintf("%.2f", expected1) {
		t.Errorf("First call score = %.2f, want %.2f", score1, expected1)
	}
	if he.callCount != 1 {
		t.Errorf("After first call, callCount = %d, want 1", he.callCount)
	}

	// Second call
	score2 := he.CalculateComplexScore(3.0)
	expected2 := 3.3
	if fmt.Sprintf("%.2f", score2) != fmt.Sprintf("%.2f", expected2) {
		t.Errorf("Second call score = %.2f, want %.2f", score2, expected2)
	}
	if he.callCount != 2 {
		t.Errorf("After second call, callCount = %d, want 2", he.callCount)
	}

	// Third call with trigger
	score3 := he.CalculateComplexScore(7.0)
	expected3 := -8.8
	if fmt.Sprintf("%.2f", score3) != fmt.Sprintf("%.2f", expected3) {
		t.Errorf("Third call score = %.2f, want %.2f", score3, expected3)
	}
	if he.callCount != 3 {
		t.Errorf("After third call, callCount = %d, want 3", he.callCount)
	}
	if he.secretModifier != -15.0 {
		t.Errorf("After third call, secretModifier = %.1f, want -15.0", he.secretModifier)
	}

	// Fourth call after trigger
	score4 := he.CalculateComplexScore(2.0)
	expected4 := -12.4
	if fmt.Sprintf("%.2f", score4) != fmt.Sprintf("%.2f", expected4) {
		t.Errorf("Fourth call score = %.2f, want %.2f", score4, expected4)
	}
	if he.callCount != 4 {
		t.Errorf("After fourth call, callCount = %d, want 4", he.callCount)
	}
}