package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"kotoba/api/internal/vocabulary"
)

const maxRequestBytes = 16 << 10

type wordRequest struct {
	Japanese               string `json:"japanese" binding:"required"`
	Reading                string `json:"reading" binding:"required"`
	MeaningKO              string `json:"meaning_ko" binding:"required"`
	Context                string `json:"context"`
	PronunciationConfirmed bool   `json:"pronunciation_confirmed"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func writeError(c *gin.Context, status int, code, message string, fields map[string]string) {
	c.AbortWithStatusJSON(status, errorResponse{Error: apiError{
		Code: code, Message: message, Fields: fields,
	}})
}

func validateWord(c *gin.Context) {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Use Content-Type: application/json.", nil)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(c, http.StatusRequestEntityTooLarge, "body_too_large", "The request exceeds 16 KiB.", nil)
			return
		}
		writeError(c, http.StatusBadRequest, "invalid_body", "The request body could not be read.", nil)
		return
	}
	// Gin's JSON binding decodes a value; this also rejects trailing JSON and
	// invalid UTF-8 before the decoder can replace invalid text silently.
	if !utf8.Valid(body) || !json.Valid(body) {
		writeError(c, http.StatusBadRequest, "invalid_json", "Send one valid UTF-8 JSON object.", nil)
		return
	}
	var req wordRequest
	if err := binding.JSON.BindBody(body, &req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_request", "Provide japanese, reading, and meaning_ko as non-empty strings, and a boolean pronunciation_confirmed.", nil)
		return
	}
	draft, err := vocabulary.Validate(vocabulary.Draft{
		Japanese: req.Japanese, Reading: req.Reading, MeaningKO: req.MeaningKO,
		Context: req.Context, PronunciationConfirmed: req.PronunciationConfirmed,
	})
	if err != nil {
		var fields vocabulary.ValidationError
		if errors.As(err, &fields) {
			writeError(c, http.StatusUnprocessableEntity, "validation_failed", "Check the highlighted fields.", fields)
			return
		}
		writeError(c, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
		return
	}
	c.JSON(http.StatusOK, struct {
		Status    string      `json:"status"`
		Persisted bool        `json:"persisted"`
		Word      wordRequest `json:"word"`
	}{
		Status: "validated", Persisted: false,
		Word: wordRequest{
			Japanese: draft.Japanese, Reading: draft.Reading, MeaningKO: draft.MeaningKO,
			Context: draft.Context, PronunciationConfirmed: draft.PronunciationConfirmed,
		},
	})
}
