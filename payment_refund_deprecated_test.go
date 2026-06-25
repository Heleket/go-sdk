package heleket

// This test lives in package heleket (not heleket_test) for two reasons:
// calling the deprecated PaymentClient.Refund from the declaring package does
// not trip staticcheck's SA1019, and it lets us use a local transport spy
// without the import cycle that internal/testutil would introduce.

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

// refusingTransport fails the test if any request is dispatched. The deprecated
// PaymentClient.Refund must short-circuit before reaching the transport.
type refusingTransport struct{ called bool }

func (r *refusingTransport) RoundTrip(_ context.Context, method, url string, _ http.Header, _ []byte) (*Response, error) {
	r.called = true
	return nil, errors.New("transport must not be called: " + method + " " + url)
}

func TestPaymentClientRefund_DeprecatedStubFailsWithoutRequest(t *testing.T) {
	spy := &refusingTransport{}
	client, err := NewPaymentClient("MERCHANT", "payment-key", WithTransport(spy), WithMaxRetries(0))
	if err != nil {
		t.Fatalf("NewPaymentClient: %v", err)
	}

	out, err := client.Refund(context.Background(), RefundRequest{
		UUID:       "inv-1",
		Address:    "X",
		IsSubtract: true,
	})
	if out != nil {
		t.Errorf("result = %#v, want nil", out)
	}
	if !errors.Is(err, ErrRefundMoved) {
		t.Fatalf("err = %v, want ErrRefundMoved", err)
	}
	if spy.called {
		t.Error("PaymentClient.Refund issued an HTTP request; it must fail loudly before signing with the wrong key")
	}
}
