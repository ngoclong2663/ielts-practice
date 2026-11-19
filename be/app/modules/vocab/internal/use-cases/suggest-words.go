package usecases

import (
	"context"
	"net/http"
	"strings"

	"ielts-practice/app/modules/vocab/domain/rules"

	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/dto"
	"ielts-practice/pkg/exceptions"
)

const maxPromptExclusions = 300

func (uc *VocabUseCases) ListTopics() dto.TopicListDto {
	out := dto.TopicListDto{Items: make([]dto.TopicDto, 0, len(rules.Topics))}
	for _, t := range rules.Topics {
		out.Items = append(out.Items, dto.TopicDto{ID: t.ID})
	}
	return out
}

// suggestionQuota: AI word-list requests used in the last 24 h (Epic 4: 5 per account by default).
func (uc *VocabUseCases) suggestionQuota(ctx context.Context, userID string) (limit, used int, err error) {
	used, err = uc.repo.SuggestionRequestsSince(ctx, userID, uc.now().Add(-quotaWindow))
	return uc.cfg.Vocab.DailySuggestions, used, err
}

func (uc *VocabUseCases) checkSuggestionQuota(ctx context.Context, userID string) (int, error) {
	if !uc.suggester.Configured() {
		return 0, appErr(http.StatusServiceUnavailable, exceptions.CodeVocabNotConfigured, "Word suggestions are not set up")
	}
	limit, used, err := uc.suggestionQuota(ctx, userID)
	if err != nil {
		return 0, internalErr("Failed to load suggestion limit", err)
	}
	if used >= limit {
		return 0, appErr(http.StatusTooManyRequests, exceptions.CodeSuggestionLimit, "Daily suggestion limit reached").
			WithParams(map[string]any{"limit": limit})
	}
	return limit - used, nil
}

// CreateSuggestion asks the model for `count` new words for a topic (Epic 4 §4).
func (uc *VocabUseCases) CreateSuggestion(ctx context.Context, userID string, req dto.CreateSuggestionRequestDto) (dto.SuggestionDto, error) {
	// A preset wins over typed text; an unknown preset id is ignored.
	text := strings.Join(strings.Fields(req.TopicText), " ")
	if _, ok := rules.FindTopic(req.TopicID); ok {
		text = ""
	} else {
		req.TopicID = ""
	}
	if req.TopicID == "" && text == "" {
		e := appErr(http.StatusBadRequest, exceptions.CodeTopicRequired, "Choose or type a topic")
		e.Fields = map[string]string{"topic_text": "VALIDATION_REQUIRED"}
		return dto.SuggestionDto{}, e
	}
	remaining, err := uc.checkSuggestionQuota(ctx, userID)
	if err != nil {
		return dto.SuggestionDto{}, err
	}
	s := &entity.Suggestion{
		UserID: userID, DeckID: req.DeckID, TopicID: req.TopicID, TopicText: text,
		Requested: req.Count, Words: []entity.SuggestedWord{},
	}
	added, err := uc.suggestInto(ctx, s, req.Count)
	if err != nil {
		return dto.SuggestionDto{}, err
	}
	created, err := uc.repo.CreateSuggestion(ctx, s)
	if err != nil {
		return dto.SuggestionDto{}, internalErr("Failed to save suggestions", err)
	}
	return uc.finishRequest(ctx, created, added.words, added.usage, remaining)
}

// MoreSuggestions appends `count` more words, never repeating a word already in the list.
func (uc *VocabUseCases) MoreSuggestions(ctx context.Context, userID, id string, req dto.MoreSuggestionsRequestDto) (dto.SuggestionDto, error) {
	s, err := uc.repo.FindSuggestion(ctx, userID, id)
	if err != nil {
		return dto.SuggestionDto{}, internalErr("Failed to load suggestions", err)
	}
	if s == nil {
		return dto.SuggestionDto{}, appErr(http.StatusNotFound, exceptions.CodeSuggestionNotFound, "Word list not found")
	}
	remaining, err := uc.checkSuggestionQuota(ctx, userID)
	if err != nil {
		return dto.SuggestionDto{}, err
	}
	added, err := uc.suggestInto(ctx, s, req.Count)
	if err != nil {
		return dto.SuggestionDto{}, err
	}
	if err := uc.repo.UpdateSuggestionWords(ctx, s); err != nil {
		return dto.SuggestionDto{}, internalErr("Failed to save suggestions", err)
	}
	return uc.finishRequest(ctx, s, added.words, added.usage, remaining)
}

func (uc *VocabUseCases) GetSuggestion(ctx context.Context, userID, id string) (dto.SuggestionDto, error) {
	s, err := uc.repo.FindSuggestion(ctx, userID, id)
	if err != nil {
		return dto.SuggestionDto{}, internalErr("Failed to load suggestions", err)
	}
	if s == nil {
		return dto.SuggestionDto{}, appErr(http.StatusNotFound, exceptions.CodeSuggestionNotFound, "Word list not found")
	}
	limit, used, err := uc.suggestionQuota(ctx, userID)
	if err != nil {
		return dto.SuggestionDto{}, internalErr("Failed to load suggestion limit", err)
	}
	return uc.suggestionDto(ctx, s, 0, max(0, limit-used))
}

// finishRequest counts a request toward the daily limit only when it added words.
func (uc *VocabUseCases) finishRequest(ctx context.Context, s *entity.Suggestion, added int, usage entity.Usage, remaining int) (dto.SuggestionDto, error) {
	if added > 0 {
		if err := uc.repo.LogSuggestionRequest(ctx, s.UserID, s.ID, added, usage); err != nil {
			return dto.SuggestionDto{}, internalErr("Failed to record suggestion request", err)
		}
		remaining--
	}
	return uc.suggestionDto(ctx, s, added, max(0, remaining))
}

// suggestionDto hides words the user has marked as known since the list was made.
func (uc *VocabUseCases) suggestionDto(ctx context.Context, s *entity.Suggestion, added, remaining int) (dto.SuggestionDto, error) {
	known, err := uc.repo.KnownWords(ctx, s.UserID)
	if err != nil {
		return dto.SuggestionDto{}, internalErr("Failed to load known words", err)
	}
	knownSet := toSet(known)
	words := make([]entity.SuggestedWord, 0, len(s.Words))
	for _, w := range s.Words {
		if !knownSet[strings.ToLower(w.Lemma)] {
			words = append(words, w)
		}
	}
	return dto.SuggestionDto{
		ID: s.ID, TopicID: s.TopicID, TopicText: s.TopicText, DeckID: s.DeckID,
		Requested: s.Requested, Words: words, SkippedExisting: s.SkippedExisting, Added: added,
		CreatedAt: s.CreatedAt, Remaining: remaining,
	}, nil
}

type addedWords struct {
	words int
	usage entity.Usage
}

// suggestInto asks the model for extra candidates, then removes on the server every word that is
// already in the chosen deck, known, or already in this list (Epic 4 §6.3), keeping at most `count`.
func (uc *VocabUseCases) suggestInto(ctx context.Context, s *entity.Suggestion, count int) (addedWords, error) {
	var deckWords []string
	if s.DeckID != nil {
		words, err := uc.decks.DeckWords(ctx, s.UserID, *s.DeckID)
		if err != nil {
			return addedWords{}, err
		}
		deckWords = words
	}
	known, err := uc.repo.KnownWords(ctx, s.UserID)
	if err != nil {
		return addedWords{}, internalErr("Failed to load known words", err)
	}
	listed := make([]string, 0, len(s.Words))
	for _, w := range s.Words {
		listed = append(listed, strings.ToLower(w.Lemma))
	}
	// A preset topic the library covers (Epic 8): the learner adds those words from the library, so the AI
	// suggests the words the library does not have.
	var libraryWords []string
	if s.TopicID != "" {
		if libraryWords, err = uc.library.TopicWords(ctx, s.TopicID); err != nil {
			return addedWords{}, internalErr("Failed to load the library words", err)
		}
	}
	inDeck := toSet(deckWords)
	excluded := toSet(append(append(append(listed, known...), deckWords...), libraryWords...))

	// The prompt gets the most relevant exclusions (this list, the deck, the library, then known words), capped.
	promptExclude := append(append(append(append([]string{}, listed...), deckWords...), libraryWords...), known...)
	if len(promptExclude) > maxPromptExclusions {
		promptExclude = promptExclude[:maxPromptExclusions]
	}
	topic := s.TopicText
	if t, ok := rules.FindTopic(s.TopicID); ok {
		topic = t.Description
	}
	candidates, usage, err := uc.suggester.Suggest(ctx, entity.SuggestRequest{
		Topic: topic, Count: min(45, count+count/2+2), Exclude: promptExclude,
	})
	uc.logger.Info().Str("topic", topic).Int("candidates", len(candidates)).Int("input_tokens", usage.InputTokens).
		Int("output_tokens", usage.OutputTokens).Err(err).Msg("Suggested topic words")
	if err != nil {
		return addedWords{}, appErr(http.StatusBadGateway, exceptions.CodeSuggestionFailed, "Could not suggest words right now")
	}

	added := 0
	for _, c := range candidates {
		if added == count {
			break
		}
		entries, _ := rules.Normalize(c.Lemma)
		if len(entries) != 1 || entries[0].SenseHint != "" {
			continue
		}
		lemma := strings.ToLower(entries[0].Text)
		if excluded[lemma] {
			if inDeck[lemma] {
				s.SkippedExisting++
			}
			continue
		}
		excluded[lemma] = true
		c.Lemma = entries[0].Text
		c.POS, c.MeaningVI = strings.TrimSpace(c.POS), strings.TrimSpace(c.MeaningVI)
		s.Words = append(s.Words, c)
		added++
	}
	return addedWords{words: added, usage: usage}, nil
}

func (uc *VocabUseCases) MarkKnown(ctx context.Context, userID string, req dto.KnownWordRequestDto) error {
	lemma := strings.ToLower(strings.TrimSpace(req.Lemma))
	if lemma == "" {
		return appErr(http.StatusBadRequest, exceptions.CodeValidationFailed, "Word is required")
	}
	if err := uc.repo.AddKnownWord(ctx, userID, lemma); err != nil {
		return internalErr("Failed to save known word", err)
	}
	return nil
}

func (uc *VocabUseCases) UnmarkKnown(ctx context.Context, userID, lemma string) error {
	if err := uc.repo.RemoveKnownWord(ctx, userID, strings.ToLower(strings.TrimSpace(lemma))); err != nil {
		return internalErr("Failed to remove known word", err)
	}
	return nil
}

func toSet(words []string) map[string]bool {
	out := make(map[string]bool, len(words))
	for _, w := range words {
		out[strings.ToLower(strings.TrimSpace(w))] = true
	}
	return out
}
