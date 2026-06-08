package heleket_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	heleket "github.com/heleket/go-sdk"
)

func TestHTTPTransport_ForwardsBytesVerbatim(t *testing.T) {
	receivedBody := ""
	receivedMerchant := ""
	receivedSign := ""

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q", r.Method)
		}
		raw, _ := io.ReadAll(r.Body)
		receivedBody = string(raw)
		receivedMerchant = r.Header.Get("merchant")
		receivedSign = r.Header.Get("sign")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"state":0,"result":{"uuid":"u1","order_id":"order-1"}}`))
	}))
	defer srv.Close()

	client, err := heleket.NewPaymentClient(testMerchantID, testAPIKey,
		heleket.WithBaseURL(srv.URL),
	)
	if err != nil {
		t.Fatalf("NewPaymentClient: %v", err)
	}

	_, err = client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount: "1", Currency: "USD", OrderID: "order-1",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}

	if !strings.Contains(receivedBody, `"order_id":"order-1"`) {
		t.Errorf("body lost in transport: %q", receivedBody)
	}
	if receivedMerchant != testMerchantID {
		t.Errorf("merchant header missing: %q", receivedMerchant)
	}
	if receivedSign == "" {
		t.Errorf("sign header missing")
	}
}

func TestHTTPTransport_RejectsOversizedResponse(t *testing.T) {
	// Server returns 5 KiB of slop.
	bigBody := bytes.Repeat([]byte("x"), 5*1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(bigBody)
	}))
	defer srv.Close()

	client, err := heleket.NewPaymentClient(testMerchantID, testAPIKey,
		heleket.WithBaseURL(srv.URL),
		heleket.WithMaxResponseBytes(1024), // 1 KiB cap; response is 5 KiB
		heleket.WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("NewPaymentClient: %v", err)
	}

	_, err = client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount: "1", Currency: "USD", OrderID: "x",
	})
	if err == nil {
		t.Fatal("expected error for oversized response")
	}
	if !errors.Is(err, heleket.ErrTransport) {
		t.Errorf("expected ErrTransport; got %v", err)
	}
}

func TestHTTPTransport_BlocksCrossHostRedirects(t *testing.T) {
	// Backend server returns 302 to an external host with the same path.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://attacker.example.com/v1/payment")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	client, err := heleket.NewPaymentClient(testMerchantID, testAPIKey,
		heleket.WithBaseURL(srv.URL),
		heleket.WithMaxRetries(0),
	)
	if err != nil {
		t.Fatalf("NewPaymentClient: %v", err)
	}
	// A redirect that is followed would land on attacker.example.com. The
	// default redirect policy is ErrUseLastResponse, so the SDK should see
	// the 302 response body itself — which is empty / non-JSON → APIError.
	_, err = client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount: "1", Currency: "USD", OrderID: "x",
	})
	if err == nil {
		t.Fatal("expected error when server returns redirect")
	}
	if !errors.Is(err, heleket.ErrAPI) {
		t.Errorf("expected ErrAPI (non-JSON response); got %v", err)
	}
}
