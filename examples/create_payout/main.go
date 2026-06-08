package main

import (
	"errors"
	"fmt"
	"os"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/internal/exampleutil"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: create_payout <amount> <address>")
		os.Exit(1)
	}
	amount, address := os.Args[1], os.Args[2]

	ctx, cancel := exampleutil.Context()
	defer cancel()

	client := exampleutil.PayoutFromEnv()
	payout, err := client.CreatePayout(ctx, heleket.CreatePayoutRequest{
		Amount:      amount,
		Currency:    "USDT",
		Network:     "TRON",
		OrderID:     exampleutil.RandomOrderID("payout"),
		Address:     address,
		IsSubtract:  true,
		URLCallback: exampleutil.EnvOr("HELEKET_PAYOUT_WEBHOOK_URL", ""),
	})
	if err != nil {
		var ve *heleket.ValidationError
		if errors.As(err, &ve) {
			fmt.Fprintln(os.Stderr, "Validation failed:")
			for field, msgs := range ve.Fields {
				fmt.Fprintf(os.Stderr, "  %s: %v\n", field, msgs)
			}
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "create payout:", err)
		os.Exit(3)
	}
	exampleutil.PrintResult("Payout created", payout)
}
