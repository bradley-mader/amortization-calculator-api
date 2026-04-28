package amortization

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// AdditionalPayment represents an extra payment applied to principal.
type AdditionalPayment struct {
	Date   time.Time
	Amount float64
}

// ScheduleEntry represents a single row in the amortization schedule.
type ScheduleEntry struct {
	Date               time.Time
	TotalPrincipalPaid float64
	TotalInterestPaid  float64
	PaymentAmount      float64
	PaymentPrincipal   float64
	PaymentInterest    float64
}

// CalculationInput contains the parameters for computing an amortization schedule.
type CalculationInput struct {
	Principal          float64
	Interest           float64
	Term               int32
	StartDate          time.Time
	PaymentDayOfMonth  int32
	SkipFirstMonth     bool
	AdditionalPayments []AdditionalPayment
}

// CalculationResult contains the computed monthly payment and full schedule.
type CalculationResult struct {
	MonthlyPayment float64
	Schedule       []ScheduleEntry
}

// roundToTwoDecimals rounds a float64 to two decimal places.
func roundToTwoDecimals(val float64) float64 {
	return math.Round(val*100) / 100
}

// firstPaymentDate computes the first payment date given the start date,
// desired day of month, and whether to skip the first month.
func firstPaymentDate(startDate time.Time, dayOfMonth int32, skipFirstMonth bool) time.Time {
	y, m, _ := startDate.Date()
	candidate := time.Date(y, m, int(dayOfMonth), 0, 0, 0, 0, time.UTC)

	// If the candidate is before startDate, advance one month.
	if candidate.Before(startDate) {
		candidate = time.Date(y, m+1, int(dayOfMonth), 0, 0, 0, 0, time.UTC)
	}

	// If skipFirstMonth and the candidate is less than one full month after startDate, advance again.
	if skipFirstMonth {
		oneMonthLater := time.Date(y, m+1, startDate.Day(), 0, 0, 0, 0, time.UTC)
		if candidate.Before(oneMonthLater) {
			cy, cm, _ := candidate.Date()
			candidate = time.Date(cy, cm+1, int(dayOfMonth), 0, 0, 0, 0, time.UTC)
		}
	}

	return candidate
}

// Validate checks that all fields in the input are within acceptable ranges.
func Validate(input CalculationInput) error {
	if input.Principal <= 0 {
		return fmt.Errorf("principal must be greater than 0, got %f", input.Principal)
	}
	if input.Interest < 0 || input.Interest > 100 {
		return fmt.Errorf("interest must be between 0 and 100, got %f", input.Interest)
	}
	if input.Term <= 0 || input.Term > 600 {
		return fmt.Errorf("term must be between 1 and 600, got %d", input.Term)
	}
	if input.PaymentDayOfMonth < 1 || input.PaymentDayOfMonth > 28 {
		return fmt.Errorf("paymentDayOfMonth must be between 1 and 28, got %d", input.PaymentDayOfMonth)
	}
	if input.StartDate.IsZero() {
		return fmt.Errorf("startDate must be non-zero")
	}
	for i, ap := range input.AdditionalPayments {
		if ap.Amount <= 0 {
			return fmt.Errorf("additional payment %d: amount must be positive, got %f", i, ap.Amount)
		}
		if ap.Date.IsZero() {
			return fmt.Errorf("additional payment %d: date must be non-zero", i)
		}
	}
	return nil
}

// Calculate computes the amortization schedule for the given input.
func Calculate(input CalculationInput) (*CalculationResult, error) {
	if err := Validate(input); err != nil {
		return nil, err
	}

	r := input.Interest / 100.0 / 12.0
	n := float64(input.Term)

	var M float64
	if r == 0 {
		M = input.Principal / n
	} else {
		M = input.Principal * r / (1 - math.Pow(1+r, -n))
	}
	M = roundToTwoDecimals(M)

	// Sort additional payments by date.
	extras := make([]AdditionalPayment, len(input.AdditionalPayments))
	copy(extras, input.AdditionalPayments)
	sort.Slice(extras, func(i, j int) bool {
		return extras[i].Date.Before(extras[j].Date)
	})

	paymentDate := firstPaymentDate(input.StartDate, input.PaymentDayOfMonth, input.SkipFirstMonth)

	remainingPrincipal := input.Principal
	var totalPrincipalPaid, totalInterestPaid float64
	schedule := make([]ScheduleEntry, 0, input.Term)

	for i := int32(0); i < input.Term; i++ {
		if remainingPrincipal <= 0.01 {
			break
		}

		interest := roundToTwoDecimals(remainingPrincipal * r)
		principalPortion := M - interest

		// On the final scheduled payment, or if principal portion exceeds
		// remaining balance, adjust to pay off the loan exactly.
		if i == input.Term-1 || principalPortion >= remainingPrincipal {
			principalPortion = remainingPrincipal
		}

		// Compute next payment date for the additional-payment window.
		py, pm, _ := paymentDate.Date()
		nextPaymentDate := time.Date(py, pm+1, int(input.PaymentDayOfMonth), 0, 0, 0, 0, time.UTC)

		// Sum additional payments in [paymentDate, nextPaymentDate).
		var extraPrincipal float64
		for _, ap := range extras {
			if !ap.Date.Before(paymentDate) && ap.Date.Before(nextPaymentDate) {
				extraPrincipal += ap.Amount
			}
		}

		totalPrincipalThisMonth := principalPortion + extraPrincipal

		// Cap so we don't overpay.
		if totalPrincipalThisMonth > remainingPrincipal {
			totalPrincipalThisMonth = remainingPrincipal
			// Adjust principal portion (extra may have pushed us over).
			if principalPortion > remainingPrincipal {
				principalPortion = remainingPrincipal
				extraPrincipal = 0
			} else {
				extraPrincipal = remainingPrincipal - principalPortion
			}
		}

		paymentAmount := interest + totalPrincipalThisMonth

		remainingPrincipal -= totalPrincipalThisMonth
		remainingPrincipal = roundToTwoDecimals(remainingPrincipal)

		totalPrincipalPaid += totalPrincipalThisMonth
		totalInterestPaid += interest

		schedule = append(schedule, ScheduleEntry{
			Date:               paymentDate,
			TotalPrincipalPaid: roundToTwoDecimals(totalPrincipalPaid),
			TotalInterestPaid:  roundToTwoDecimals(totalInterestPaid),
			PaymentAmount:      roundToTwoDecimals(paymentAmount),
			PaymentPrincipal:   roundToTwoDecimals(totalPrincipalThisMonth),
			PaymentInterest:    interest,
		})

		paymentDate = nextPaymentDate
	}

	return &CalculationResult{
		MonthlyPayment: M,
		Schedule:       schedule,
	}, nil
}
