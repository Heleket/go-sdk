package transport

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/http"
)

const (
	MerchantHeader = "merchant"
	Header         = "sign"
)

type APIKeyAuthProvider struct {
	paymentKey string
	merchantId string
	payoffKey  string
}

func NewAPIKeyAuth(paymentKey, payoffKey, merchantId string) (*APIKeyAuthProvider, error) {
	if merchantId == "" || paymentKey == "" || payoffKey == "" {
		return nil, fmt.Errorf("required parameters are missing")
	}

	return &APIKeyAuthProvider{
		paymentKey: paymentKey,
		merchantId: merchantId,
		payoffKey:  payoffKey,
	}, nil
}

func (a *APIKeyAuthProvider) Headers(body []byte, isPayment bool) (http.Header, error) {
	if a.paymentKey == "" {
		return nil, fmt.Errorf("api paymentKey is empty")
	}
	h := http.Header{}

	// Кодируем body в base64
	encoded := base64.StdEncoding.EncodeToString(body)

	key := a.paymentKey

	if !isPayment {
		key = a.payoffKey
	}

	// Создаем подпись: md5(base64(data) + apiKey)
	signString := encoded + key
	hash := md5.Sum([]byte(signString))
	sign := fmt.Sprintf("%x", hash)

	h.Set(Header, sign)
	h.Set(MerchantHeader, a.merchantId)
	return h, nil
}

type BearerAuthProvider struct{ token string }

func NewBearerAuth(token string) *BearerAuthProvider { return &BearerAuthProvider{token: token} }

func (b *BearerAuthProvider) Headers(method string, path string, body []byte) (http.Header, error) {
	if b.token == "" {
		return nil, fmt.Errorf("bearer token is empty")
	}
	h := http.Header{}
	h.Set("Authorization", "Bearer "+b.token)
	return h, nil
}
