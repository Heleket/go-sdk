package heleket

// PaymentStatus is the status of an invoice. See
// https://doc.heleket.com/methods/payments/payment-statuses for details.
type PaymentStatus string

const (
	// PaymentStatusPaid: exact amount received. Final + successful.
	PaymentStatusPaid PaymentStatus = "paid"
	// PaymentStatusPaidOver: overpayment received. Final + successful.
	PaymentStatusPaidOver PaymentStatus = "paid_over"
	// PaymentStatusWrongAmount: underpaid, no further attempts allowed. Final.
	PaymentStatusWrongAmount PaymentStatus = "wrong_amount"
	// PaymentStatusProcess: payment is being processed. Intermediate.
	PaymentStatusProcess PaymentStatus = "process"
	// PaymentStatusConfirmCheck: seen on-chain; awaiting confirmations. Intermediate.
	PaymentStatusConfirmCheck PaymentStatus = "confirm_check"
	// PaymentStatusWrongAmountWaiting: underpaid, additional top-ups accepted. Intermediate.
	PaymentStatusWrongAmountWaiting PaymentStatus = "wrong_amount_waiting"
	// PaymentStatusCheck: waiting for the transaction to appear on-chain. Intermediate.
	PaymentStatusCheck PaymentStatus = "check"
	// PaymentStatusFail: payment error. Final.
	PaymentStatusFail PaymentStatus = "fail"
	// PaymentStatusCancel: invoice abandoned by the client. Final.
	PaymentStatusCancel PaymentStatus = "cancel"
	// PaymentStatusSystemFail: system-side error. Final.
	PaymentStatusSystemFail PaymentStatus = "system_fail"
	// PaymentStatusRefundProcess: refund in flight. Intermediate.
	PaymentStatusRefundProcess PaymentStatus = "refund_process"
	// PaymentStatusRefundFail: refund failed. Final.
	PaymentStatusRefundFail PaymentStatus = "refund_fail"
	// PaymentStatusRefundPaid: refund completed. Final.
	PaymentStatusRefundPaid PaymentStatus = "refund_paid"
	// PaymentStatusLocked: AML hold. Final.
	PaymentStatusLocked PaymentStatus = "locked"
)

// IsFinal reports whether the payment status is terminal — the invoice will
// not transition further.
func (s PaymentStatus) IsFinal() bool {
	switch s {
	case PaymentStatusPaid,
		PaymentStatusPaidOver,
		PaymentStatusWrongAmount,
		PaymentStatusFail,
		PaymentStatusCancel,
		PaymentStatusSystemFail,
		PaymentStatusRefundFail,
		PaymentStatusRefundPaid,
		PaymentStatusLocked:
		return true
	}
	return false
}

// IsSuccessful reports whether the payment represents a successful payment
// (paid exactly or overpaid).
func (s PaymentStatus) IsSuccessful() bool {
	return s == PaymentStatusPaid || s == PaymentStatusPaidOver
}

// PayoutStatus is the status of a withdrawal. See
// https://doc.heleket.com/methods/payouts/payout-statuses for details.
type PayoutStatus string

const (
	PayoutStatusProcess    PayoutStatus = "process"     // intermediate
	PayoutStatusCheck      PayoutStatus = "check"       // intermediate
	PayoutStatusPaid       PayoutStatus = "paid"        // final, successful
	PayoutStatusFail       PayoutStatus = "fail"        // final
	PayoutStatusCancel     PayoutStatus = "cancel"      // final
	PayoutStatusSystemFail PayoutStatus = "system_fail" // final
)

// IsFinal reports whether the payout status is terminal.
func (s PayoutStatus) IsFinal() bool {
	switch s {
	case PayoutStatusPaid, PayoutStatusFail, PayoutStatusCancel, PayoutStatusSystemFail:
		return true
	}
	return false
}

// IsSuccessful reports whether the payout settled successfully.
func (s PayoutStatus) IsSuccessful() bool {
	return s == PayoutStatusPaid
}

// CourseSource identifies the exchange-rate source used when converting fiat
// invoice amounts. See https://doc.heleket.com/methods/payments/creating-invoice.
type CourseSource string

const (
	CourseSourceBinance    CourseSource = "Binance"
	CourseSourceBinanceP2P CourseSource = "BinanceP2P"
	CourseSourceExmo       CourseSource = "Exmo"
	CourseSourceKucoin     CourseSource = "Kucoin"
)
