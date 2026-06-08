// Webhook listener example. Run with:
//
//	go run ./examples/handle_webhook
//
// Then point your Heleket invoice's url_callback at http://<host>:8000/
// (use ngrok or similar for local development).
//
// The handler verifies the signature against HELEKET_PAYMENT_KEY and responds
// 200 only after successful verification.
package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	heleket "github.com/heleket/go-sdk"
	"github.com/heleket/go-sdk/internal/exampleutil"
	"github.com/heleket/go-sdk/webhook"
)

func main() {
	paymentKey := exampleutil.MustEnv("HELEKET_PAYMENT_KEY")
	verifier := webhook.NewVerifier(paymentKey)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		payload, err := verifier.VerifyRaw(body)
		if err != nil {
			var se *heleket.SignatureError
			if errors.As(err, &se) {
				logger.Warn("heleket signature mismatch", slog.String("reason", se.Reason))
			} else {
				logger.Warn("heleket webhook error", slog.String("error", err.Error()))
			}
			http.Error(w, "invalid signature", http.StatusBadRequest)
			return
		}

		logger.Info("heleket webhook received",
			slog.String("type", payload.Type),
			slog.String("order_id", payload.OrderID),
			slog.String("status", payload.Status),
			slog.Bool("is_final", payload.IsFinalStatus()),
		)

		// PRODUCTION: this example does NOT de-duplicate. Heleket will replay
		// events after retries and operators can resend manually. See
		// docs/06-webhooks.md ("Idempotency and replay protection") for the
		// recommended pattern using a unique (uuid, status) key in your DB.

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	})

	addr := ":" + exampleutil.EnvOr("PORT", "8000")
	logger.Info("listening", slog.String("addr", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		logger.Error("server stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
