package heleket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// baseClient holds the shared plumbing used by PaymentClient and PayoutClient:
// JSON encoding, signing, transport dispatch, debug logging, retry handling,
// and the response-to-error translation rules.
//
// The Config is stored by value so callers can't race-modify the live
// configuration after the client is built.
type baseClient struct {
	config Config
}

// post sends a signed POST request and decodes the API "result" field into out.
// out may be nil for endpoints whose only meaningful return is the success
// state (e.g. ResendWebhook).
//
// Returns:
//   - *HTTPError on transport failure (DNS, timeout, response too large)
//   - *ValidationError on HTTP 422 with a parsable errors map
//   - *APIError on any other state != 0
//
// Transport errors and HTTP 5xx responses are retried up to MaxRetries times
// using exponential backoff (100ms, 200ms, 400ms, ...), bounded by ctx.
func (c *baseClient) post(ctx context.Context, path string, params any, out any) error {
	body, err := encodeBody(params)
	if err != nil {
		return err
	}
	sign := Sign(body, c.config.APIKey)

	url := c.config.BaseURL + path
	headers := http.Header{}
	headers.Set("merchant", c.config.MerchantID)
	headers.Set("sign", sign)
	headers.Set("Content-Type", "application/json")
	if c.config.UserAgent != "" {
		headers.Set("User-Agent", c.config.UserAgent)
	}

	var lastErr error
	maxAttempts := c.config.MaxRetries + 1
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			if err := sleepBackoff(ctx, attempt); err != nil {
				if lastErr != nil {
					return lastErr
				}
				return &HTTPError{Cause: err}
			}
		}

		debugRequest(c.config.Logger, c.config.Debug, "POST", url, body)
		resp, err := c.config.Transport.RoundTrip(ctx, "POST", url, headers, body)
		if err != nil {
			lastErr = &HTTPError{Cause: err}
			if !isRetryableTransportError(err) || attempt == maxAttempts-1 {
				return lastErr
			}
			continue
		}

		debugResponse(c.config.Logger, c.config.Debug, resp.StatusCode, resp.Body)

		err = parseResponse(resp.StatusCode, resp.Body, out)
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryableResponseError(err) || attempt == maxAttempts-1 {
			return lastErr
		}
	}
	return lastErr
}

// isRetryableTransportError returns true for errors worth retrying.
// Context cancellation is not retryable — the caller has given up.
func isRetryableTransportError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// Body-too-large is a server / protocol problem, retrying won't help.
	if errors.Is(err, errBodyTooLarge) {
		return false
	}
	return true
}

// isRetryableResponseError returns true for *APIError with HTTPStatus >= 500.
// ValidationError (422) and other 4xx errors are caller-fault and not retried.
func isRetryableResponseError(err error) bool {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae.HTTPStatus >= 500
	}
	return false
}

// sleepBackoff waits for an exponential backoff interval based on attempt
// number (1-indexed). attempt 1 → 100ms, 2 → 200ms, 3 → 400ms, capped at 5s.
// Honours ctx so a cancelled request returns quickly.
func sleepBackoff(ctx context.Context, attempt int) error {
	base := 100 * time.Millisecond
	d := base << (attempt - 1) // 100, 200, 400, 800, ...
	const maxD = 5 * time.Second
	if d > maxD {
		d = maxD
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// encodeBody marshals the request struct. Nil params (or an explicit empty
// struct) yield an empty byte slice — matching PHP's empty-params signing rule.
func encodeBody(params any) ([]byte, error) {
	if params == nil {
		return []byte{}, nil
	}
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("heleket: encode request body: %w", err)
	}
	// json.Marshal(nil) returns "null"; we want empty bytes for "no params".
	if string(body) == "null" {
		return []byte{}, nil
	}
	return body, nil
}

// envelope is the standard Heleket response wrapper.
type envelope struct {
	State   int                 `json:"state"`
	Result  json.RawMessage     `json:"result,omitempty"`
	Message string              `json:"message,omitempty"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

// parseResponse converts an HTTP response into either a populated `out` or an
// SDK error. HTTP 422 produces *ValidationError; any other state != 0 produces
// *APIError. Decoding failures of `result` propagate as wrapped errors.
func parseResponse(statusCode int, body []byte, out any) error {
	var env envelope
	hasJSON := len(body) > 0 && json.Unmarshal(body, &env) == nil

	if statusCode == http.StatusUnprocessableEntity {
		fields := env.Errors
		if fields == nil {
			fields = map[string][]string{}
		}
		return &ValidationError{
			APIError: &APIError{
				Message:    "validation failed",
				HTTPStatus: statusCode,
				RawBody:    body,
			},
			Fields: fields,
		}
	}

	if !hasJSON {
		return &APIError{
			Message:    fmt.Sprintf("non-JSON response (HTTP %d)", statusCode),
			HTTPStatus: statusCode,
			RawBody:    body,
		}
	}

	if env.State != 0 {
		message := env.Message
		if message == "" {
			message = fmt.Sprintf("api error (state=%d)", env.State)
		}
		return &APIError{Message: message, HTTPStatus: statusCode, RawBody: body}
	}

	if out == nil {
		return nil
	}
	if len(env.Result) == 0 || string(env.Result) == "null" {
		// Endpoints that return empty results are still successful.
		return nil
	}
	if err := json.Unmarshal(env.Result, out); err != nil {
		return fmt.Errorf("heleket: decode response result: %w", err)
	}
	return nil
}
