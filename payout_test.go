package heleket_test

import (
	"context"
	"encoding/json"
	"testing"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/internal/testutil"
)

func newPayoutWithFake(t *testing.T, fake *testutil.FakeTransport) *heleket.PayoutClient {
	t.Helper()
	c, err := heleket.NewPayoutClient(testMerchantID, "payout-key-xyz",
		heleket.WithTransport(fake),
		heleket.WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("NewPayoutClient: %v", err)
	}
	return c
}

func TestCreatePayout_SendsExpectedRequest(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state": 0,
		"result": map[string]any{
			"uuid":     "po-1",
			"status":   "process",
			"is_final": false,
			"amount":   "5",
			"currency": "USDT",
		},
	}, 200)
	client := newPayoutWithFake(t, fake)

	out, err := client.CreatePayout(context.Background(), heleket.CreatePayoutRequest{
		Amount:     "5",
		Currency:   "USDT",
		Network:    "TRON",
		OrderID:    "po-1",
		Address:    "TDD97yguPESTpcrJMqU6h2ozZbibv4Vaqm",
		IsSubtract: true,
	})
	if err != nil {
		t.Fatalf("CreatePayout: %v", err)
	}
	if out.UUID != "po-1" || out.Status != heleket.PayoutStatusProcess {
		t.Errorf("unexpected payout: %+v", out)
	}
	if want := "https://api.heleket.com/v1/payout"; fake.LastRequest().URL != want {
		t.Errorf("url = %q", fake.LastRequest().URL)
	}
}

func TestCalculateWithdrawal_SerializesBoolean(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state":  0,
		"result": map[string]any{"commission": "0.5"},
	}, 200)
	client := newPayoutWithFake(t, fake)

	_, err := client.CalculateWithdrawal(context.Background(), heleket.CalculateRequest{
		Currency:   "USDT",
		Network:    "TRON",
		Amount:     "10",
		IsSubtract: true,
	})
	if err != nil {
		t.Fatalf("CalculateWithdrawal: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(fake.LastRequest().Body, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["amount"] != "10" || body["is_subtract"] != true {
		t.Errorf("body = %#v", body)
	}
}

func TestTransferToPersonalAndBusiness(t *testing.T) {
	fake := testutil.NewFakeTransport().
		EnqueueJSON(map[string]any{"state": 0, "result": map[string]any{"amount": "1.5"}}, 200).
		EnqueueJSON(map[string]any{"state": 0, "result": map[string]any{"amount": "1.5"}}, 200)
	client := newPayoutWithFake(t, fake)

	if _, err := client.TransferToPersonal(context.Background(), heleket.TransferRequest{Amount: "1.5", Currency: "USDT"}); err != nil {
		t.Fatalf("TransferToPersonal: %v", err)
	}
	if _, err := client.TransferToBusiness(context.Background(), heleket.TransferRequest{Amount: "1.5", Currency: "USDT"}); err != nil {
		t.Fatalf("TransferToBusiness: %v", err)
	}

	reqs := fake.Requests()
	if reqs[0].URL != "https://api.heleket.com/v1/transfer/to-personal" {
		t.Errorf("first url = %q", reqs[0].URL)
	}
	if reqs[1].URL != "https://api.heleket.com/v1/transfer/to-business" {
		t.Errorf("second url = %q", reqs[1].URL)
	}
}

func TestPayoutRefund_HitsPaymentRefundSignedWithPayoutKey(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state":  0,
		"result": map[string]any{"uuid": "inv-1", "status": "refund_process"},
	}, 200)
	client := newPayoutWithFake(t, fake)

	if _, err := client.Refund(context.Background(), heleket.RefundRequest{
		UUID:       "inv-1",
		Address:    "TBaCkAdDrEsS",
		IsSubtract: true,
	}); err != nil {
		t.Fatalf("Refund: %v", err)
	}

	req := fake.LastRequest()
	if req.Method != "POST" {
		t.Errorf("method = %q, want POST", req.Method)
	}
	// The endpoint is /v1/payment/refund even though it lives on PayoutClient.
	if want := "https://api.heleket.com/v1/payment/refund"; req.URL != want {
		t.Errorf("url = %q, want %q", req.URL, want)
	}

	var body map[string]any
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["uuid"] != "inv-1" || body["address"] != "TBaCkAdDrEsS" || body["is_subtract"] != true {
		t.Errorf("body = %#v", body)
	}

	// The whole point of the move: the refund is signed with the PAYOUT key the
	// client was built with, not the payment key.
	if want := heleket.Sign(req.Body, "payout-key-xyz"); req.Headers.Get("sign") != want {
		t.Errorf("sign = %q, want %q (signed with payout key)", req.Headers.Get("sign"), want)
	}
}
