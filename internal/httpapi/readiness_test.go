package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadiness(t *testing.T) {
	for _, tc := range []struct {
		name   string
		check  func(context.Context) error
		status int
	}{
		{"unconfigured", nil, 503},
		{"available", func(ctx context.Context) error {
			if _, ok := ctx.Deadline(); !ok {
				t.Error("missing deadline")
			}
			return nil
		}, 200},
		{"unavailable", func(context.Context) error { return errors.New("private-driver-detail") }, 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			router := NewRouterWithReadiness(slog.New(slog.NewTextHandler(io.Discard, nil)), tc.check)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", "/readyz", nil))
			if response.Code != tc.status {
				t.Fatalf("got %d want %d", response.Code, tc.status)
			}
			if strings.Contains(response.Body.String(), "private-driver-detail") {
				t.Fatal("leaked driver detail")
			}
		})
	}
}
