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
		fmt.Fprintln(os.Stderr, "Usage: get_payout_info <uuid-or-order-id>")
		os.Exit(1)
	}
	id := os.Args[1]

	ctx, cancel := exampleutil.Context()
	defer cancel()

	client := exampleutil.PayoutFromEnv()

	var opts heleket.InfoOptions
	if uuidRe.MatchString(id) {
		opts.UUID = id
	} else {
		opts.OrderID = id
	}

	info, err := client.GetInfo(ctx, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get payout info:", err)
		os.Exit(3)
	}
	exampleutil.PrintResult("Payout info", info)
}
