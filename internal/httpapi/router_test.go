package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const validWord = `{"japanese":" 懐かしい ","reading":"なつかしい","meaning_ko":"그립다","pronunciation_confirmed":true}`

func TestValidateWord(t *testing.T) {
	router := NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, tc := range []struct {
		name, body, contentType string
		status                  int
		code                    string
	}{
		{"valid", validWord, "application/json; charset=utf-8", 200, ""},
		{"malformed", `{`, "application/json", 400, "invalid_json"},
		{"trailing JSON", validWord + `{}`, "application/json", 400, "invalid_json"},
		{"null", `null`, "application/json", 400, "invalid_request"},
		{"wrong type", `{"japanese":42}`, "application/json", 400, "invalid_request"},
		{"missing fields", `{}`, "application/json", 400, "invalid_request"},
		{"whitespace", strings.Replace(validWord, "그립다", "　", 1), "application/json", 422, "validation_failed"},
		{"unconfirmed", strings.Replace(validWord, "true", "false", 1), "application/json", 422, "validation_failed"},
		{"too large", strings.Repeat(" ", maxRequestBytes+1), "application/json", 413, "body_too_large"},
		{"invalid UTF-8", "{\"japanese\":\"\xff\"}", "application/json", 400, "invalid_json"},
		{"wrong content type", validWord, "text/plain", 415, "unsupported_media_type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/words/validate", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tc.status {
				t.Fatalf("got status %d, want %d: %s", w.Code, tc.status, w.Body)
			}
			if tc.code != "" {
				var response errorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if response.Error.Code != tc.code {
					t.Fatalf("got code %q, want %q", response.Error.Code, tc.code)
				}
				return
			}
			var response struct {
				Status    string      `json:"status"`
				Persisted *bool       `json:"persisted"`
				Word      wordRequest `json:"word"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Status != "validated" || response.Persisted == nil || *response.Persisted || response.Word.Japanese != "懐かしい" || response.Word.MeaningKO != "그립다" {
				t.Fatalf("unexpected response: %s", w.Body)
			}
		})
	}
}

func TestRoutesAndRecovery(t *testing.T) {
	var logs bytes.Buffer
	router := NewRouter(slog.New(slog.NewJSONHandler(&logs, nil)))
	router.GET("/test-panic", func(c *gin.Context) { panic("private details") })
	for _, tc := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/healthz", 200},
		{http.MethodGet, "/missing?meaning=private", 404},
		{http.MethodGet, "/api/v1/words/validate", 405},
		{http.MethodGet, "/test-panic", 500},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status || !json.Valid(w.Body.Bytes()) {
			t.Errorf("%s %s: got %d: %s", tc.method, tc.path, w.Code, w.Body)
		}
		if strings.Contains(w.Body.String(), "private") {
			t.Error("response leaked private content")
		}
	}
	if strings.Contains(logs.String(), "private") {
		t.Error("logs leaked query or panic content")
	}
}
