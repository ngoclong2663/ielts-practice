package usecases

import (
	"context"

	libraryin "ielts-practice/app/modules/library/ports/in"
)

var _ libraryin.ITopicWords = &LibraryUseCases{}

// TopicWords lists the library words of a topic, for the topic suggestions (Epic 4) to skip.
func (uc *LibraryUseCases) TopicWords(ctx context.Context, topicID string) ([]string, error) {
	return uc.repo.TopicHeadwords(ctx, topicID)
}
