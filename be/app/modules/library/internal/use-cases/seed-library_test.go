package usecases

import (
	"encoding/json"
	"io/fs"
	"path"
	"strings"
	"testing"

	"ielts-practice/seeds"
)

// Every committed library file must load: complete cards that pass the AI card checks (Epic 8 §3.4).
func TestCommittedLibrarySeeds(t *testing.T) {
	dirs, err := fs.ReadDir(seeds.Library, "library")
	if err != nil {
		t.Fatal(err)
	}
	topicOf := map[string]string{} // a headword is one card: it may not appear in two topics
	for _, d := range dirs {
		dir := path.Join("library", d.Name())
		raw, err := fs.ReadFile(seeds.Library, path.Join(dir, "collection.json"))
		if err != nil {
			t.Fatal(err)
		}
		var cf collectionFile
		if err := json.Unmarshal(raw, &cf); err != nil {
			t.Fatal(err)
		}
		files, _ := fs.Glob(seeds.Library, dir+"/*.json")
		for _, file := range files {
			topic := strings.TrimSuffix(path.Base(file), ".json")
			if topic == "collection" {
				continue
			}
			raw, _ := fs.ReadFile(seeds.Library, file)
			words, err := ParseTopicFile(raw, topic)
			if err != nil {
				t.Errorf("%s: %v", file, err)
				continue
			}
			for _, w := range words {
				if other, dup := topicOf[w.Key]; dup {
					t.Errorf("%s: %q is also a word of %s", file, w.Card.Headword, other)
				}
				topicOf[w.Key] = topic
			}
			t.Logf("%s: %d words", file, len(words))
		}
	}
}

func TestParseTopicFileRefusesUnsafeCards(t *testing.T) {
	card := `{"headword":"bank","lemma":"bank","ipa":{"uk":"/bæŋk/","us":"/bæŋk/"},"part_of_speech":"noun","cefr":"A2",
		"meaning_vi":"MEANING","word_forms":[{"form":"bank","pos":"noun","ipa":"/bæŋk/","meaning_vi":"ngân hàng",
		"example":{"en":"She works at a bank in the city centre.","vi":"Cô ấy làm ở một ngân hàng."}}],
		"collocations":[{"phrase":"a bank account"},{"phrase":"a bank loan"},{"phrase":"central bank"}]}`
	file := func(meaning string) []byte {
		return []byte(`{"topic":"money","words":[{"source":"core","card":` + strings.Replace(card, "MEANING", meaning, 1) + `}]}`)
	}
	if _, err := ParseTopicFile(file("ngân hàng"), "money"); err != nil {
		t.Fatalf("a valid card: %v", err)
	}
	if _, err := ParseTopicFile(file("xem https://example.com"), "money"); err == nil {
		t.Fatal("a card with a link must be refused")
	}
	if _, err := ParseTopicFile(file("ngân hàng"), "health"); err == nil {
		t.Fatal("the topic must match the file name")
	}
}
