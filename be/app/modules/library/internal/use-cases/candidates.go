package usecases

import (
	"context"
	"net/http"
	"slices"

	"ielts-practice/app/modules/library/domain/entity"
	"ielts-practice/app/modules/library/domain/rules"
	"ielts-practice/app/modules/library/dto"
	vocabrules "ielts-practice/app/modules/vocab/domain/rules"
	"ielts-practice/pkg/exceptions"
)

// Library candidates (Epic 9): AI-made cards several learners keep, approved into the library by the owner.

const maxCandidates = 200

// CandidateTopics are the topics a word may be approved into: the Epic 4 presets, then "other".
func CandidateTopics() []string { return vocabrules.TopicIDs() }

func (uc *LibraryUseCases) ListCandidates(ctx context.Context) (dto.CandidatesResponseDto, error) {
	minUsers := uc.cfg.Library.CandidateMinUsers
	items, err := uc.repo.ListCandidates(ctx, minUsers, maxCandidates)
	if err != nil {
		return dto.CandidatesResponseDto{}, internalErr("Failed to load the library candidates", err)
	}
	cols, err := uc.repo.ListCollections(ctx)
	if err != nil {
		return dto.CandidatesResponseDto{}, internalErr("Failed to load the library", err)
	}
	out := dto.CandidatesResponseDto{MinUsers: minUsers, Topics: CandidateTopics(),
		Collections: make([]dto.CandidateCollectionDto, 0, len(cols)), Items: make([]dto.CandidateDto, 0, len(items))}
	for _, c := range cols {
		topics := make([]string, 0, len(c.Topics))
		for _, t := range c.Topics {
			topics = append(topics, t.TopicID)
		}
		out.Collections = append(out.Collections, dto.CandidateCollectionDto{
			Slug: c.Slug, Title: dto.LocalizedDto{VI: c.Title.VI, EN: c.Title.EN}, Topics: topics})
	}
	vocab, err := uc.topicVocab(ctx)
	if err != nil {
		return dto.CandidatesResponseDto{}, internalErr("Failed to load the library topics", err)
	}
	presets := out.Topics[:len(out.Topics)-1] // without "other"
	for i := range items {
		c := &items[i]
		// The topic the generator tagged; cards made before it did get a guess from shared words.
		topic := c.Topic
		if !slices.Contains(out.Topics, topic) {
			topic = rules.SuggestTopic(&c.Card, presets, vocab)
		}
		out.Items = append(out.Items, dto.CandidateDto{
			LexemeID: c.LexemeID, Card: c.Card, Users: c.Users, ReportCount: c.ReportCount, LastAddedAt: c.LastAddedAt,
			SuggestedTopic: topic})
	}
	return out, nil
}

// ApproveCandidates adds the words to their topics in the collection (the first one when none is given).
// Ids that are not undecided AI-made cards are skipped.
func (uc *LibraryUseCases) ApproveCandidates(ctx context.Context, req dto.ApproveCandidatesRequestDto) (dto.DecisionResponseDto, error) {
	cols, err := uc.repo.ListCollections(ctx)
	if err != nil {
		return dto.DecisionResponseDto{}, internalErr("Failed to load the library", err)
	}
	var col *entity.Collection
	for i := range cols {
		if req.Collection == "" || cols[i].Slug == req.Collection {
			col = &cols[i]
			break
		}
	}
	if col == nil {
		return dto.DecisionResponseDto{}, notFound()
	}
	topics := CandidateTopics()
	ids := make([]string, 0, len(req.Items))
	for _, it := range req.Items {
		if !slices.Contains(topics, it.TopicID) {
			e := appErr(http.StatusBadRequest, exceptions.CodeValidationFailed, "Unknown topic")
			e.Fields = map[string]string{"topic_id": "VALIDATION_ONEOF"}
			return dto.DecisionResponseDto{}, e
		}
		ids = append(ids, it.LexemeID)
	}
	cards, err := uc.repo.FindCandidateCards(ctx, ids)
	if err != nil {
		return dto.DecisionResponseDto{}, internalErr("Failed to load the cards", err)
	}
	approvals := make([]entity.Approval, 0, len(cards))
	for _, it := range req.Items {
		card, ok := cards[it.LexemeID]
		if !ok {
			continue
		}
		cefr := card.CEFR
		if !slices.Contains(entity.Levels, cefr) {
			cefr = ""
		}
		approvals = append(approvals, entity.Approval{LexemeID: it.LexemeID, TopicID: it.TopicID,
			Word: entity.SeedWord{Card: card, CEFR: cefr, SearchText: rules.SearchText(&card)}})
		delete(cards, it.LexemeID) // an id listed twice is approved once
	}
	if err := uc.repo.ApproveCandidates(ctx, col.ID, approvals); err != nil {
		return dto.DecisionResponseDto{}, internalErr("Failed to approve the candidates", err)
	}
	uc.logger.Info().Int("approved", len(approvals)).Str("collection", col.Slug).Msg("Library candidates approved")
	return dto.DecisionResponseDto{Count: len(approvals)}, nil
}

func (uc *LibraryUseCases) RejectCandidates(ctx context.Context, req dto.RejectCandidatesRequestDto) (dto.DecisionResponseDto, error) {
	n, err := uc.repo.RejectCandidates(ctx, req.LexemeIDs)
	if err != nil {
		return dto.DecisionResponseDto{}, internalErr("Failed to reject the candidates", err)
	}
	return dto.DecisionResponseDto{Count: n}, nil
}

// topicVocab is each preset topic's vocabulary: its description (the words the AI sees for the topic) and the
// words of its library topic, when the library has one.
func (uc *LibraryUseCases) topicVocab(ctx context.Context) (map[string]string, error) {
	vocab, err := uc.repo.TopicVocab(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range vocabrules.Topics {
		vocab[t.ID] += " " + t.Description
	}
	return vocab, nil
}
