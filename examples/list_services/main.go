package main

import (
	"fmt"
	"os"

	"github.com/heleket/go-sdk/internal/exampleutil"
)

func main() {
	kind := "payment"
	if len(os.Args) > 1 {
		kind = os.Args[1]
	}

	ctx, cancel := exampleutil.Context()
	defer cancel()

	switch kind {
	case "payment":
		services, err := exampleutil.PaymentFromEnv().ListServices(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "list services:", err)
			os.Exit(3)
		}
		exampleutil.PrintResult("Payment services", services)
	case "payout":
		services, err := exampleutil.PayoutFromEnv().ListServices(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "list services:", err)
			os.Exit(3)
		}
		exampleutil.PrintResult("Payout services", services)
	default:
		fmt.Fprintln(os.Stderr, "Usage: list_services [payment|payout]")
		os.Exit(1)
	}
}
