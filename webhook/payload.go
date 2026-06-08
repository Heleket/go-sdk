// Package webhook handles incoming Heleket webhooks.
//
// The webhook lives in its own subpackage so merchants who only need to
// receive callbacks can import "github.com/heleket/go-sdk/webhook" without
// pulling in the full API client.
package webhook

import (
	heleket "github.com/heleket/go-sdk"
)

// Payload is a verified webhook event. Instances are constructed exclusively
// by Verifier and should not be created directly.
//
// IsFinal is a *bool because Heleket's server may omit the field entirely;
// distinguishing "absent" from "false" lets the SDK respect an explicit
// is_final=false even when the local status enum would classify the status
// as terminal.
type Payload struct {
	Type           string         `json:"type"`
	UUID           string         `json:"uuid"`
	OrderID        string         `json:"order_id"`
	Amount         string         `json:"amount"`
	PaymentAmount  string         `json:"payment_amount,omitempty"`
	MerchantAmount string         `json:"merchant_amount,omitempty"`
	Commission     string         `json:"commission,omitempty"`
	IsFinal        *bool          `json:"is_final,omitempty"`
	Status         string         `json:"status"`
	From           string         `json:"from,omitempty"`
	Network        string         `json:"network,omitempty"`
	Currency       string         `json:"currency,omitempty"`
	PayerCurrency  string         `json:"payer_currency,omitempty"`
	TxID           string         `json:"txid,omitempty"`
	Sign           string         `json:"sign,omitempty"`
	Raw            map[string]any `json:"-"`
}

// IsPayment reports whether this is a payment-side webhook (invoice or static wallet).
func (p *Payload) IsPayment() bool {
	return p.Type == "payment" || p.Type == "wallet"
}

// IsPayout reports whether this is a payout webhook.
func (p *Payload) IsPayout() bool {
	return p.Type == "payout"
}

// IsSuccessful reports whether the event represents a successful payment or payout.
func (p *Payload) IsSuccessful() bool {
	if p.IsPayout() {
		return heleket.PayoutStatus(p.Status).IsSuccessful()
	}
	return heleket.PaymentStatus(p.Status).IsSuccessful()
}

// IsFinalStatus reports whether the status will not transition further. The
// server-provided is_final field takes precedence when present (even when
// explicitly false). When absent, the SDK falls back to the typed status enum.
func (p *Payload) IsFinalStatus() bool {
	if p.IsFinal != nil {
		return *p.IsFinal
	}
	if p.IsPayout() {
		return heleket.PayoutStatus(p.Status).IsFinal()
	}
	return heleket.PaymentStatus(p.Status).IsFinal()
}
