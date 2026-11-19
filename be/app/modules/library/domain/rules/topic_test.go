package rules

import (
	"testing"

	cards "ielts-practice/app/modules/decks/domain/entity"
	"ielts-practice/app/modules/library/domain/entity"
)

func TestSuggestTopic(t *testing.T) {
	topics := []string{"environment", "health", "crime"}
	vocab := map[string]string{
		"environment": "|pollution|emission|climate change|toxic|waste|the environment: pollution, climate change, energy, conservation, waste",
		"health":      "|disease|symptom|hospital|treatment|health and hospitals: illness, treatment, hospitals",
		"crime":       "|prison|offender|police|crime and law: crime, punishment, police, courts, justice",
	}
	card := &entity.CardContent{
		Headword: "exacerbate", DefinitionEN: "to make a problem or bad situation worse",
		WordForms:    []cards.WordForm{{Form: "exacerbate", Example: &cards.Example{EN: "Traffic emissions exacerbate air pollution in big cities."}}},
		Collocations: []cards.Collocation{{Phrase: "exacerbate pollution"}, {Phrase: "exacerbate climate change"}},
	}
	if got := SuggestTopic(card, topics, vocab); got != "environment" {
		t.Fatalf("got %q, want environment", got)
	}
	plain := &entity.CardContent{Headword: "whereas", DefinitionEN: "used to compare two facts"}
	if got := SuggestTopic(plain, topics, vocab); got != entity.OtherTopic {
		t.Fatalf("no shared words: got %q, want other", got)
	}
}
