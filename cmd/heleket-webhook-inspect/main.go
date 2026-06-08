// Heleket webhook inspector.
//
// Reads a JSON webhook payload from stdin (or --file=<path>), verifies its
// signature against the API key, and prints a human-readable report.
//
// The API key is read from HELEKET_PAYMENT_KEY (or HELEKET_PAYOUT_KEY when
// --type=payout) by default. Pass --key= to override, but be aware that
// command-line arguments are visible to other users via `ps` and shell history.
//
//	cat webhook.json | heleket-webhook-inspect
//	heleket-webhook-inspect --file=webhook.json
//	heleket-webhook-inspect --type=payout < webhook.json
//	heleket-webhook-inspect --key=$KEY --file=webhook.json   # explicit override
//
// Exit codes:
//
//	0 — payload valid, signature verified
//	1 — payload could not be parsed
//	2 — signature does not match
//	3 — missing arguments
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/webhook"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

func main() {
	os.Exit(run())
}

func run() int {
	keyFlag := flag.String("key", "", "Heleket API key (overrides env; visible in `ps` — prefer env vars)")
	file := flag.String("file", "", "Path to JSON payload (omit to read stdin)")
	declaredType := flag.String("type", "", "Expected type for sanity check: payment|wallet|payout")
	flag.Parse()

	key := resolveKey(*keyFlag, *declaredType)
	if key == "" {
		fmt.Fprintln(os.Stderr, "No API key provided.")
		fmt.Fprintln(os.Stderr, "Set HELEKET_PAYMENT_KEY (default) or HELEKET_PAYOUT_KEY (with --type=payout),")
		fmt.Fprintln(os.Stderr, "or pass --key=<api-key> (visible in `ps` — env is safer).")
		return 3
	}
	if *keyFlag != "" {
		fmt.Fprintln(os.Stderr, "Warning: --key= leaks the API key into `ps` and shell history. Prefer HELEKET_PAYMENT_KEY / HELEKET_PAYOUT_KEY env vars.")
	}

	body, err := readPayload(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		fmt.Fprintln(os.Stderr, "Input is not valid JSON:", err)
		return 1
	}

	fmt.Println("Heleket webhook inspector")
	fmt.Println("----------------------------------------")
	printField("type", raw["type"])
	printField("uuid", raw["uuid"])
	printField("order_id", raw["order_id"])
	printField("status", raw["status"])
	printField("amount", raw["amount"])
	printField("network", raw["network"])
	printField("txid", raw["txid"])
	if v, ok := raw["is_final"]; ok {
		printField("is_final", boolish(v))
	} else {
		printField("is_final", "(missing)")
	}
	if v, ok := raw["sign"].(string); ok && len(v) > 16 {
		printField("sign", v[:16]+"…")
	} else if v, ok := raw["sign"].(string); ok {
		printField("sign", v)
	} else {
		printField("sign", "(missing)")
	}

	if *declaredType != "" && raw["type"] != *declaredType {
		fmt.Fprintf(os.Stderr, "\nWarning: --type=%s but payload type is %v\n", *declaredType, raw["type"])
	}

	if _, err := webhook.NewVerifier(key).VerifyRaw(body); err != nil {
		var se *heleket.SignatureError
		if errors.As(err, &se) {
			fmt.Printf("\nsignature: %sINVALID%s — %s\n", colorRed, colorReset, se.Reason)
		} else {
			fmt.Printf("\nsignature: %sINVALID%s — %s\n", colorRed, colorReset, err.Error())
		}
		fmt.Println("Hints:")
		fmt.Println("  - Are you using the same API key that was configured when the webhook was sent?")
		fmt.Println("  - Payment webhooks sign with the payment key; payout webhooks sign with the payout key.")
		fmt.Println("  - If the payload comes from a non-Heleket-sourced re-serializer, JSON byte-fidelity may differ.")
		return 2
	}

	fmt.Printf("\nsignature: %svalid%s\n", colorGreen, colorReset)
	return 0
}

func readPayload(path string) ([]byte, error) {
	if path != "" {
		return os.ReadFile(path)
	}
	return io.ReadAll(os.Stdin)
}

// resolveKey picks the API key in priority order: --key flag, then the env
// var matching the declared type. Payout webhooks use HELEKET_PAYOUT_KEY;
// everything else defaults to HELEKET_PAYMENT_KEY.
func resolveKey(flagKey, declaredType string) string {
	if flagKey != "" {
		return flagKey
	}
	if declaredType == "payout" {
		if v := os.Getenv("HELEKET_PAYOUT_KEY"); v != "" {
			return v
		}
	}
	return os.Getenv("HELEKET_PAYMENT_KEY")
}

func printField(name string, value any) {
	if value == nil {
		fmt.Printf("  %-10s %s\n", name, "(missing)")
		return
	}
	switch v := value.(type) {
	case string:
		fmt.Printf("  %-10s %s\n", name, v)
	case bool:
		fmt.Printf("  %-10s %s\n", name, boolish(v))
	default:
		b, _ := json.Marshal(v)
		fmt.Printf("  %-10s %s\n", name, string(b))
	}
}

func boolish(v any) string {
	if b, ok := v.(bool); ok {
		if b {
			return "yes"
		}
		return "no"
	}
	return fmt.Sprintf("%v", v)
}
