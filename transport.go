package heleket

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Transport is the HTTP contract the SDK relies on. The default implementation
// wraps *http.Client. Tests inject internal/testutil.FakeTransport and merchants
// who want custom HTTP behaviour (proxies, retries, instrumentation) can plug in
// their own implementation via WithTransport.
type Transport interface {
	// RoundTrip sends an HTTP request and returns the response. Implementations
	// MUST forward the body bytes verbatim — the SDK has already signed those
	// exact bytes, and any mutation would invalidate the signature.
	RoundTrip(ctx context.Context, method, url string, headers http.Header, body []byte) (*Response, error)
}

// Response captures the status code, raw body, and headers of an HTTP reply.
// Body is retained verbatim so signature checks and the debug logger see the
// exact bytes Heleket sent.
type Response struct {
	StatusCode int
	Body       []byte
	Header     http.Header
}

// HTTPTransport is the default Transport. It uses an *http.Client and forwards
// request bodies unmodified. The caller-supplied *http.Client controls
// timeouts, TLS, proxies, and redirect policy.
type HTTPTransport struct {
	Client *http.Client

	// MaxResponseBytes caps the response body the transport will read. Set
	// by the SDK from Config.MaxResponseBytes. A value <= 0 means unlimited
	// (not recommended).
	MaxResponseBytes int64
}

// NewHTTPTransport returns a Transport wrapping an *http.Client. If client is
// nil, an empty *http.Client is created — note that this has no timeout and
// follows redirects by default. Prefer letting the SDK build its own via
// newConfig, which sets sensible defaults.
func NewHTTPTransport(client *http.Client) *HTTPTransport {
	if client == nil {
		client = &http.Client{}
	}
	return &HTTPTransport{
		Client:           client,
		MaxResponseBytes: DefaultMaxResponseBytes,
	}
}

// RoundTrip implements Transport.
func (t *HTTPTransport) RoundTrip(ctx context.Context, method, url string, headers http.Header, body []byte) (*Response, error) {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}
	for name, values := range headers {
		req.Header[name] = values
	}

	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	limit := t.MaxResponseBytes
	if limit <= 0 {
		limit = DefaultMaxResponseBytes
	}
	// Read limit+1 bytes to detect overflow.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("%w: limit was %d bytes", errBodyTooLarge, limit)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       raw,
		Header:     resp.Header,
	}, nil
}

// errBodyTooLarge identifies the specific transport failure when a response
// exceeded MaxResponseBytes. Tests can detect this via errors.Is.
var errBodyTooLarge = errors.New("heleket: response body too large")
