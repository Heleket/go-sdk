package heleket_test

import (
	"testing"

	heleket "github.com/heleket/go-sdk"
)

func TestPaymentStatusClassification(t *testing.T) {
	cases := []struct {
		s          heleket.PaymentStatus
		isFinal    bool
		successful bool
	}{
		{heleket.PaymentStatusPaid, true, true},
		{heleket.PaymentStatusPaidOver, true, true},
		{heleket.PaymentStatusWrongAmount, true, false},
		{heleket.PaymentStatusCheck, false, false},
		{heleket.PaymentStatusConfirmCheck, false, false},
		{heleket.PaymentStatusWrongAmountWaiting, false, false},
		{heleket.PaymentStatusCancel, true, false},
		{heleket.PaymentStatusFail, true, false},
		{heleket.PaymentStatusLocked, true, false},
		{heleket.PaymentStatusRefundProcess, false, false},
		{heleket.PaymentStatusRefundPaid, true, false},
	}
	for _, tc := range cases {
		if got := tc.s.IsFinal(); got != tc.isFinal {
			t.Errorf("%s.IsFinal() = %v; want %v", tc.s, got, tc.isFinal)
		}
		if got := tc.s.IsSuccessful(); got != tc.successful {
			t.Errorf("%s.IsSuccessful() = %v; want %v", tc.s, got, tc.successful)
		}
	}
}

func TestAmlLinkStatusClassification(t *testing.T) {
	cases := []struct {
		s          heleket.AmlLinkStatus
		isFinal    bool
		successful bool
	}{
		{heleket.AmlLinkStatusCompleted, true, true},
		{heleket.AmlLinkStatusExpired, true, false},
		{heleket.AmlLinkStatusInit, false, false},
		{heleket.AmlLinkStatusPending, false, false},
	}
	for _, tc := range cases {
		if got := tc.s.IsFinal(); got != tc.isFinal {
			t.Errorf("%s.IsFinal() = %v; want %v", tc.s, got, tc.isFinal)
		}
		if got := tc.s.IsSuccessful(); got != tc.successful {
			t.Errorf("%s.IsSuccessful() = %v; want %v", tc.s, got, tc.successful)
		}
	}
}

func TestPayoutStatusClassification(t *testing.T) {
	cases := []struct {
		s          heleket.PayoutStatus
		isFinal    bool
		successful bool
	}{
		{heleket.PayoutStatusPaid, true, true},
		{heleket.PayoutStatusFail, true, false},
		{heleket.PayoutStatusCancel, true, false},
		{heleket.PayoutStatusSystemFail, true, false},
		{heleket.PayoutStatusProcess, false, false},
		{heleket.PayoutStatusCheck, false, false},
	}
	for _, tc := range cases {
		if got := tc.s.IsFinal(); got != tc.isFinal {
			t.Errorf("%s.IsFinal() = %v; want %v", tc.s, got, tc.isFinal)
		}
		if got := tc.s.IsSuccessful(); got != tc.successful {
			t.Errorf("%s.IsSuccessful() = %v; want %v", tc.s, got, tc.successful)
		}
	}
}
