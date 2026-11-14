package steps

import "testing"

func TestParseMinutes(t *testing.T) {
	got, err := ParseMinutes("1m 10m 1h 1d")
	if err != nil || len(got) != 4 || got[0] != 1 || got[1] != 10 || got[2] != 60 || got[3] != 1440 {
		t.Fatalf("got %v %v", got, err)
	}
	if got, err := ParseMinutes(""); err != nil || len(got) != 0 {
		t.Fatalf("empty: %v %v", got, err)
	}
	for _, bad := range []string{"10", "5x", "-1m", "0m", "1m 2m 3m 4m 5m 6m 7m", "m"} {
		if _, err := ParseMinutes(bad); err == nil {
			t.Errorf("%q should be invalid", bad)
		}
	}
}
