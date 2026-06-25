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
		fmt.Fprintln(os.Stderr, "Usage: refund <invoice-uuid> <refund-address>")
		os.Exit(1)
	}
	uuid, address := os.Args[1], os.Args[2]

	ctx, cancel := exampleutil.Context()
	defer cancel()

	// Refund lives on the PAYOUT client: /v1/payment/refund is signed with the
	// payout API key, not the payment key.
	client := exampleutil.PayoutFromEnv()
	refund, err := client.Refund(ctx, heleket.RefundRequest{
		UUID:       uuid,
		Address:    address,
		IsSubtract: true,
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
		fmt.Fprintln(os.Stderr, "refund:", err)
		os.Exit(3)
	}
	exampleutil.PrintResult("Refund requested", refund)
}
