package webhook_test

import (
	"encoding/json"
	"errors"
	"testing"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/webhook"
)

const apiKey = "webhook-secret"

func makeSignedBody(t *testing.T, fields map[string]any) []byte {
	t.Helper()
	body, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	signed := make(map[string]any, len(fields)+1)
	for k, v := range fields {
		signed[k] = v
	}
	signed["sign"] = heleket.Sign(body, apiKey)
	out, err := json.Marshal(signed)
	if err != nil {
		t.Fatalf("marshal signed: %v", err)
	}
	return out
}

func TestVerifyRaw_ValidPayloadDecodesTypedFields(t *testing.T) {
	body := makeSignedBody(t, map[string]any{
		"type":     "payment",
		"uuid":     "pmt-1",
		"order_id": "order-1",
		"status":   "paid",
		"is_final": true,
		"amount":   "15.00",
		"txid":     "0xabc",
		"network":  "bsc",
	})

	v := webhook.NewVerifier(apiKey)
	result, err := v.VerifyRaw(body)
	if err != nil {
		t.Fatalf("VerifyRaw: %v", err)
	}
	if !result.IsPayment() || result.IsPayout() {
		t.Errorf("type detection failed: %+v", result)
	}
	if result.UUID != "pmt-1" || result.OrderID != "order-1" || result.Status != "paid" {
		t.Errorf("typed fields wrong: %+v", result)
	}
	if !result.IsFinalStatus() {
		t.Errorf("IsFinalStatus = false; want true")
	}
	if !result.IsSuccessful() {
		t.Errorf("IsSuccessful = false; want true")
	}
}

func TestVerifyRaw_TamperedPayloadRejected(t *testing.T) {
	// Sign with status=paid, then tamper the body to status=fail.
	original := makeSignedBody(t, map[string]any{
		"type":   "payment",
		"uuid":   "pmt-1",
		"status": "paid",
	})
	tampered := []byte(string(original))
	// Replace "paid" with "fail" — same length, keeps JSON valid.
	for i := 0; i < len(tampered)-3; i++ {
		if string(tampered[i:i+4]) == "paid" {
			copy(tampered[i:i+4], []byte("fail"))
			break
		}
	}

	_, err := webhook.NewVerifier(apiKey).VerifyRaw(tampered)
	var se *heleket.SignatureError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SignatureError; got %T (%v)", err, err)
	}
	if !errors.Is(err, heleket.ErrSignature) {
		t.Error("errors.Is(err, ErrSignature) = false; want true")
	}
}

func TestVerifyRaw_MissingSignRejected(t *testing.T) {
	body, _ := json.Marshal(map[string]any{"type": "payment"})
	_, err := webhook.NewVerifier(apiKey).VerifyRaw(body)
	var se *heleket.SignatureError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SignatureError; got %T", err)
	}
}

func TestVerifyRaw_WrongKeyRejected(t *testing.T) {
	body := makeSignedBody(t, map[string]any{"type": "payment", "status": "paid"})
	_, err := webhook.NewVerifier("a-different-key").VerifyRaw(body)
	var se *heleket.SignatureError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SignatureError")
	}
}

func TestVerifyRaw_DecodesAndVerifies(t *testing.T) {
	body := makeSignedBody(t, map[string]any{
		"type":     "payout",
		"uuid":     "po-7",
		"status":   "paid",
		"is_final": true,
	})

	result, err := webhook.NewVerifier(apiKey).VerifyRaw(body)
	if err != nil {
		t.Fatalf("VerifyRaw: %v", err)
	}
	if !result.IsPayout() {
		t.Errorf("expected payout type")
	}
	if !result.IsSuccessful() {
		t.Errorf("expected successful")
	}
}

// TestVerifyRaw_PHPStylePayload covers the real-world case: a webhook signed
// by Heleket's PHP server, with insertion-order keys and PHP-style \/ slash
// escapes. A naive map round-trip would fail this because Go's json.Marshal
// sorts keys and doesn't escape slashes.
func TestVerifyRaw_PHPStylePayload(t *testing.T) {
	bodyWithoutSign := `{"type":"payment","uuid":"u1","status":"paid","url":"https:\/\/pay.heleket.com\/pay\/abc"}`

	sign := heleket.Sign([]byte(bodyWithoutSign), apiKey)
	body := `{"type":"payment","uuid":"u1","status":"paid","url":"https:\/\/pay.heleket.com\/pay\/abc","sign":"` + sign + `"}`

	result, err := webhook.NewVerifier(apiKey).VerifyRaw([]byte(body))
	if err != nil {
		t.Fatalf("VerifyRaw on PHP-style payload: %v", err)
	}
	if result.UUID != "u1" || result.Status != "paid" {
		t.Errorf("decoded fields wrong: %+v", result)
	}
}

// TestVerifyRaw_SignFirst covers the edge case where sign is the FIRST field.
// The byte-stripper must handle leading vs trailing comma correctly.
func TestVerifyRaw_SignFirst(t *testing.T) {
	bodyWithoutSign := `{"type":"payment","uuid":"u1","status":"paid"}`
	sign := heleket.Sign([]byte(bodyWithoutSign), apiKey)
	body := `{"sign":"` + sign + `","type":"payment","uuid":"u1","status":"paid"}`

	if _, err := webhook.NewVerifier(apiKey).VerifyRaw([]byte(body)); err != nil {
		t.Fatalf("VerifyRaw with sign-first: %v", err)
	}
}

// TestVerifyRaw_NonObjectRejected guards against a "JSON array body" attack.
func TestVerifyRaw_NonObjectRejected(t *testing.T) {
	if _, err := webhook.NewVerifier(apiKey).VerifyRaw([]byte(`[1,2,3]`)); err == nil {
		t.Fatal("expected error for non-object body")
	}
}

func TestVerifyRaw_RejectsNonJSON(t *testing.T) {
	_, err := webhook.NewVerifier(apiKey).VerifyRaw([]byte("not json"))
	var se *heleket.SignatureError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SignatureError; got %T (%v)", err, err)
	}
}

// TestVerifyRaw_RespectsExplicitIsFinalFalse covers the IsFinal *bool
// behaviour: when the server explicitly sends is_final=false, the SDK must
// NOT override it with the local status enum.
func TestVerifyRaw_RespectsExplicitIsFinalFalse(t *testing.T) {
	// Status "paid" would locally classify as final, but the server says false.
	body := makeSignedBody(t, map[string]any{
		"type":     "payment",
		"uuid":     "u1",
		"status":   "paid",
		"is_final": false,
	})
	result, err := webhook.NewVerifier(apiKey).VerifyRaw(body)
	if err != nil {
		t.Fatalf("VerifyRaw: %v", err)
	}
	if result.IsFinal == nil {
		t.Fatal("IsFinal nil; expected explicit false")
	}
	if *result.IsFinal {
		t.Errorf("IsFinal = true; want false")
	}
	if result.IsFinalStatus() {
		t.Errorf("IsFinalStatus() = true; server said false")
	}
}

// TestVerifyRaw_FallsBackToStatusEnumWhenAbsent: server omits is_final, SDK
// must use the local enum.
func TestVerifyRaw_FallsBackToStatusEnumWhenAbsent(t *testing.T) {
	body := makeSignedBody(t, map[string]any{
		"type":   "payment",
		"uuid":   "u1",
		"status": "paid",
	})
	result, err := webhook.NewVerifier(apiKey).VerifyRaw(body)
	if err != nil {
		t.Fatalf("VerifyRaw: %v", err)
	}
	if result.IsFinal != nil {
		t.Errorf("IsFinal should be nil when absent; got %v", *result.IsFinal)
	}
	if !result.IsFinalStatus() {
		t.Errorf("IsFinalStatus() should fall back to enum and return true for paid")
	}
}
