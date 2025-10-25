package subpackage

import (
	"fmt"
)

type HistoricalEvaluator struct {
	secretModifier float64 // Внутренний фактор, который меняется
	callCount      int     // Счетчик вызовов (история)
}

func NewHistoricalEvaluator() *HistoricalEvaluator {
	fmt.Println("Инициализировано: Секретный модификатор = 10.0")
	return &HistoricalEvaluator{
		secretModifier: 10.0,
		callCount:      0,
	}
}

//go:generate testifai --func=CalculateComplexScore --type=table --output=CalculateComplexScore_ai_test.go
func (h *HistoricalEvaluator) CalculateComplexScore(input float64) float64 {
	h.callCount++

	baseScore := input + h.secretModifier // 7.0

	if input == 7.0 {
		h.secretModifier = -15.0
		fmt.Printf("\n   -> ВНИМАНИЕ: Триггер (%.1f)! Модификатор изменился до: %.1f\n", input, h.secretModifier)
	}

	historyFactor := 1.0 + (float64(h.callCount) * 0.05) // -1.15

	finalScore := baseScore * historyFactor //

	fmt.Printf("   Вызов %d. Ввод: %.1f. Модификатор: %.1f. Базовый: %.2f. Фактор истории: %.2f. Результат: %.2f\n",
		h.callCount, input, h.secretModifier, baseScore, historyFactor, finalScore)

	return finalScore
}
