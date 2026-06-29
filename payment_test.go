package heleket_test

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/internal/testutil"
)

const (
	testMerchantID = "8b03432e-385b-4670-8d06-064591096795"
	testAPIKey     = "payment-key-abc"
)

func newPaymentWithFake(t *testing.T, fake *testutil.FakeTransport) *heleket.PaymentClient {
	t.Helper()
	c, err := heleket.NewPaymentClient(testMerchantID, testAPIKey,
		heleket.WithTransport(fake),
		heleket.WithMaxRetries(0), // tests enqueue one response per call; no retries by default
	)
	if err != nil {
		t.Fatalf("NewPaymentClient: %v", err)
	}
	return c
}

func TestCreateInvoice_SendsSignedJSONPost(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state": 0,
		"result": map[string]any{
			"uuid":     "uuid-1",
			"order_id": "order-1",
			"amount":   "15.00",
			"currency": "USD",
		},
	}, 200)
	client := newPaymentWithFake(t, fake)

	invoice, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount:   "15",
		Currency: "USD",
		OrderID:  "order-1",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	if invoice.UUID != "uuid-1" || invoice.OrderID != "order-1" {
		t.Errorf("unexpected invoice: %+v", invoice)
	}

	req := fake.LastRequest()
	if req.Method != "POST" {
		t.Errorf("method = %q; want POST", req.Method)
	}
	if req.URL != "https://api.heleket.com/v1/payment" {
		t.Errorf("url = %q", req.URL)
	}
	if got := req.Headers.Get("merchant"); got != testMerchantID {
		t.Errorf("merchant header = %q; want %q", got, testMerchantID)
	}
	if got := req.Headers.Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type = %q", got)
	}

	expectedBody, _ := json.Marshal(heleket.CreateInvoiceRequest{
		Amount:   "15",
		Currency: "USD",
		OrderID:  "order-1",
	})
	if string(req.Body) != string(expectedBody) {
		t.Errorf("body = %q; want %q", req.Body, expectedBody)
	}

	expectedSign := md5.Sum([]byte(base64.StdEncoding.EncodeToString(expectedBody) + testAPIKey))
	wantSign := hex.EncodeToString(expectedSign[:])
	if got := req.Headers.Get("sign"); got != wantSign {
		t.Errorf("sign = %q; want %q", got, wantSign)
	}
}

func TestGetBalance_EmptyParamsSignsEmptyString(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state": 0,
		"result": []map[string]any{
			{"balance": map[string]any{"merchant": []any{}, "user": []any{}}},
		},
	}, 200)
	client := newPaymentWithFake(t, fake)

	if _, err := client.GetBalance(context.Background()); err != nil {
		t.Fatalf("GetBalance: %v", err)
	}

	req := fake.LastRequest()
	if len(req.Body) != 0 {
		t.Errorf("expected empty body; got %q", req.Body)
	}
	// md5(base64("") + key) == md5(key)
	want := md5.Sum([]byte(testAPIKey))
	if got := req.Headers.Get("sign"); got != hex.EncodeToString(want[:]) {
		t.Errorf("sign for empty body = %q", got)
	}
}

func TestCreateInvoice_ValidationError(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state": 1,
		"errors": map[string]any{
			"amount": []string{"validation.required"},
		},
	}, 422)
	client := newPaymentWithFake(t, fake)

	_, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{})
	if err == nil {
		t.Fatal("expected ValidationError")
	}
	var ve *heleket.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationError; got %T (%v)", err, err)
	}
	if ve.HTTPStatus != 422 {
		t.Errorf("HTTPStatus = %d; want 422", ve.HTTPStatus)
	}
	if got := ve.Fields["amount"]; len(got) != 1 || got[0] != "validation.required" {
		t.Errorf("Fields[amount] = %v", got)
	}
}

func TestCreateInvoice_APIError(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state":   1,
		"message": "The network was not found",
	}, 400)
	client := newPaymentWithFake(t, fake)

	_, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount: "1", Currency: "USD", OrderID: "x",
	})
	var ae *heleket.APIError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *APIError; got %T (%v)", err, err)
	}
	if ae.HTTPStatus != 400 {
		t.Errorf("HTTPStatus = %d", ae.HTTPStatus)
	}
	if !strings.Contains(ae.Message, "network was not found") {
		t.Errorf("Message = %q", ae.Message)
	}
}

func TestGetInfo_RequiresUUIDOrOrderID(t *testing.T) {
	client := newPaymentWithFake(t, testutil.NewFakeTransport())
	if _, err := client.GetInfo(context.Background(), heleket.InfoOptions{}); err == nil {
		t.Fatal("expected error for empty InfoOptions")
	}
}

func TestListHistory_AppendsCursorQueryString(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state":  0,
		"result": map[string]any{"items": []any{}, "paginate": map[string]any{}},
	}, 200)
	client := newPaymentWithFake(t, fake)

	_, err := client.ListHistory(context.Background(), heleket.HistoryOptions{Cursor: "abc=def"})
	if err != nil {
		t.Fatalf("ListHistory: %v", err)
	}
	want := "https://api.heleket.com/v1/payment/list?cursor=abc%3Ddef"
	if got := fake.LastRequest().URL; got != want {
		t.Errorf("url = %q; want %q", got, want)
	}
}

func TestTransportFailure_BecomesHTTPError(t *testing.T) {
	fake := testutil.NewFakeTransport().FailNext(errors.New("simulated network drop"))
	client := newPaymentWithFake(t, fake)

	_, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount: "1", Currency: "USD", OrderID: "x",
	})
	var he *heleket.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("expected *HTTPError; got %T (%v)", err, err)
	}
	if !errors.Is(err, heleket.ErrTransport) {
		t.Errorf("errors.Is(err, ErrTransport) = false; want true")
	}
}

func TestUserAgent_DefaultAndCustom(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state": 0, "result": map[string]any{"uuid": "u"},
	}, 200)
	client, err := heleket.NewPaymentClient(testMerchantID, testAPIKey,
		heleket.WithTransport(fake),
		heleket.WithMaxRetries(0),
		heleket.WithUserAgent("myapp/1.2"),
	)
	if err != nil {
		t.Fatalf("NewPaymentClient: %v", err)
	}
	if _, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount: "1", Currency: "USD", OrderID: "x",
	}); err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	got := fake.LastRequest().Headers.Get("User-Agent")
	if !strings.Contains(got, "heleket-go-sdk/") {
		t.Errorf("User-Agent missing SDK prefix: %q", got)
	}
	if !strings.Contains(got, "myapp/1.2") {
		t.Errorf("User-Agent missing custom token: %q", got)
	}
}

func TestRetry_On5xxThenSuccess(t *testing.T) {
	fake := testutil.NewFakeTransport().
		EnqueueJSON(map[string]any{"state": 1, "message": "server fail"}, 500).
		EnqueueJSON(map[string]any{"state": 0, "result": map[string]any{"uuid": "ok"}}, 200)
	client, err := heleket.NewPaymentClient(testMerchantID, testAPIKey,
		heleket.WithTransport(fake),
		heleket.WithMaxRetries(2),
	)
	if err != nil {
		t.Fatalf("NewPaymentClient: %v", err)
	}
	invoice, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount: "1", Currency: "USD", OrderID: "x",
	})
	if err != nil {
		t.Fatalf("CreateInvoice after retry: %v", err)
	}
	if invoice.UUID != "ok" {
		t.Errorf("expected invoice from retried response; got %+v", invoice)
	}
	if got := len(fake.Requests()); got != 2 {
		t.Errorf("expected 2 transport calls; got %d", got)
	}
}

func TestRetry_DoesNotRetry4xx(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state": 1, "message": "Wrong key",
	}, 401)
	client, err := heleket.NewPaymentClient(testMerchantID, testAPIKey,
		heleket.WithTransport(fake),
		heleket.WithMaxRetries(3),
	)
	if err != nil {
		t.Fatalf("NewPaymentClient: %v", err)
	}
	_, err = client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{
		Amount: "1", Currency: "USD", OrderID: "x",
	})
	if !errors.Is(err, heleket.ErrAPI) {
		t.Errorf("errors.Is(err, ErrAPI) = false; want true (got %v)", err)
	}
	if got := len(fake.Requests()); got != 1 {
		t.Errorf("expected 1 call (no retry on 4xx); got %d", got)
	}
}

func TestSentinelErrors_ValidationAndAPI(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state":  1,
		"errors": map[string]any{"amount": []string{"required"}},
	}, 422)
	client := newPaymentWithFake(t, fake)

	_, err := client.CreateInvoice(context.Background(), heleket.CreateInvoiceRequest{})
	if !errors.Is(err, heleket.ErrValidation) {
		t.Errorf("errors.Is(err, ErrValidation) = false; want true")
	}
	if !errors.Is(err, heleket.ErrAPI) {
		t.Errorf("ValidationError should also satisfy errors.Is(err, ErrAPI)")
	}
}

func TestGetInfo_ReturnsErrIdentifierRequired(t *testing.T) {
	client := newPaymentWithFake(t, testutil.NewFakeTransport())
	_, err := client.GetInfo(context.Background(), heleket.InfoOptions{})
	if !errors.Is(err, heleket.ErrIdentifierRequired) {
		t.Errorf("errors.Is(err, ErrIdentifierRequired) = false; want true (got %v)", err)
	}
}

func TestGetExchangeRates_IssuesSignedGetWithEmptyBody(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state": 0,
		"result": []map[string]any{
			{"from": "USD", "to": "BTC", "course": "0.000016", "source": "Binance"},
		},
	}, 200)
	client := newPaymentWithFake(t, fake)

	rates, err := client.GetExchangeRates(context.Background(), "USD")
	if err != nil {
		t.Fatalf("GetExchangeRates: %v", err)
	}
	if len(rates) != 1 || rates[0].From != "USD" || rates[0].To != "BTC" {
		t.Errorf("unexpected rates: %+v", rates)
	}

	req := fake.LastRequest()
	if req.Method != "GET" {
		t.Errorf("method = %q; want GET", req.Method)
	}
	if req.URL != "https://api.heleket.com/v1/exchange-rate/USD/list" {
		t.Errorf("url = %q", req.URL)
	}
	if len(req.Body) != 0 {
		t.Errorf("a GET carries its input in the path; want empty body, got %q", req.Body)
	}
	if got := req.Headers.Get("merchant"); got != testMerchantID {
		t.Errorf("merchant header = %q; want %q", got, testMerchantID)
	}
	// md5(base64("") + key) == md5(key)
	want := md5.Sum([]byte(testAPIKey))
	if got := req.Headers.Get("sign"); got != hex.EncodeToString(want[:]) {
		t.Errorf("sign for empty GET body = %q; want %q", got, hex.EncodeToString(want[:]))
	}
}

func TestGetAmlLinks_SendsSignedPostByUUID(t *testing.T) {
	links := []map[string]any{
		{"link": "https://some.link/1", "expired_at": "2025-10-23T18:23:40.000000Z", "status": "completed"},
		{"link": "https://some.link/2", "expired_at": "2026-05-13T11:32:38.000000Z", "status": "init"},
	}
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state":  0,
		"result": links,
	}, 200)
	client := newPaymentWithFake(t, fake)

	out, err := client.GetAmlLinks(context.Background(), heleket.InfoOptions{UUID: "uuid-1"})
	if err != nil {
		t.Fatalf("GetAmlLinks: %v", err)
	}
	if len(out) != 2 || out[0].Link != "https://some.link/1" || out[0].Status != heleket.AmlLinkStatusCompleted {
		t.Errorf("unexpected links: %+v", out)
	}

	req := fake.LastRequest()
	if req.Method != "POST" {
		t.Errorf("method = %q; want POST", req.Method)
	}
	if req.URL != "https://api.heleket.com/v1/payment/aml-links" {
		t.Errorf("url = %q", req.URL)
	}
	if got := req.Headers.Get("merchant"); got != testMerchantID {
		t.Errorf("merchant header = %q; want %q", got, testMerchantID)
	}

	expectedBody, _ := json.Marshal(heleket.InfoOptions{UUID: "uuid-1"})
	if string(req.Body) != string(expectedBody) {
		t.Errorf("body = %q; want %q", req.Body, expectedBody)
	}
	expectedSign := md5.Sum([]byte(base64.StdEncoding.EncodeToString(expectedBody) + testAPIKey))
	if got := req.Headers.Get("sign"); got != hex.EncodeToString(expectedSign[:]) {
		t.Errorf("sign = %q; want %q", got, hex.EncodeToString(expectedSign[:]))
	}
}

func TestGetAmlLinks_LooksUpByOrderID(t *testing.T) {
	fake := testutil.NewFakeTransport().EnqueueJSON(map[string]any{
		"state":  0,
		"result": []any{},
	}, 200)
	client := newPaymentWithFake(t, fake)

	if _, err := client.GetAmlLinks(context.Background(), heleket.InfoOptions{OrderID: "order-1"}); err != nil {
		t.Fatalf("GetAmlLinks: %v", err)
	}

	req := fake.LastRequest()
	if req.URL != "https://api.heleket.com/v1/payment/aml-links" {
		t.Errorf("url = %q", req.URL)
	}
	want, _ := json.Marshal(heleket.InfoOptions{OrderID: "order-1"})
	if string(req.Body) != string(want) {
		t.Errorf("body = %q; want %q", req.Body, want)
	}
}

func TestGetAmlLinks_RequiresUUIDOrOrderID(t *testing.T) {
	client := newPaymentWithFake(t, testutil.NewFakeTransport())
	_, err := client.GetAmlLinks(context.Background(), heleket.InfoOptions{})
	if !errors.Is(err, heleket.ErrIdentifierRequired) {
		t.Errorf("errors.Is(err, ErrIdentifierRequired) = false; want true (got %v)", err)
	}
}
