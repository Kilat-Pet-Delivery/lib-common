package domain

import (
	"fmt"
	"math"
)

// Currency constants.
const (
	CurrencyMYR = "MYR"
	CurrencyUSD = "USD"
	CurrencySGD = "SGD"
)

// Money is an immutable value object representing a monetary amount in the smallest unit (cents).
type Money struct {
	amount   int64
	currency string
}

// NewMoney creates a new Money value object from cents.
func NewMoney(amountCents int64, currency string) Money {
	return Money{amount: amountCents, currency: currency}
}

// MYR creates a Money value object in Malaysian Ringgit from cents (sen).
func MYR(amountSen int64) Money {
	return Money{amount: amountSen, currency: CurrencyMYR}
}

// MYRFromFloat creates a Money value object in MYR from a float (e.g., 10.50).
func MYRFromFloat(amount float64) Money {
	return Money{amount: int64(math.Round(amount * 100)), currency: CurrencyMYR}
}

// Zero returns a zero amount in the given currency.
func Zero(currency string) Money {
	return Money{amount: 0, currency: currency}
}

// FromCents creates a Money from a cent amount and currency.
func FromCents(cents int64, currency string) Money {
	return Money{amount: cents, currency: currency}
}

// Amount returns the amount in cents.
func (m Money) Amount() int64 { return m.amount }

// Currency returns the currency code.
func (m Money) Currency() string { return m.currency }

// Float64 returns the amount as a float with 2 decimal places.
func (m Money) Float64() float64 {
	return float64(m.amount) / 100.0
}

// Add adds two Money values of the same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add %s and %s", m.currency, other.currency)
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Subtract subtracts another Money value.
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot subtract %s from %s", other.currency, m.currency)
	}
	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

// Multiply multiplies the amount by a factor.
func (m Money) Multiply(factor float64) Money {
	return Money{amount: int64(math.Round(float64(m.amount) * factor)), currency: m.currency}
}

// Percentage calculates a percentage of the amount.
func (m Money) Percentage(percent float64) Money {
	return Money{amount: int64(math.Round(float64(m.amount) * percent / 100.0)), currency: m.currency}
}

// IsNegative returns true if the amount is negative.
func (m Money) IsNegative() bool { return m.amount < 0 }

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool { return m.amount == 0 }

// IsPositive returns true if the amount is positive.
func (m Money) IsPositive() bool { return m.amount > 0 }

// Equals checks if two Money values are equal.
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// GreaterThan checks if this Money is greater than another.
func (m Money) GreaterThan(other Money) bool {
	return m.currency == other.currency && m.amount > other.amount
}

// String returns a human-readable money string.
func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.currency, m.Float64())
}
