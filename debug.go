package heleket

import "log/slog"

// debugRequest logs an outbound request at Debug level. The "sign" header and
// the API key are intentionally not passed to slog — only the URL, method,
// and body bytes are.
func debugRequest(logger *slog.Logger, debug bool, method, url string, body []byte) {
	if !debug {
		return
	}
	logger.Debug("heleket request",
		slog.String("method", method),
		slog.String("url", url),
		slog.String("body", string(body)),
	)
}

// debugResponse logs an inbound response at Debug level.
func debugResponse(logger *slog.Logger, debug bool, statusCode int, body []byte) {
	if !debug {
		return
	}
	logger.Debug("heleket response",
		slog.Int("status", statusCode),
		slog.String("body", string(body)),
	)
}
