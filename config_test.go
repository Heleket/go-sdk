package heleket_test

import (
	"strings"
	"testing"
	"time"

	heleket "github.com/heleket/go-sdk"
)

func TestNewPaymentClient_RejectsEmptyCredentials(t *testing.T) {
	if _, err := heleket.NewPaymentClient("   ", "key"); err == nil {
		t.Error("expected error for empty merchantID")
	}
	if _, err := heleket.NewPaymentClient("merchant", ""); err == nil {
		t.Error("expected error for empty apiKey")
	}
}

func TestNewPaymentClient_RejectsNegativeTimeout(t *testing.T) {
	_, err := heleket.NewPaymentClient("merchant", "key", heleket.WithTimeout(0))
	if err == nil || !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error; got %v", err)
	}
}

func TestNewPaymentClient_AcceptsValidOptions(t *testing.T) {
	c, err := heleket.NewPaymentClient("merchant", "key",
		heleket.WithBaseURL("https://example.com/"),
		heleket.WithTimeout(45*time.Second),
		heleket.WithDebug(true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewPayoutClient_Construction(t *testing.T) {
	c, err := heleket.NewPayoutClient("merchant", "payout-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithBaseURL_RejectsInsecureRemoteHost(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"plain http remote", "http://api.attacker.com"},
		{"ftp scheme", "ftp://api.heleket.com"},
		{"empty host", "https:///path"},
		{"userinfo", "https://user:pass@api.heleket.com"},
		{"query string", "https://api.heleket.com/?evil=1"},
		{"fragment", "https://api.heleket.com/#frag"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := heleket.NewPaymentClient("m", "k", heleket.WithBaseURL(tc.url))
			if err == nil {
				t.Errorf("expected error for %q", tc.url)
			}
		})
	}
}

func TestWithBaseURL_AcceptsLoopbackHTTP(t *testing.T) {
	for _, url := range []string{
		"http://127.0.0.1:8080",
		"http://localhost:9999/",
		"https://api.heleket.com",
	} {
		if _, err := heleket.NewPaymentClient("m", "k", heleket.WithBaseURL(url)); err != nil {
			t.Errorf("expected %q to be accepted; got %v", url, err)
		}
	}
}

func TestWithMaxRetries_RejectsNegative(t *testing.T) {
	if _, err := heleket.NewPaymentClient("m", "k", heleket.WithMaxRetries(-1)); err == nil {
		t.Error("expected error for negative retry count")
	}
}
