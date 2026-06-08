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

	var dateFrom, dateTo string
	if len(os.Args) > 1 {
		dateFrom = os.Args[1]
	}
	if len(os.Args) > 2 {
		dateTo = os.Args[2]
	}

	page, err := client.ListHistory(ctx, heleket.HistoryOptions{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "list history:", err)
		os.Exit(3)
	}
	exampleutil.PrintResult("Page 1", page)

	if page.Paginate.NextCursor != "" {
		next, err := client.ListHistory(ctx, heleket.HistoryOptions{
			DateFrom: dateFrom,
			DateTo:   dateTo,
			Cursor:   page.Paginate.NextCursor,
		})
		if err == nil {
			exampleutil.PrintResult("Page 2", next)
		}
	}
}
