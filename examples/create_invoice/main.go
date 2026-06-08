package main

import (
	"errors"
	"fmt"
	"os"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/internal/exampleutil"
)

func main() {
	ctx, cancel := exampleutil.Context()
	defer cancel()

	client := exampleutil.PaymentFromEnv()

	invoice, err := client.CreateInvoice(ctx, heleket.CreateInvoiceRequest{
		Amount:      "15.00",
		Currency:    "USD",
		OrderID:     exampleutil.RandomOrderID("demo"),
		Lifetime:    3600,
		URLCallback: exampleutil.EnvOr("HELEKET_WEBHOOK_URL", ""),
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
		fmt.Fprintln(os.Stderr, "create invoice:", err)
		os.Exit(3)
	}

	exampleutil.PrintResult("Invoice created", invoice)
	fmt.Println("\nPayment page:", invoice.URL)
}
