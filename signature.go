package heleket

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
)

// Sign returns the Heleket request signature for the given JSON body and API key.
//
// Formula (per https://doc.heleket.com/general/request-format):
//
//	sign = md5( base64_encode(json_body) . apiKey )
//
// For requests with no parameters, pass an empty byte slice — base64-encoding it
// yields the empty string, so the hash collapses to md5(apiKey).
//
// The body argument MUST be the exact bytes that will be sent over the wire.
// The same Sign function is used to produce outgoing request signatures and to
// verify incoming webhook signatures.
func Sign(body []byte, apiKey string) string {
	encoded := base64.StdEncoding.EncodeToString(body)
	sum := md5.Sum([]byte(encoded + apiKey))
	return hex.EncodeToString(sum[:])
}

// SignatureEqual performs a constant-time comparison of two hex signatures.
// Use this in webhook verification to avoid timing side-channels.
func SignatureEqual(expected, actual string) bool {
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}
