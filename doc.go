// Package heleket is a Go SDK for the Heleket cryptocurrency payment API.
//
// See https://doc.heleket.com for the upstream documentation.
//
// # Quickstart
//
//	import "github.com/heleket/go-sdk"
//
//	client, err := heleket.NewPaymentClient(merchantID, paymentKey)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	invoice, err := client.CreateInvoice(ctx, heleket.CreateInvoiceRequest{
//	    Amount:   "15.00",
//	    Currency: "USD",
//	    OrderID:  "order-42",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(invoice.URL)
//
// # Webhooks
//
// The webhook subpackage handles incoming Heleket callbacks:
//
//	import "github.com/heleket/go-sdk/webhook"
//
//	verifier := webhook.NewVerifier(paymentKey)
//	payload, err := verifier.VerifyRaw(rawBody)
//
// # Concurrency
//
// PaymentClient and PayoutClient are safe for concurrent use across goroutines.
//
// # Errors
//
// All API failures return typed errors that can be inspected with errors.As:
//
//	var ve *heleket.ValidationError
//	if errors.As(err, &ve) {
//	    for field, msgs := range ve.Fields { /* ... */ }
//	}
//
// Or matched against sentinel values with errors.Is:
//
//	if errors.Is(err, heleket.ErrTransport)  { /* retry */ }
//	if errors.Is(err, heleket.ErrValidation) { /* show fields to user */ }
package heleket
