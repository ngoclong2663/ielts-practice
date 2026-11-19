package domain

import (
	"context"

	"ielts-practice/app/modules/library/domain/entity"
)

type ILibraryRepository interface {
	// ListCollections returns published collections with their topics (in order) and word counts.
	ListCollections(ctx context.Context) ([]entity.Collection, error)
	FindCollection(ctx context.Context, slug string) (*entity.Collection, error)
	// ListWords returns a collection's words with the learner's flags, in topic then word order.
	ListWords(ctx context.Context, userID, collectionID string, f entity.WordFilter) ([]entity.Word, error)
	// TopicHeadwords returns the headwords of a topic's words in every published collection, lower case.
	TopicHeadwords(ctx context.Context, topicID string) ([]string, error)
	// FindCards returns the ready cards of library lexemes, keyed by lexeme id; other ids are left out.
	FindCards(ctx context.Context, lexemeIDs []string) (map[string]entity.CardContent, error)

	// Candidates (Epic 9)
	// ListCandidates returns AI-made cards kept by at least minUsers learners, with no decision yet.
	ListCandidates(ctx context.Context, minUsers, limit int) ([]entity.Candidate, error)
	// TopicVocab returns, per topic, the search text of its library words (for suggesting a candidate's topic).
	TopicVocab(ctx context.Context) (map[string]string, error)
	// FindCandidateCards returns the cards of the given ids that are still undecided AI-made cards.
	FindCandidateCards(ctx context.Context, lexemeIDs []string) (map[string]entity.CardContent, error)
	// ApproveCandidates adds the words to their topics (creating a topic when needed) and locks their cards.
	ApproveCandidates(ctx context.Context, collectionID string, items []entity.Approval) error
	// RejectCandidates records the decision; it returns how many undecided cards it covered.
	RejectCandidates(ctx context.Context, lexemeIDs []string) (int, error)

	// Seed (Epic 8 §3.4)
	UpsertCollection(ctx context.Context, c *entity.Collection) error
	TopicHash(ctx context.Context, collectionID, topicID string) (string, error)
	// ReplaceTopic stores a topic and its words; words no longer in the file leave the topic.
	ReplaceTopic(ctx context.Context, collectionID, topicID string, sort int, hash string, words []entity.SeedWord) error
	// UpsertLibraryLexeme stores a reviewed card in the shared cache (origin 'library') and returns its id.
	UpsertLibraryLexeme(ctx context.Context, w entity.SeedWord) (string, error)
	// FindLibraryLexeme returns the id of the library lexeme with this cache key ("" when there is none);
	// with wordsOnly, only one that is a word of a library topic.
	FindLibraryLexeme(ctx context.Context, key string, wordsOnly bool) (string, error)
}
