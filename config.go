package heleket

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Version is the SDK version, surfaced in the User-Agent header.
const Version = "0.3.0"

// DefaultBaseURL is the production Heleket API base URL.
const DefaultBaseURL = "https://api.heleket.com"

// DefaultTimeout is the default total request timeout.
const DefaultTimeout = 30 * time.Second

// DefaultMaxRetries is the default retry count for transport errors and HTTP
// 5xx responses. Retries use exponential backoff (100ms, 200ms, 400ms).
const DefaultMaxRetries = 3

// DefaultMaxResponseBytes caps the response body the SDK is willing to read,
// preventing a hostile or misconfigured server from exhausting memory.
const DefaultMaxResponseBytes = 16 << 20 // 16 MiB

// defaultUserAgent is sent on every request unless WithUserAgent overrides it.
const defaultUserAgent = "heleket-go-sdk/" + Version

// Config holds the per-client configuration. Construct it via NewPaymentClient
// / NewPayoutClient with the Option-style helpers; do NOT mutate Config fields
// after a client has been constructed — the client snapshots its config at
// build time and later changes are ignored at best, racy at worst.
type Config struct {
	MerchantID       string
	APIKey           string
	BaseURL          string
	Timeout          time.Duration
	Debug            bool
	Logger           *slog.Logger
	Transport        Transport
	UserAgent        string
	MaxRetries       int
	MaxResponseBytes int64
}

// Option mutates a Config during client construction.
type Option func(*Config)

// WithBaseURL overrides the API base URL. The URL must use https://, except
// for loopback hosts (localhost, 127.0.0.1, ::1) where http:// is also
// accepted for local testing. Trailing slashes are stripped.
func WithBaseURL(baseURL string) Option {
	return func(c *Config) {
		c.BaseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithTimeout overrides the default per-request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Config) { c.Timeout = d }
}

// WithDebug toggles debug-level logging of requests and responses.
// When enabled and no logger has been set, the SDK writes to os.Stderr.
func WithDebug(enabled bool) Option {
	return func(c *Config) { c.Debug = enabled }
}

// WithLogger sets the *slog.Logger used by the SDK. The logger receives debug
// messages only when WithDebug(true) is also active. The "sign" request header
// and the API key are NEVER passed to the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Config) { c.Logger = logger }
}

// WithTransport injects a custom Transport — for tests or alternative HTTP
// backends. When unset, the SDK uses an HTTPTransport wrapping a default
// http.Client with the configured timeout and a no-cross-host-redirect policy.
func WithTransport(t Transport) Option {
	return func(c *Config) { c.Transport = t }
}

// WithHTTPClient is a convenience that wraps the given *http.Client in an
// HTTPTransport. Use this when you want to customise *http.Client (TLS config,
// proxy, instrumentation) but keep the SDK's default wire behaviour.
//
// The supplied client's CheckRedirect is NOT modified; if you set one, make
// sure it does not follow cross-host redirects, otherwise the SDK's signed
// "sign" header may be forwarded to an attacker-controlled host.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) { c.Transport = NewHTTPTransport(client) }
}

// WithUserAgent appends additional tokens to the User-Agent header. The SDK
// always sends "heleket-go-sdk/<Version>" plus any extra tokens, in the order
// added. Typical use: WithUserAgent("myapp/1.2 (+https://example.com)").
func WithUserAgent(tokens string) Option {
	return func(c *Config) {
		tokens = strings.TrimSpace(tokens)
		if tokens == "" {
			return
		}
		if c.UserAgent == "" {
			c.UserAgent = defaultUserAgent + " " + tokens
		} else {
			c.UserAgent = c.UserAgent + " " + tokens
		}
	}
}

// WithMaxRetries sets how many times the SDK will retry on transport errors
// (DNS, connection-reset, timeout) and HTTP 5xx responses. Set to 0 to
// disable retries entirely. Retries use exponential backoff and respect ctx.
//
// Heleket rejects duplicate OrderIDs and returns the existing record, so
// retrying create-* endpoints with the same OrderID is safe.
func WithMaxRetries(n int) Option {
	return func(c *Config) { c.MaxRetries = n }
}

// WithMaxResponseBytes overrides the per-response body cap. Bodies larger
// than this limit produce *HTTPError instead of being read fully.
func WithMaxResponseBytes(n int64) Option {
	return func(c *Config) { c.MaxResponseBytes = n }
}

// newConfig builds a Config from credentials + options, applying defaults and
// validating required fields. Returns Config by value so the client can
// snapshot it — mutations to the original Config struct after construction
// are not reflected in the live client.
func newConfig(merchantID, apiKey string, opts ...Option) (Config, error) {
	merchantID = strings.TrimSpace(merchantID)
	apiKey = strings.TrimSpace(apiKey)

	if merchantID == "" {
		return Config{}, errors.New("heleket: merchantID must not be empty")
	}
	if apiKey == "" {
		return Config{}, errors.New("heleket: apiKey must not be empty")
	}

	cfg := &Config{
		MerchantID:       merchantID,
		APIKey:           apiKey,
		BaseURL:          DefaultBaseURL,
		Timeout:          DefaultTimeout,
		MaxRetries:       DefaultMaxRetries,
		MaxResponseBytes: DefaultMaxResponseBytes,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Timeout <= 0 {
		return Config{}, errors.New("heleket: timeout must be > 0")
	}
	if cfg.MaxRetries < 0 {
		return Config{}, errors.New("heleket: MaxRetries must be >= 0")
	}
	if cfg.MaxResponseBytes <= 0 {
		return Config{}, errors.New("heleket: MaxResponseBytes must be > 0")
	}
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return Config{}, err
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}
	if cfg.Logger == nil {
		if cfg.Debug {
			cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		} else {
			cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
		}
	}
	if cfg.Transport == nil {
		cfg.Transport = NewHTTPTransport(newDefaultHTTPClient(cfg.Timeout))
	}
	// Propagate MaxResponseBytes into the default HTTPTransport. Fully-custom
	// Transport implementations are responsible for their own caps.
	if ht, ok := cfg.Transport.(*HTTPTransport); ok {
		ht.MaxResponseBytes = cfg.MaxResponseBytes
	}
	return *cfg, nil
}

// validateBaseURL enforces https:// for non-loopback hosts and rejects URLs
// with userinfo, query strings, or fragments. http:// is allowed only for
// loopback hostnames (localhost, 127.0.0.1, ::1) to support local testing.
func validateBaseURL(raw string) error {
	if raw == "" {
		return errors.New("heleket: base URL must not be empty")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("heleket: invalid base URL: %w", err)
	}
	if u.Host == "" {
		return errors.New("heleket: base URL must include a host")
	}
	if u.User != nil {
		return errors.New("heleket: base URL must not include userinfo (user:pass@)")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return errors.New("heleket: base URL must not include a query string or fragment")
	}
	hostname := strings.ToLower(u.Hostname())
	isLoopback := hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
	switch u.Scheme {
	case "https":
		return nil
	case "http":
		if !isLoopback {
			return fmt.Errorf("heleket: base URL %q must use https:// (http:// allowed only for loopback hosts)", raw)
		}
		return nil
	default:
		return fmt.Errorf("heleket: base URL %q must use https:// scheme", raw)
	}
}

// newDefaultHTTPClient builds the SDK's default *http.Client. It uses a clone
// of http.DefaultTransport (so it doesn't share connection pool state with
// unrelated code) and a CheckRedirect that blocks all redirects — Heleket's
// API does not redirect, and forwarding the signed "sign" header to another
// host would leak authentication.
func newDefaultHTTPClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
