// Package vocabulary defines vocabulary rules independently of HTTP and storage.
package vocabulary

import (
	"strings"
	"unicode/utf8"
)

// Draft is a user's proposed Japanese entry and intended Korean meaning.
// Confirmation records the user's assertion; it does not verify pronunciation.
type Draft struct {
	Japanese               string
	Reading                string
	MeaningKO              string
	Context                string
	PronunciationConfirmed bool
}

// ValidationError describes fields that need correction.
type ValidationError map[string]string

func (e ValidationError) Error() string {
	return "word draft needs correction"
}

// Validate trims surrounding whitespace and checks capture requirements.
// It does not verify linguistic accuracy, identify a sense, or save anything.
func Validate(draft Draft) (Draft, error) {
	draft.Japanese = strings.TrimSpace(draft.Japanese)
	draft.Reading = strings.TrimSpace(draft.Reading)
	draft.MeaningKO = strings.TrimSpace(draft.MeaningKO)
	draft.Context = strings.TrimSpace(draft.Context)

	fields := make(ValidationError)
	for _, field := range []struct {
		name     string
		value    string
		maxRunes int
		required bool
	}{
		{"japanese", draft.Japanese, 100, true},
		{"reading", draft.Reading, 200, true},
		{"meaning_ko", draft.MeaningKO, 500, true},
		{"context", draft.Context, 2000, false},
	} {
		switch {
		case field.required && field.value == "":
			fields[field.name] = "This field is required."
		case utf8.RuneCountInString(field.value) > field.maxRunes:
			fields[field.name] = "This field is too long."
		}
	}
	if !draft.PronunciationConfirmed {
		fields["pronunciation_confirmed"] = "Confirm the pronunciation before continuing."
	}
	if len(fields) != 0 {
		return Draft{}, fields
	}
	return draft, nil
}
