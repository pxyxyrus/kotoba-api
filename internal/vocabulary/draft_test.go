package vocabulary

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateNormalizesWithoutChangingMeaning(t *testing.T) {
	draft, err := Validate(Draft{
		Japanese: "　懐かしい　", Reading: " なつかしい ", MeaningKO: " 그립다 ",
		Context: " この曲、懐かしい。 ", PronunciationConfirmed: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if draft.Japanese != "懐かしい" || draft.Reading != "なつかしい" || draft.MeaningKO != "그립다" || draft.Context != "この曲、懐かしい。" {
		t.Fatalf("unexpected normalization: %+v", draft)
	}
}

func TestValidateRequiresContentAndConfirmation(t *testing.T) {
	_, err := Validate(Draft{Japanese: "　", Reading: " ", MeaningKO: "\n"})
	var fields ValidationError
	if !errors.As(err, &fields) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	for _, field := range []string{"japanese", "reading", "meaning_ko", "pronunciation_confirmed"} {
		if fields[field] == "" {
			t.Errorf("missing error for %s", field)
		}
	}
}

func TestValidateCountsUnicodeRunes(t *testing.T) {
	draft := Draft{Japanese: "語", Reading: "ご", MeaningKO: strings.Repeat("가", 500), PronunciationConfirmed: true}
	if _, err := Validate(draft); err != nil {
		t.Fatalf("500 Korean code points should fit: %v", err)
	}
	draft.MeaningKO += "가"
	var fields ValidationError
	if _, err := Validate(draft); !errors.As(err, &fields) || fields["meaning_ko"] == "" {
		t.Fatalf("expected length error, got %v", err)
	}
}
