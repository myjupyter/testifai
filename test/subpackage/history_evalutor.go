package subpackage

import (
	"fmt"
)

// HistoricalEvaluator — структура, хранящая внутреннее состояние
// для нетривиального расчета.
type HistoricalEvaluator struct {
	secretModifier float64 // Внутренний фактор, который меняется
	callCount      int     // Счетчик вызовов (история)
}

// NewHistoricalEvaluator — конструктор (идиома Go)
func NewHistoricalEvaluator() *HistoricalEvaluator {
	fmt.Println("Инициализировано: Секретный модификатор = 10.0")
	return &HistoricalEvaluator{
		secretModifier: 10.0,
		callCount:      0,
	}
}

// CalculateComplexScore — метод, выдающий неочевидные результаты.
// Результат зависит от: 1) input, 2) secretModifier, 3) callCount.
//
//go:generate testifai --func=CalculateComplexScore --type=table --output=CalculateComplexScore_ai_test.go
func (h *HistoricalEvaluator) CalculateComplexScore(input float64) float64 {
	h.callCount++

	// 1. Базовый расчет, зависящий от внутреннего состояния (Модификатора)
	baseScore := input + h.secretModifier

	// 2. Неочевидный триггер и изменение состояния (Скрытое Правило)
	// Если входное значение равно 7.0, модификатор резко и нелинейно меняется.
	if input == 7.0 {
		h.secretModifier = -15.0 // Смена знака И увеличение по модулю
		fmt.Printf("\n   -> ВНИМАНИЕ: Триггер (%.1f)! Модификатор изменился до: %.1f\n", input, h.secretModifier)
	}

	// 3. Фактор истории (зависимость от того, сколько раз метод вызывался)
	// С каждым вызовом балл увеличивается на 5% от базового
	historyFactor := 1.0 + (float64(h.callCount) * 0.05)

	// 4. Финальный результат
	finalScore := baseScore * historyFactor

	fmt.Printf("   Вызов %d. Ввод: %.1f. Модификатор: %.1f. Базовый: %.2f. Фактор истории: %.2f. Результат: %.2f\n",
		h.callCount, input, h.secretModifier, baseScore, historyFactor, finalScore)

	return finalScore
}
