package main

import (
	"fmt"
	"os"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/internal/exampleutil"
)

func main() {
	ctx, cancel := exampleutil.Context()
	defer cancel()

	client := exampleutil.PaymentFromEnv()

	wallet, err := client.CreateStaticWallet(ctx, heleket.CreateStaticWalletRequest{
		Currency: "USDT",
		Network:  "tron",
		OrderID:  exampleutil.RandomOrderID("wallet"),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "create static wallet:", err)
		os.Exit(3)
	}

	exampleutil.PrintResult("Static wallet", wallet)
	fmt.Println("\nAddress:", wallet.Address)
}
