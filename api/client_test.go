package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestResponse(statusCode int, body string, headers map[string]string) *http.Response {
	header := make(http.Header)
	for k, v := range headers {
		header.Set(k, v)
	}
	return &http.Response{
		StatusCode: statusCode,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestDoRetriesThrottledResponsesWithJitter(t *testing.T) {
	origSleep := retrySleep
	origJitter := retryJitter
	defer func() {
		retrySleep = origSleep
		retryJitter = origJitter
	}()

	var sleeps []time.Duration
	retrySleep = func(delay time.Duration) {
		sleeps = append(sleeps, delay)
	}
	retryJitter = func(max time.Duration) time.Duration {
		if max != 2500*time.Millisecond && max != 5*time.Second {
			t.Fatalf("unexpected jitter range %v", max)
		}
		return 750 * time.Millisecond
	}

	attempts := 0
	var events []RetryEvent
	client := &Client{
		key: "test-key",
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				if attempts < 3 {
					return newTestResponse(http.StatusTooManyRequests, `{"error":{"message":"rate limited"}}`, nil), nil
				}
				return newTestResponse(http.StatusOK, `{"data":"ok"}`, nil), nil
			}),
		},
	}
	client.SetRetryHook(func(event RetryEvent) {
		events = append(events, event)
	})

	data, err := client.do("GET", "/instances", nil)
	if err != nil {
		t.Fatalf("do() error = %v", err)
	}
	if string(data) != `{"data":"ok"}` {
		t.Fatalf("do() data = %s", data)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
	if len(sleeps) != 2 || sleeps[0] != 5750*time.Millisecond || sleeps[1] != 10750*time.Millisecond {
		t.Fatalf("sleeps = %v, want [5.75s 10.75s]", sleeps)
	}
	if len(events) != 2 {
		t.Fatalf("retry events = %d, want 2", len(events))
	}
	if events[0].Method != "GET" || events[0].Path != "/instances" || events[0].Attempt != 1 {
		t.Fatalf("unexpected first retry event: %+v", events[0])
	}
}

func TestDoHonorsRetryAfterWithoutJitter(t *testing.T) {
	origSleep := retrySleep
	origJitter := retryJitter
	defer func() {
		retrySleep = origSleep
		retryJitter = origJitter
	}()

	var sleeps []time.Duration
	retrySleep = func(delay time.Duration) {
		sleeps = append(sleeps, delay)
	}
	retryJitter = func(time.Duration) time.Duration {
		t.Fatalf("retry jitter should not be used when Retry-After is present")
		return 0
	}

	attempts := 0
	client := &Client{
		key: "test-key",
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				if attempts == 1 {
					return newTestResponse(http.StatusTooManyRequests, `{"error":{"message":"slow down"}}`, map[string]string{"Retry-After": "12"}), nil
				}
				return newTestResponse(http.StatusOK, `{"data":"ok"}`, nil), nil
			}),
		},
	}

	if _, err := client.do("GET", "/instances", nil); err != nil {
		t.Fatalf("do() error = %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if len(sleeps) != 1 || sleeps[0] != 12*time.Second {
		t.Fatalf("sleeps = %v, want [12s]", sleeps)
	}
}

func TestDoDoesNotRetryNonThrottleErrors(t *testing.T) {
	origSleep := retrySleep
	origJitter := retryJitter
	defer func() {
		retrySleep = origSleep
		retryJitter = origJitter
	}()

	calledSleep := false
	retrySleep = func(time.Duration) {
		calledSleep = true
	}
	retryJitter = func(time.Duration) time.Duration { return 0 }

	attempts := 0
	client := &Client{
		key: "test-key",
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				return newTestResponse(http.StatusBadRequest, `{"error":{"message":"Not enough capacity to fulfill launch request."}}`, nil), nil
			}),
		},
	}

	_, err := client.do("POST", "/instance-operations/launch", map[string]string{"name": "test"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
	if calledSleep {
		t.Fatalf("sleep should not be called for non-throttle errors")
	}
}
