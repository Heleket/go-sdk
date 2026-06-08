// Package testutil provides test doubles for the Heleket SDK. The package is
// importable inside this module (and by external test code that vendors the
// SDK) but is not part of the SDK's stable public API.
package testutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	heleket "github.com/heleket/go-sdk"
)

// RecordedRequest captures everything a client sent in a single call.
type RecordedRequest struct {
	Method  string
	URL     string
	Headers http.Header
	Body    []byte
}

// FakeTransport implements heleket.Transport in memory. Enqueue canned
// responses (or transport failures) and inspect what the client sent via
// Requests() / LastRequest().
type FakeTransport struct {
	mu        sync.Mutex
	queue     []*heleket.Response
	failures  []error
	recorded  []RecordedRequest
	failModes []bool // parallel to queue: true means "fail this call"
}

// NewFakeTransport returns an empty FakeTransport.
func NewFakeTransport() *FakeTransport {
	return &FakeTransport{}
}

// Enqueue queues a verbatim response for the next call.
func (f *FakeTransport) Enqueue(resp *heleket.Response) *FakeTransport {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.queue = append(f.queue, resp)
	f.failures = append(f.failures, nil)
	f.failModes = append(f.failModes, false)
	return f
}

// EnqueueJSON marshals payload and queues it as a JSON response with the given
// status code.
func (f *FakeTransport) EnqueueJSON(payload any, statusCode int) *FakeTransport {
	body, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("FakeTransport.EnqueueJSON marshal failed: %v", err))
	}
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	return f.Enqueue(&heleket.Response{StatusCode: statusCode, Body: body, Header: header})
}

// FailNext queues a transport-level failure for the next call. The error
// surfaces wrapped in heleket.HTTPError.
func (f *FakeTransport) FailNext(err error) *FakeTransport {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err == nil {
		err = errors.New("simulated transport failure")
	}
	f.queue = append(f.queue, nil)
	f.failures = append(f.failures, err)
	f.failModes = append(f.failModes, true)
	return f
}

// RoundTrip implements heleket.Transport.
func (f *FakeTransport) RoundTrip(ctx context.Context, method, url string, headers http.Header, body []byte) (*heleket.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Copy the headers so later test mutations don't change recorded data.
	headerCopy := make(http.Header, len(headers))
	for k, v := range headers {
		headerCopy[k] = append([]string(nil), v...)
	}
	bodyCopy := append([]byte(nil), body...)

	f.recorded = append(f.recorded, RecordedRequest{
		Method:  method,
		URL:     url,
		Headers: headerCopy,
		Body:    bodyCopy,
	})

	if len(f.queue) == 0 {
		return nil, fmt.Errorf("FakeTransport: no enqueued response for %s %s", method, url)
	}
	resp := f.queue[0]
	failure := f.failures[0]
	failMode := f.failModes[0]
	f.queue = f.queue[1:]
	f.failures = f.failures[1:]
	f.failModes = f.failModes[1:]
	if failMode {
		return nil, failure
	}
	return resp, nil
}

// Requests returns all recorded calls in order.
func (f *FakeTransport) Requests() []RecordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]RecordedRequest, len(f.recorded))
	copy(out, f.recorded)
	return out
}

// LastRequest returns the most recent recorded call. Panics if none were made.
func (f *FakeTransport) LastRequest() RecordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.recorded) == 0 {
		panic("FakeTransport: no requests recorded yet")
	}
	return f.recorded[len(f.recorded)-1]
}
