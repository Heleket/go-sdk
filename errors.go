package heleket

import (
	"errors"
	"fmt"
)

// Sentinel errors. Use with errors.Is to detect error categories without
// extracting a typed pointer.
//
//	if errors.Is(err, heleket.ErrValidation) { ... }
//	if errors.Is(err, heleket.ErrTransport)  { ... }
//
// For field-level details (ValidationError.Fields, APIError.HTTPStatus, ...)
// still use errors.As to recover the concrete type.
var (
	// ErrAPI matches any *APIError, including *ValidationError.
	ErrAPI = errors.New("heleket: api error")
	// ErrValidation matches *ValidationError specifically (HTTP 422).
	ErrValidation = errors.New("heleket: validation error")
	// ErrTransport matches *HTTPError (DNS, TCP, TLS, timeout).
	ErrTransport = errors.New("heleket: transport error")
	// ErrSignature matches *SignatureError from the webhook subpackage.
	ErrSignature = errors.New("heleket: signature error")
	// ErrPayloadDecode matches *PayloadDecodeError when a signature-valid
	// webhook payload cannot be decoded into the typed Payload struct.
	ErrPayloadDecode = errors.New("heleket: payload decode error")

	// ErrIdentifierRequired is returned when an endpoint requires UUID OR
	// OrderID and neither was set.
	ErrIdentifierRequired = errors.New("heleket: one of UUID or OrderID must be set")
	// ErrInvalidTestWebhookType is returned when TestWebhookRequest.Type is
	// neither "payment" nor "wallet".
	ErrInvalidTestWebhookType = errors.New(`heleket: TestWebhookRequest.Type must be "payment" or "wallet"`)
)

// APIError is returned when Heleket responds with state != 0 (a business
// error). HTTPStatus and RawBody carry the underlying response so callers can
// log or surface the exact server reply.
type APIError struct {
	Message    string
	HTTPStatus int
	RawBody    []byte
}

func (e *APIError) Error() string {
	if e.HTTPStatus > 0 {
		return fmt.Sprintf("heleket api error (HTTP %d): %s", e.HTTPStatus, e.Message)
	}
	return "heleket api error: " + e.Message
}

// Is reports whether target is ErrAPI. errors.Is(err, ErrAPI) recognises any
// APIError regardless of the underlying HTTPStatus.
func (e *APIError) Is(target error) bool {
	return target == ErrAPI
}

// ValidationError is returned when Heleket responds with HTTP 422 and a
// field-level errors map. Fields is keyed by request field name.
type ValidationError struct {
	*APIError
	Fields map[string][]string
}

func (e *ValidationError) Error() string {
	return "heleket validation failed: " + summariseFields(e.Fields)
}

// Unwrap exposes the embedded APIError so errors.As recovers either type.
func (e *ValidationError) Unwrap() error {
	return e.APIError
}

// Is reports whether target is ErrValidation or ErrAPI. ValidationError is-a
// APIError, so both sentinels match.
func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation || target == ErrAPI
}

// HTTPError wraps transport-layer failures: DNS, TCP, TLS, timeout, broken
// pipe. The API was either never reached or no response was received.
type HTTPError struct {
	Cause error
}

func (e *HTTPError) Error() string {
	return "heleket http transport error: " + e.Cause.Error()
}

func (e *HTTPError) Unwrap() error {
	return e.Cause
}

func (e *HTTPError) Is(target error) bool {
	return target == ErrTransport
}

// SignatureError is returned by webhook.Verifier when the signature does not
// match. Treat as a hard security failure — never trust the payload.
type SignatureError struct {
	Reason string
}

func (e *SignatureError) Error() string {
	return "heleket signature error: " + e.Reason
}

func (e *SignatureError) Is(target error) bool {
	return target == ErrSignature
}

// PayloadDecodeError is returned by webhook.Verifier when the signature was
// valid but the typed Payload decode failed. This distinguishes "Heleket
// signed a payload we couldn't parse" from "signature did not verify."
type PayloadDecodeError struct {
	Cause error
}

func (e *PayloadDecodeError) Error() string {
	return "heleket payload decode error: " + e.Cause.Error()
}

func (e *PayloadDecodeError) Unwrap() error {
	return e.Cause
}

func (e *PayloadDecodeError) Is(target error) bool {
	return target == ErrPayloadDecode
}

func summariseFields(fields map[string][]string) string {
	if len(fields) == 0 {
		return "(no field details)"
	}
	out := ""
	for name, msgs := range fields {
		if out != "" {
			out += "; "
		}
		out += name + " ("
		for i, m := range msgs {
			if i > 0 {
				out += ", "
			}
			out += m
		}
		out += ")"
	}
	return out
}
