package webhook

import (
	"bytes"
	"encoding/json"
	"strings"

	heleket "github.com/heleket/go-sdk"
)

// Verifier checks Heleket webhook signatures.
//
// Use the API key that matches the webhook type:
//   - payment / wallet → payment API key
//   - payout           → payout API key
//
// Verification recipe (matches Heleket's server, which is PHP-based):
//
//  1. Parse the JSON body preserving the original key ordering and value bytes.
//  2. Extract and remove the "sign" field.
//  3. Concatenate the remaining (key, raw-value) pairs back into a JSON object,
//     preserving PHP-style slash escaping inherited from the raw input.
//  4. Compute md5( base64(concatenated) . apiKey ).
//  5. Constant-time compare against the received sign.
//
// IMPORTANT: PHP's json_encode preserves insertion order and escapes forward
// slashes by default. Go's encoding/json sorts map keys and does NOT escape
// slashes. The verifier therefore CANNOT round-trip the payload through a
// Go map[string]any — that would change both the order and the escaping. Use
// VerifyRaw with the original request bytes.
type Verifier struct {
	apiKey string
}

// NewVerifier returns a Verifier configured with the given API key.
func NewVerifier(apiKey string) *Verifier {
	return &Verifier{apiKey: apiKey}
}

// VerifyRaw checks the signature of a raw JSON body, preserving the original
// key ordering and per-value byte representation. This is the only correct
// entry point against real Heleket webhooks (which are PHP-encoded with
// `\/` slash escapes and insertion-order keys).
//
// Returns:
//   - *SignatureError when the signature is missing, malformed, or mismatched
//   - *PayloadDecodeError when the signature was valid but the typed Payload
//     could not be decoded (distinguishes "Heleket lied" from "we couldn't parse")
func (v *Verifier) VerifyRaw(rawBody []byte) (*Payload, error) {
	sign, withoutSign, ordered, err := stripSignField(rawBody)
	if err != nil {
		return nil, err
	}

	expected := heleket.Sign(withoutSign, v.apiKey)
	if !heleket.SignatureEqual(expected, sign) {
		return nil, &heleket.SignatureError{Reason: "signature mismatch"}
	}

	// For the typed Payload we decode normally — Go's loss of key order doesn't
	// matter here because we've already verified the signature on the raw bytes.
	var p Payload
	if err := json.Unmarshal(rawBody, &p); err != nil {
		return nil, &heleket.PayloadDecodeError{Cause: err}
	}
	p.Sign = sign
	p.Raw = ordered
	return &p, nil
}

// stripSignField walks the input body in order and returns:
//   - the sign value;
//   - the JSON body with the sign field removed, preserving key order and the
//     original byte representation of each value (so PHP-style slash escapes
//     survive);
//   - a flat map of every original field (including sign) for convenience.
//
// Returns *heleket.SignatureError on malformed input or missing sign.
func stripSignField(body []byte) (string, []byte, map[string]any, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	tok, err := dec.Token()
	if err != nil {
		return "", nil, nil, &heleket.SignatureError{Reason: "body is not valid JSON: " + err.Error()}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return "", nil, nil, &heleket.SignatureError{Reason: "webhook body must be a JSON object"}
	}

	var out strings.Builder
	out.WriteByte('{')

	rawMap := make(map[string]any)
	var sign string
	signSeen := false
	first := true

	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return "", nil, nil, &heleket.SignatureError{Reason: "decode key: " + err.Error()}
		}
		key, ok := keyTok.(string)
		if !ok {
			return "", nil, nil, &heleket.SignatureError{Reason: "non-string key"}
		}

		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return "", nil, nil, &heleket.SignatureError{Reason: "decode value: " + err.Error()}
		}

		// Populate the convenience map (decoded values).
		var decoded any
		_ = json.Unmarshal(raw, &decoded)
		rawMap[key] = decoded

		if key == "sign" {
			var s string
			if err := json.Unmarshal(raw, &s); err != nil {
				return "", nil, nil, &heleket.SignatureError{Reason: "sign field must be a string: " + err.Error()}
			}
			sign = s
			signSeen = true
			continue
		}

		if !first {
			out.WriteByte(',')
		}
		first = false

		// Encode the key as JSON (PHP and Go agree on string-key encoding for
		// the ASCII identifiers Heleket uses).
		keyBytes, _ := json.Marshal(key)
		out.Write(keyBytes)
		out.WriteByte(':')
		out.Write(raw)
	}

	out.WriteByte('}')

	if !signSeen {
		return "", nil, nil, &heleket.SignatureError{Reason: "missing sign field"}
	}
	if sign == "" {
		return "", nil, nil, &heleket.SignatureError{Reason: "sign field must be a non-empty string"}
	}

	return sign, []byte(out.String()), rawMap, nil
}
