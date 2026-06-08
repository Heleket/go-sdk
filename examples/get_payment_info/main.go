package main

import (
	"fmt"
	"os"
	"regexp"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/internal/exampleutil"
)

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: get_payment_info <uuid-or-order-id>")
		os.Exit(1)
	}
	id := os.Args[1]

	ctx, cancel := exampleutil.Context()
	defer cancel()

	client := exampleutil.PaymentFromEnv()

	var opts heleket.InfoOptions
	if uuidRe.MatchString(id) {
		opts.UUID = id
	} else {
		opts.OrderID = id
	}

	info, err := client.GetInfo(ctx, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get info:", err)
		os.Exit(3)
	}

	exampleutil.PrintResult("Payment info", info)

	status := info.Status
	if status == "" {
		status = info.PaymentStatus
	}
	fmt.Println("\nStatus:", status)
	if status.IsFinal() {
		fmt.Println("(final)")
	} else {
		fmt.Println("(intermediate — keep polling or wait for webhook)")
	}
}
