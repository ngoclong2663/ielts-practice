package rules

import "testing"

func TestKeyForFile(t *testing.T) {
	if got := KeyForFile("ab.mp4"); got != "guides/videos/ab.mp4" {
		t.Errorf("video: %s", got)
	}
	if got := KeyForFile("ab.jpg"); got != "guides/posters/ab.jpg" {
		t.Errorf("poster: %s", got)
	}
}
