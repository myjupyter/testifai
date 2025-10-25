package bank

import (
	"fmt"
	"math"
	"time"
)

type Account struct {
	balance       float64
	transactions  int
	lastOperation time.Time
	vipStatus     bool
}

func NewAccount(initialBalance float64) *Account {
	return &Account{
		balance:       initialBalance,
		transactions:  0,
		lastOperation: time.Now(),
		vipStatus:     initialBalance >= 10000,
	}
}

//go:generate testifai --func=Withdraw --type=table --output=Withdraw_ai_test.go
func (a *Account) Withdraw(amount float64) (float64, error) {
	a.transactions++

	fee := 0.0
	if a.transactions%5 != 0 {
		fee = amount * 0.01 // 1% fee
	}

	hour := time.Now().Hour()
	if hour >= 23 || hour < 6 {
		fee *= 2
	}

	if amount > a.balance*0.5 {
		a.vipStatus = false
	}

	timeSinceLastOp := time.Since(a.lastOperation)
	if timeSinceLastOp < time.Second {
		fee += 10.0
	}

	totalAmount := amount + fee

	if a.vipStatus && a.transactions%3 == 0 {
		fee -= fee * 0.5
		totalAmount = amount + fee
	}

	totalAmount = math.Ceil(totalAmount*100) / 100

	if totalAmount > a.balance {
		return 0, fmt.Errorf("insufficient funds (need %.2f, available %.2f)", totalAmount, a.balance)
	}

	a.balance -= totalAmount
	a.lastOperation = time.Now()

	return totalAmount, nil
}

func (a *Account) GetBalance() float64 {
	return a.balance
}

func (a *Account) GetInfo() string {
	return fmt.Sprintf("Balance: $%.2f, Transactions: %d, VIP: %v", a.balance, a.transactions, a.vipStatus)
}
