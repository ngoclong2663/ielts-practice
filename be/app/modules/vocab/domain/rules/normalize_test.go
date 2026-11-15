package rules

import (
	"testing"

	"ielts-practice/app/modules/vocab/domain/entity"
)

func TestNormalize(t *testing.T) {
	entries, rejected := Normalize("Decide, take into account\n1. carbon footprint; bank | river\nbank (river)\n- UNESCO\nthis is a very long sentence that is not a word\ndecide\nhello123")
	want := []entity.Entry{
		{Text: "decide"}, {Text: "take into account"}, {Text: "carbon footprint"},
		{Text: "bank", SenseHint: "river"}, {Text: "UNESCO"},
	}
	if len(entries) != len(want) {
		t.Fatalf("entries = %+v", entries)
	}
	for i := range want {
		if entries[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, entries[i], want[i])
		}
	}
	reasons := map[string]string{}
	for _, r := range rejected {
		reasons[r.Text] = r.Reason
	}
	if reasons["bank (river)"] != entity.RejectDuplicate || reasons["decide"] != entity.RejectDuplicate {
		t.Errorf("duplicates not rejected: %+v", rejected)
	}
	if reasons["this is a very long sentence that is not a word"] != entity.RejectTooManyWords {
		t.Errorf("sentence not rejected: %+v", rejected)
	}
	if reasons["hello123"] != entity.RejectInvalidChars {
		t.Errorf("digits not rejected: %+v", rejected)
	}
}

func TestLexemeKey(t *testing.T) {
	if LexemeKey(" Bank ", "River") != "bank|river|vi" {
		t.Fatal(LexemeKey(" Bank ", "River"))
	}
}
