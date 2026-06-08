// Package exampleutil is shared scaffolding for the examples/ programs. It
// loads .env, builds clients, and pretty-prints results. It is intentionally
// in internal/ so external consumers don't accidentally depend on it.
package exampleutil

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	heleket "github.com/heleket/go-sdk"
)

// Context returns a Background context with a 60-second deadline — long enough
// for any Heleket API call, short enough that the example exits if the network
// is broken.
func Context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 60*time.Second)
}

// PaymentFromEnv returns a PaymentClient built from the merchant credentials
// in .env. Exits the process with a friendly message if any required variable
// is missing or the SDK rejects the configuration.
func PaymentFromEnv() *heleket.PaymentClient {
	loadEnv()
	client, err := heleket.NewPaymentClient(
		MustEnv("HELEKET_MERCHANT_ID"),
		MustEnv("HELEKET_PAYMENT_KEY"),
		heleket.WithDebug(os.Getenv("HELEKET_DEBUG") == "1"),
		heleket.WithUserAgent("heleket-go-examples/0.1"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "client construction:", err)
		os.Exit(1)
	}
	return client
}

// PayoutFromEnv returns a PayoutClient built from .env.
func PayoutFromEnv() *heleket.PayoutClient {
	loadEnv()
	client, err := heleket.NewPayoutClient(
		MustEnv("HELEKET_MERCHANT_ID"),
		MustEnv("HELEKET_PAYOUT_KEY"),
		heleket.WithDebug(os.Getenv("HELEKET_DEBUG") == "1"),
		heleket.WithUserAgent("heleket-go-examples/0.1"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "client construction:", err)
		os.Exit(1)
	}
	return client
}

// MustEnv returns the value of name or exits with a helpful error.
func MustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		fmt.Fprintln(os.Stderr, "missing env var:", name)
		fmt.Fprintln(os.Stderr, "Copy .env.example to .env and fill it in.")
		os.Exit(1)
	}
	return v
}

// EnvOr returns the value of name, or fallback if unset.
func EnvOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

// PrintResult dumps a value as indented JSON for human-readable example output.
func PrintResult(label string, value any) {
	fmt.Printf("\n=== %s ===\n", label)
	b, _ := json.MarshalIndent(value, "", "  ")
	fmt.Println(string(b))
}

// RandomOrderID returns a short hex suffix suitable for example order_id.
func RandomOrderID(prefix string) string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return prefix + "-" + hex.EncodeToString(b[:])
}

// loadEnv reads a sibling .env file if one exists. Tiny inline implementation,
// no dependency on godotenv. Existing env vars are not overwritten.
func loadEnv() {
	for _, path := range []string{".env", "../.env", "../../.env"} {
		if loadFile(path) {
			return
		}
	}
}

func loadFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		value := strings.TrimSpace(line[eq+1:])
		value = strings.Trim(value, `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
	return true
}
