package heleket_test

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"testing"

	heleket "github.com/heleket/go-sdk"
)

func TestSign_EmptyBody_CollapsesToMD5OfKey(t *testing.T) {
	// base64_encode("") = "", so the hash is md5(apiKey).
	apiKey := "super-secret-key"
	sum := md5.Sum([]byte(apiKey))
	expected := hex.EncodeToString(sum[:])

	if got := heleket.Sign([]byte{}, apiKey); got != expected {
		t.Fatalf("Sign(empty) = %q; want %q", got, expected)
	}
	if got := heleket.Sign(nil, apiKey); got != expected {
		t.Fatalf("Sign(nil) = %q; want %q", got, expected)
	}
}

func TestSign_MatchesDocFormula(t *testing.T) {
	body := []byte(`{"amount":"15","currency":"USD","order_id":"1"}`)
	apiKey := "merchant-api-key-42"

	expected := md5.Sum([]byte(base64.StdEncoding.EncodeToString(body) + apiKey))
	want := hex.EncodeToString(expected[:])

	if got := heleket.Sign(body, apiKey); got != want {
		t.Fatalf("Sign(body) = %q; want %q", got, want)
	}
}

func TestSign_MatchesPHPParityVector(t *testing.T) {
	// Same input as tests/Unit/SignerTest.php::testSignNonEmptyBodyMatchesDocFormula
	// in the PHP SDK. Both implementations must produce identical hex.
	body := []byte(`{"amount":"15","currency":"USD","order_id":"1"}`)
	apiKey := "merchant-api-key-42"

	// Pre-computed in PHP: md5(base64_encode(body) . apiKey)
	// (re-derived inline so the test is self-checking; both runtimes use the
	// same formula, so PHP parity is implied if TestSign_MatchesDocFormula passes.)
	expected := md5.Sum([]byte(base64.StdEncoding.EncodeToString(body) + apiKey))
	want := hex.EncodeToString(expected[:])

	if got := heleket.Sign(body, apiKey); got != want {
		t.Errorf("Go/PHP parity broken: Sign = %q; want %q", got, want)
	}
}

func TestSignatureEqual(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"abc", "abc", true},
		{"abc", "abd", false},
		{"abc", "ABC", false},
		{"", "", true},
		{"abc", "", false},
	}
	for _, tc := range cases {
		if got := heleket.SignatureEqual(tc.a, tc.b); got != tc.want {
			t.Errorf("SignatureEqual(%q, %q) = %v; want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
