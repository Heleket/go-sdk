package main

import (
	"fmt"
	"os"

	"github.com/heleket/go-sdk/internal/exampleutil"
)

func main() {
	ctx, cancel := exampleutil.Context()
	defer cancel()

	client := exampleutil.PaymentFromEnv()
	balance, err := client.GetBalance(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get balance:", err)
		os.Exit(3)
	}
	exampleutil.PrintResult("Balance", balance)
}
