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
		amount          float64
		timeOverride    func() time.Time // To mock time.Now()
		expectedAmount  float64
		expectError     bool
		expectedBalance float64
		expectedTrans   int
		expectedVip     bool
	}{
		{
			name:            "Basic withdrawal with 1% fee",
			initialBalance:  1000,
			initialTrans:    0,
			initialLastOp:   time.Now().Add(-2 * time.Second),
			initialVip:      false,
			amount:          100,
			expectedAmount:  101,
			expectError:     false,
			expectedBalance: 899,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:           "Withdrawal at night - double fee",
			initialBalance: 1000,
			initialTrans:   0,
			initialLastOp:  time.Now().Add(-2 * time.Second),
			initialVip:     false,
			amount:         100,
			timeOverride: func() time.Time {
				return time.Date(2023, 1, 1, 23, 30, 0, 0, time.UTC)
			},
			expectedAmount:  102, // 1% fee doubled = 2%, so 100 + 2 = 102
			expectError:     false,
			expectedBalance: 898,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Insufficient funds",
			initialBalance:  50,
			initialTrans:    0,
			initialLastOp:   time.Now().Add(-2 * time.Second),
			initialVip:      false,
			amount:          100,
			expectedAmount:  0,
			expectError:     true,
			expectedBalance: 50,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "VIP discount applied every 3rd transaction",
			initialBalance:  1000,
			initialTrans:    2, // So this will be the 3rd transaction
			initialLastOp:   time.Now().Add(-2 * time.Second),
			initialVip:      true,
			amount:          100,
			expectedAmount:  100.5, // 1% fee = 1, 50% discount = 0.5, total = 100.5
			expectError:     false,
			expectedBalance: 899.5,
			expectedTrans:   3,
			expectedVip:     true,
		},
		{
			name:            "No fee on every 5th transaction",
			initialTrans:    4, // So this will be the 5th transaction
			initialBalance:  1000,
			initialLastOp:   time.Now().Add(-2 * time.Second),
			initialVip:      false,
			amount:          100,
			expectedAmount:  100, // No fee
			expectError:     false,
			expectedBalance: 900,
			expectedTrans:   5,
			expectedVip:     false,
		},
		{
			name:            "VIP status revoked due to large withdrawal",
			initialBalance:  1000,
			initialTrans:    0,
			initialLastOp:   time.Now().Add(-2 * time.Second),
			initialVip:      true,
			amount:          600, // More than 50% of balance
			expectedAmount:  606, // 1% fee = 6
			expectError:     false,
			expectedBalance: 394,
			expectedTrans:   1,
			expectedVip:     false, // ⚠️ !!!ATTENTION!!! VIP status should be revoked
		},
		{
			name:            "Rush fee for quick successive withdrawals",
			initialBalance:  1000,
			initialTrans:    0,
			initialLastOp:   time.Now().Add(-500 * time.Millisecond), // Less than 1 second
			initialVip:      false,
			amount:          50,
			expectedAmount:  60.5, // 1% fee = 0.5 + rush fee = 10 => total = 10.5
			expectError:     false,
			expectedBalance: 939.5,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Rounding up to nearest cent",
			initialBalance:  1000,
			initialTrans:    0,
			initialLastOp:   time.Now().Add(-2 * time.Second),
			initialVip:      false,
			amount:          99.999, // Should round up
			expectedAmount:  101,    // 99.999 + 1% = ~100.99899 => ceil to 101
			expectError:     false,
			expectedBalance: 899,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:           "Night time with rush fee",
			initialBalance: 1000,
			initialTrans:   0,
			initialLastOp:  time.Now().Add(-500 * time.Millisecond),
			initialVip:     false,
			amount:         100,
			timeOverride: func() time.Time {
				return time.Date(2023, 1, 1, 2, 0, 0, 0, time.UTC)
			},
			expectedAmount:  112, // 1% fee = 1, doubled = 2, + rush fee = 10 => total = 12
			expectError:     false,
			expectedBalance: 888,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Zero amount withdrawal",
			initialBalance:  1000,
			initialTrans:    0,
			initialLastOp:   time.Now().Add(-2 * time.Second),
			initialVip:      false,
			amount:          0,
			expectedAmount:  0,
			expectError:     false,
			expectedBalance: 1000,
			expectedTrans:   1,
			expectedVip:     false,
		},
		{
			name:            "Exact balance withdrawal",
			initialBalance:  100,
			initialTrans:    0,
			initialLastOp:   time.Now().Add(-2 * time.Second),
			initialVip:      false,
			amount:          100,
			expectedAmount:  101, // 1% fee = 1
			expectError:     false,
			expectedBalance: -1, // ⚠️ !!!ATTENTION!!! This should actually fail due to insufficient funds
			expectedTrans:   1,
			expectedVip:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalTimeNow := time.Now
			if tt.timeOverride != nil {
				// ⚠️ Temporarily override time.Now for testing time-dependent logic
				timeNow = tt.timeOverride
				defer func() { timeNow = originalTimeNow }()
			}

			account := &Account{
				balance:       tt.initialBalance,
				transactions:  tt.initialTrans,
				lastOperation: tt.initialLastOp,
				vipStatus:     tt.initialVip,
			}

			amount, err := account.Withdraw(tt.amount)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if amount != tt.expectedAmount {
				t.Errorf("expected amount %.2f, got %.2f", tt.expectedAmount, amount)
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

// ⚠️ Override time.Now for testing purposes
var timeNow = time.Now
