package transport

import "net/http"

type RequestAdapter interface {
	Marshal(v interface{}) ([]byte, error)
}

type ResponseAdapter interface {
	Unmarshal(data []byte, v interface{}) error
}

type AuthProvider interface {
	Headers(body []byte, isPayment bool) (http.Header, error)
}
