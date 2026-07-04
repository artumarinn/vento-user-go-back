package postgres

import (
	"errors"
	"testing"

	"github.com/lib/pq"
)

// Regression: live testing found "cuántos llaveros me quedan?" returned no
// results even though "Llavero generico PLA" (stock 25) exists in the
// catalog — plural "llaveros" doesn't match singular "Llavero", and matching
// was case-sensitive-ish (relied only on ILIKE, no accent folding).
//
// These are pure unit tests over the query-building helpers so the fix is
// verified by `go test ./...` without requiring a live Postgres instance.

func TestStemSearchWord_StripsTrailingS(t *testing.T) {
	cases := map[string]string{
		"llaveros": "llavero",
		"Llavero":  "llavero",
		"PLA":      "pla",
		"pla":      "pla",
		"es":       "e", // trailing 's' stripped like any other word
		"s":        "s", // len==1 guard: stripping would leave an empty stem
	}
	for in, want := range cases {
		if got := stemSearchWord(in); got != want {
			t.Errorf("stemSearchWord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildTextSearchConditions_PluralMatchesSingularViaILIKE(t *testing.T) {
	conditions, args, nextIdx := buildTextSearchConditions("llaveros", false, 2)

	if len(conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d: %v", len(conditions), conditions)
	}
	if len(args) != 1 || args[0] != "%llavero%" {
		t.Fatalf("expected pattern %%llavero%%, got %v", args)
	}
	if nextIdx != 3 {
		t.Fatalf("expected nextIdx 3, got %d", nextIdx)
	}
	want := "(name ILIKE $2 OR description ILIKE $2)"
	if conditions[0] != want {
		t.Fatalf("condition = %q, want %q", conditions[0], want)
	}
}

func TestBuildTextSearchConditions_MultiWordAndsAcrossWords(t *testing.T) {
	conditions, args, nextIdx := buildTextSearchConditions("llavero pla", false, 2)

	if len(conditions) != 2 {
		t.Fatalf("expected 2 conditions (AND across words), got %d: %v", len(conditions), conditions)
	}
	if args[0] != "%llavero%" || args[1] != "%pla%" {
		t.Fatalf("unexpected args: %v", args)
	}
	if nextIdx != 4 {
		t.Fatalf("expected nextIdx 4, got %d", nextIdx)
	}
}

func TestBuildTextSearchConditions_UnaccentAvailable_UsesUnaccentFunction(t *testing.T) {
	conditions, args, _ := buildTextSearchConditions("cancion", true, 5)

	want := "(unaccent(name) ILIKE unaccent($5) OR unaccent(description) ILIKE unaccent($5))"
	if len(conditions) != 1 || conditions[0] != want {
		t.Fatalf("condition = %v, want %q", conditions, want)
	}
	if args[0] != "%cancion%" {
		t.Fatalf("args = %v", args)
	}
}

func TestBuildTextSearchConditions_UnaccentUnavailable_FoldsQueryAccentsManually(t *testing.T) {
	conditions, args, _ := buildTextSearchConditions("canción", false, 2)

	if len(conditions) != 1 {
		t.Fatalf("expected 1 condition, got %v", conditions)
	}
	if args[0] != "%cancion%" {
		t.Fatalf("expected accent-folded pattern %%cancion%%, got %v", args)
	}
}

// Regression: GetProductDetails/GetStock 500'd when the caller passed a
// non-UUID string (e.g. a product NAME the LLM mistook for an id) — Postgres
// rejects it at the type level (22P02) rather than returning zero rows.
// isInvalidUUIDError lets GetByID treat that the same as "not found".
func TestIsInvalidUUIDError(t *testing.T) {
	invalidUUID := &pq.Error{Code: "22P02"}
	if !isInvalidUUIDError(invalidUUID) {
		t.Errorf("expected 22P02 to be recognized as an invalid UUID error")
	}

	otherPQErr := &pq.Error{Code: "23505"} // unique_violation, unrelated
	if isInvalidUUIDError(otherPQErr) {
		t.Errorf("did not expect unrelated pq error code to be treated as invalid UUID")
	}

	if isInvalidUUIDError(errors.New("plain error")) {
		t.Errorf("did not expect a non-pq error to be treated as invalid UUID")
	}

	if isInvalidUUIDError(nil) {
		t.Errorf("did not expect nil error to be treated as invalid UUID")
	}
}

func TestFoldAccents(t *testing.T) {
	cases := map[string]string{
		"canción": "cancion",
		"Ñandú":   "Nandu",
		"café":    "cafe",
		"llavero": "llavero",
	}
	for in, want := range cases {
		if got := foldAccents(in); got != want {
			t.Errorf("foldAccents(%q) = %q, want %q", in, got, want)
		}
	}
}
