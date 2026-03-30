package api

import (
	"net/http"
	"testing"
	"time"
)

func TestErrorIsThrottle(t *testing.T) {
	cases := []struct {
		name string
		err  *Error
		want bool
	}{
		{
			name: "http429",
			err:  &Error{StatusCode: 429},
			want: true,
		},
		{
			name: "cloudflare1015",
			err:  &Error{StatusCode: 403, Body: "error code: 1015"},
			want: true,
		},
		{
			name: "capacity",
			err:  &Error{StatusCode: 400, Message: "Not enough capacity to fulfill launch request."},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.IsThrottle(); got != tc.want {
				t.Fatalf("IsThrottle() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	if got := parseRetryAfter("12"); got != 12*time.Second {
		t.Fatalf("parseRetryAfter(seconds) = %v, want 12s", got)
	}

	future := time.Now().Add(20 * time.Second).UTC().Format(http.TimeFormat)
	got := parseRetryAfter(future)
	if got < 18*time.Second || got > 20*time.Second {
		t.Fatalf("parseRetryAfter(date) = %v, want about 20s", got)
	}
}
