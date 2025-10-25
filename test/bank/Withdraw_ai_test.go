// ✅ Tests for Withdraw method covering all logical branches and edge cases
package bank

import (
	"testing"
	"time"
)

func TestAccount_Withdraw(t *testing.T) {
	tests := []struct {
		name            string
		initialBalance  float64
		initialTrans    int
		initialLastOp   time.Time
		initialVip      bool
		withdrawAmount  float64
		currentHour     int // Mocking time.Now().Hour()
		timeSinceLastOp time.Duration
		expectedAmount  float64
		expectError     bool
		expectedBalance float64
		expectedTrans   int
		expectedVip     bool
	}{
		{
			name:            "Basic withdrawal with 1% fee",
			initialBalance:  1000.0,
			initialTrans:    0,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      false,
			withdrawAmount:  100.0,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  101.0,
			expectError:     false,
			expectedBalance: 899.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Withdrawal at night - double fee",
			initialBalance:  1000.0,
			initialTrans:    0,
			initialLastOp:   now(23).Add(-2 * time.Second),
			initialVip:      false,
			withdrawAmount:  100.0,
			currentHour:     23,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  102.0,
			expectError:     false,
			expectedBalance: 898.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Withdrawal during early morning - double fee",
			initialBalance:  1000.0,
			initialTrans:    0,
			initialLastOp:   now(5).Add(-2 * time.Second),
			initialVip:      false,
			withdrawAmount:  100.0,
			currentHour:     5,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  102.0,
			expectError:     false,
			expectedBalance: 898.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Transaction divisible by 5 - no 1% fee",
			initialBalance:  1000.0,
			initialTrans:    4,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      false,
			withdrawAmount:  100.0,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  100.0,
			expectError:     false,
			expectedBalance: 900.0,
			expectedTrans:   5,
			expectedVip:     false,
		},
		{
			name:            "VIP status revoked due to large withdrawal",
			initialBalance:  1000.0,
			initialTrans:    0,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      true,
			withdrawAmount:  600.0,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  606.0,
			expectError:     false,
			expectedBalance: 394.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Fast consecutive transactions - additional $10 fee",
			initialBalance:  1000.0,
			initialTrans:    0,
			initialLastOp:   now(12).Add(-500 * time.Millisecond),
			initialVip:      false,
			withdrawAmount:  100.0,
			currentHour:     12,
			timeSinceLastOp: 500 * time.Millisecond,
			expectedAmount:  111.0,
			expectError:     false,
			expectedBalance: 889.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "VIP discount applied on every 3rd transaction",
			initialBalance:  1000.0,
			initialTrans:    2,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      true,
			withdrawAmount:  100.0,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  100.5,
			expectError:     false,
			expectedBalance: 899.5,
			expectedTrans:   3,
			expectedVip:     true,
		},
		{
			name:            "Insufficient funds after fees",
			initialBalance:  100.0,
			initialTrans:    0,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      false,
			withdrawAmount:  99.5,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  0.0,
			expectError:     true,
			expectedBalance: 100.0,
			expectedTrans:   0, // Not incremented in case of error
			expectedVip:     false,
		},
		{
			name:            "Exact balance match after fees",
			initialBalance:  101.0,
			initialTrans:    0,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      false,
			withdrawAmount:  100.0,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  101.0,
			expectError:     false,
			expectedBalance: 0.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Zero withdrawal",
			initialBalance:  1000.0,
			initialTrans:    0,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      false,
			withdrawAmount:  0.0,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  0.0,
			expectError:     false,
			expectedBalance: 1000.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Withdrawal with rounding up",
			initialBalance:  1000.0,
			initialTrans:    0,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      false,
			withdrawAmount:  99.999,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  101.0, // Rounded up from 100.999
			expectError:     false,
			expectedBalance: 899.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		// ⚠️ !!!ATTENTION!!! Test case for edge case where transaction count is divisible by both 3 and 5
		{
			name:            "Transaction count divisible by 3 and 5 - no 1% fee but VIP discount",
			initialBalance:  1000.0,
			initialTrans:    14,
			initialLastOp:   now(12).Add(-2 * time.Second),
			initialVip:      true,
			withdrawAmount:  100.0,
			currentHour:     12,
			timeSinceLastOp: 2 * time.Second,
			expectedAmount:  100.0, // No 1% fee, VIP discount on zero fee = still zero
			expectError:     false,
			expectedBalance: 900.0,
			expectedTrans:   15,
			expectedVip:     true,
		},
		// ⚠️ Test case for night time with fast transaction
		{
			name:            "Night time with fast transaction - both fees apply",
			initialBalance:  1000.0,
			initialTrans:    0,
			initialLastOp:   now(23).Add(-500 * time.Millisecond),
			initialVip:      false,
			withdrawAmount:  100.0,
			currentHour:     23,
			timeSinceLastOp: 500 * time.Millisecond,
			expectedAmount:  112.0, // 100 + (100*0.01*2) + 10
			expectError:     false,
			expectedBalance: 888.0,
			expectedTrans:   1,
			expectedVip:     false,
		},
		// ⚠️ Test case for VIP discount with night time and fast transaction
		{
			name:            "VIP discount with night and fast fees",
			initialBalance:  1000.0,
			initialTrans:    2,
			initialLastOp:   now(23).Add(-500 * time.Millisecond),
			initialVip:      true,
			withdrawAmount:  100.0,
			currentHour:     23,
			timeSinceLastOp: 500 * time.Millisecond,
			expectedAmount:  106.0, // 100 + ((100*0.01*2)*0.5) + 10
			expectError:     false,
			expectedBalance: 894.0,
			expectedTrans:   3,
			expectedVip:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			account := &Account{
				balance:       tt.initialBalance,
				transactions:  tt.initialTrans,
				lastOperation: tt.initialLastOp,
				vipStatus:     tt.initialVip,
			}

			// Adjust lastOperation to simulate timeSinceLastOp
			account.lastOperation = time.Now().Add(-tt.timeSinceLastOp)

			amount, err := account.Withdraw(tt.withdrawAmount)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if amount != tt.expectedAmount {
					t.Errorf("expected amount %.2f, got %.2f", tt.expectedAmount, amount)
				}
			}

			if account.balance != tt.expectedBalance {
				t.Errorf("expected balance %.2f, got %.2f", tt.expectedBalance, account.balance)
			}
			if account.transactions != tt.expectedTrans {
				t.Errorf("expected transactions %d, got %d", tt.expectedTrans, account.transactions)
			}
			if account.vipStatus != tt.expectedVip {
				t.Errorf("expected VIP status %v, got %v", tt.expectedVip, account.vipStatus)
			}
		})
	}
}

var originalNow = time.Now()

func now(currentHour int) time.Time {
	return time.Date(2023, 1, 1, currentHour, 0, 0, 0, time.UTC)
}
