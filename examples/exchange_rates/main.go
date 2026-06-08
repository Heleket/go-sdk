package main

import (
	"fmt"
	"os"

	"github.com/heleket/go-sdk/internal/exampleutil"
)

func main() {
	currency := "USD"
	if len(os.Args) > 1 {
		currency = os.Args[1]
	}

	ctx, cancel := exampleutil.Context()
	defer cancel()

	client := exampleutil.PaymentFromEnv()
	rates, err := client.GetExchangeRates(ctx, currency)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get exchange rates:", err)
		os.Exit(3)
	}
	exampleutil.PrintResult("Exchange rates for "+currency, rates)
}
