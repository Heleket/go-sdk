package main

import (
	"fmt"
	"os"
	"regexp"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/internal/exampleutil"
)

// uuidPattern matches a canonical UUID so the example can accept either a
// payment UUID or a merchant order_id as its single argument.
var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run ./examples/aml_links <uuid-or-order-id>")
		os.Exit(1)
	}
	identifier := os.Args[1]

	// Caller can pass either a UUID or an order_id; we try UUID first.
	opts := heleket.InfoOptions{OrderID: identifier}
	if uuidPattern.MatchString(identifier) {
		opts = heleket.InfoOptions{UUID: identifier}
	}

	ctx, cancel := exampleutil.Context()
	defer cancel()

	client := exampleutil.PaymentFromEnv()
	links, err := client.GetAmlLinks(ctx, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "get aml links:", err)
		os.Exit(3)
	}
	exampleutil.PrintResult("AML links", links)

	// Hand each Link to the end user so they can complete the questionnaire.
	for _, link := range links {
		state := "in progress"
		if link.Status.IsFinal() {
			state = "final"
		}
		fmt.Printf("\n%s\n  status: %s (%s)\n  expires: %s\n", link.Link, link.Status, state, link.ExpiredAt)
	}
}
