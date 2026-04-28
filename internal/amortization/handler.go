package amortization

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	amortizationv1 "github.com/bradley-mader/amortization-calculator-api/gen/amortization/v1"
	"github.com/bradley-mader/amortization-calculator-api/gen/amortization/v1/amortizationv1connect"
)

// Compile-time interface assertion.
var _ amortizationv1connect.AmortizationServiceHandler = (*AmortizationServiceHandler)(nil)

// AmortizationServiceHandler implements the ConnectRPC AmortizationService.
type AmortizationServiceHandler struct{}

// NewAmortizationServiceHandler creates a new AmortizationServiceHandler.
func NewAmortizationServiceHandler() *AmortizationServiceHandler {
	return &AmortizationServiceHandler{}
}

// GetAmortizationSchedule maps the proto request to internal types, computes the
// amortization schedule, and maps the result back to the proto response.
func (h *AmortizationServiceHandler) GetAmortizationSchedule(
	ctx context.Context,
	req *connect.Request[amortizationv1.GetAmortizationScheduleRequest],
) (*connect.Response[amortizationv1.GetAmortizationScheduleResponse], error) {
	msg := req.Msg

	// Map proto request to internal CalculationInput.
	input := CalculationInput{
		Principal:         msg.GetPrincipal(),
		Interest:          msg.GetInterest(),
		Term:              msg.GetTerm(),
		PaymentDayOfMonth: msg.GetPaymentDayOfMonth(),
		SkipFirstMonth:    msg.GetSkipFirstMonth(),
	}

	if msg.GetStartDate() != nil {
		input.StartDate = msg.GetStartDate().AsTime()
	}

	for _, p := range msg.GetAdditionalPayments() {
		ap := AdditionalPayment{
			Amount: p.GetAmount(),
		}
		if p.GetDate() != nil {
			ap.Date = p.GetDate().AsTime()
		}
		input.AdditionalPayments = append(input.AdditionalPayments, ap)
	}

	// Calculate the amortization schedule.
	result, err := Calculate(input)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Map internal result back to proto response.
	payments := make([]*amortizationv1.SchedulePayment, len(result.Schedule))
	for i, entry := range result.Schedule {
		payments[i] = &amortizationv1.SchedulePayment{
			Date:               timestamppb.New(entry.Date),
			TotalPrincipalPaid: entry.TotalPrincipalPaid,
			TotalInterestPaid:  entry.TotalInterestPaid,
			PaymentAmount:      entry.PaymentAmount,
			PaymentPrincipal:   entry.PaymentPrincipal,
			PaymentInterest:    entry.PaymentInterest,
		}
	}

	resp := &amortizationv1.GetAmortizationScheduleResponse{
		MonthlyPayment: result.MonthlyPayment,
		Schedule: &amortizationv1.AmortizationSchedule{
			Payments: payments,
		},
	}

	return connect.NewResponse(resp), nil
}
