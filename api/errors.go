package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Error struct {
	StatusCode int
	Message    string
	Body       string
	RetryAfter time.Duration
}

func (e *Error) Error() string {
	switch {
	case e.Message != "":
		return e.Message
	case e.Body != "":
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
	default:
		return fmt.Sprintf("HTTP %d", e.StatusCode)
	}
}

func (e *Error) IsThrottle() bool {
	if e.StatusCode == http.StatusTooManyRequests {
		return true
	}

	text := strings.ToLower(strings.TrimSpace(e.Message + "\n" + e.Body))
	return strings.Contains(text, "error code: 1015") ||
		strings.Contains(text, "rate limit") ||
		strings.Contains(text, "too many requests")
}

func IsThrottleError(err error) bool {
	var apiErr *Error
	return errors.As(err, &apiErr) && apiErr.IsThrottle()
}

func RetryAfterDelay(err error) (time.Duration, bool) {
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.RetryAfter <= 0 {
		return 0, false
	}
	return apiErr.RetryAfter, true
}

func parseRetryAfter(v string) time.Duration {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(v); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	if when, err := http.ParseTime(v); err == nil {
		if delay := time.Until(when); delay > 0 {
			return delay
		}
	}

	return 0
}
