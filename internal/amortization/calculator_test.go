package amortization_test

import (
	"math"
	"testing"
	"time"

	"github.com/bradley-mader/amortization-calculator-api/internal/amortization"
)

func TestStandard30YearMortgage(t *testing.T) {
	input := amortization.CalculationInput{
		Principal:         200000,
		Interest:          6.5,
		Term:              360,
		StartDate:         time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		PaymentDayOfMonth: 15,
	}

	result, err := amortization.Calculate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(result.MonthlyPayment-1264.14) > 0.01 {
		t.Errorf("expected monthly payment ~1264.14, got %f", result.MonthlyPayment)
	}

	if len(result.Schedule) != 360 {
		t.Errorf("expected 360 payments, got %d", len(result.Schedule))
	}

	firstDate := result.Schedule[0].Date
	expectedFirst := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	if !firstDate.Equal(expectedFirst) {
		t.Errorf("expected first payment date %v, got %v", expectedFirst, firstDate)
	}

	lastEntry := result.Schedule[len(result.Schedule)-1]
	if math.Abs(lastEntry.TotalPrincipalPaid-200000) > 1.0 {
		t.Errorf("expected total principal paid ~200000, got %f", lastEntry.TotalPrincipalPaid)
	}
}

func TestShortLoan(t *testing.T) {
	input := amortization.CalculationInput{
		Principal:         10000,
		Interest:          5.0,
		Term:              12,
		StartDate:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PaymentDayOfMonth: 1,
	}

	result, err := amortization.Calculate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Schedule) != 12 {
		t.Errorf("expected 12 payments, got %d", len(result.Schedule))
	}

	lastEntry := result.Schedule[len(result.Schedule)-1]
	if math.Abs(lastEntry.TotalPrincipalPaid-10000) > 0.02 {
		t.Errorf("expected total principal paid ~10000, got %f", lastEntry.TotalPrincipalPaid)
	}
}

func TestSkipFirstMonth(t *testing.T) {
	input := amortization.CalculationInput{
		Principal:         10000,
		Interest:          5.0,
		Term:              12,
		StartDate:         time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		PaymentDayOfMonth: 5,
		SkipFirstMonth:    true,
	}

	result, err := amortization.Calculate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	firstDate := result.Schedule[0].Date
	expectedFirst := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	if !firstDate.Equal(expectedFirst) {
		t.Errorf("expected first payment date %v, got %v", expectedFirst, firstDate)
	}
}

func TestAdditionalPaymentsEarlyPayoff(t *testing.T) {
	input := amortization.CalculationInput{
		Principal:         10000,
		Interest:          5.0,
		Term:              12,
		StartDate:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PaymentDayOfMonth: 1,
		AdditionalPayments: []amortization.AdditionalPayment{
			{
				Date:   time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
				Amount: 5000,
			},
		},
	}

	result, err := amortization.Calculate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Schedule) >= 12 {
		t.Errorf("expected fewer than 12 payments due to additional payment, got %d", len(result.Schedule))
	}
}

func TestValidationErrors(t *testing.T) {
	validInput := amortization.CalculationInput{
		Principal:         10000,
		Interest:          5.0,
		Term:              12,
		StartDate:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PaymentDayOfMonth: 1,
	}

	tests := []struct {
		name  string
		input amortization.CalculationInput
	}{
		{
			name: "zero principal",
			input: func() amortization.CalculationInput {
				i := validInput
				i.Principal = 0
				return i
			}(),
		},
		{
			name: "negative interest",
			input: func() amortization.CalculationInput {
				i := validInput
				i.Interest = -1
				return i
			}(),
		},
		{
			name: "interest over 100",
			input: func() amortization.CalculationInput {
				i := validInput
				i.Interest = 101
				return i
			}(),
		},
		{
			name: "zero term",
			input: func() amortization.CalculationInput {
				i := validInput
				i.Term = 0
				return i
			}(),
		},
		{
			name: "payment day < 1",
			input: func() amortization.CalculationInput {
				i := validInput
				i.PaymentDayOfMonth = 0
				return i
			}(),
		},
		{
			name: "payment day > 28",
			input: func() amortization.CalculationInput {
				i := validInput
				i.PaymentDayOfMonth = 29
				return i
			}(),
		},
		{
			name: "zero start date",
			input: func() amortization.CalculationInput {
				i := validInput
				i.StartDate = time.Time{}
				return i
			}(),
		},
		{
			name: "additional payment with zero amount",
			input: func() amortization.CalculationInput {
				i := validInput
				i.AdditionalPayments = []amortization.AdditionalPayment{
					{Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Amount: 0},
				}
				return i
			}(),
		},
		{
			name: "additional payment with zero date",
			input: func() amortization.CalculationInput {
				i := validInput
				i.AdditionalPayments = []amortization.AdditionalPayment{
					{Date: time.Time{}, Amount: 100},
				}
				return i
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := amortization.Validate(tc.input)
			if err == nil {
				t.Errorf("expected validation error for %s, got nil", tc.name)
			}
		})
	}
}

func TestZeroInterestRate(t *testing.T) {
	input := amortization.CalculationInput{
		Principal:         12000,
		Interest:          0,
		Term:              12,
		StartDate:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PaymentDayOfMonth: 1,
	}

	result, err := amortization.Calculate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.MonthlyPayment != 1000.0 {
		t.Errorf("expected monthly payment 1000.0, got %f", result.MonthlyPayment)
	}

	lastEntry := result.Schedule[len(result.Schedule)-1]
	if lastEntry.TotalInterestPaid != 0 {
		t.Errorf("expected zero total interest, got %f", lastEntry.TotalInterestPaid)
	}

	if len(result.Schedule) != 12 {
		t.Errorf("expected 12 payments, got %d", len(result.Schedule))
	}
}
