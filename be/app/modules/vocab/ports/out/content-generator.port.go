// Package out lists what the vocab module needs from outside: the card-content generator, the topic
// word suggester and stock-photo providers. Adapters live in vocab/infrastructure.
package out

import (
	"context"
	"errors"

	"ielts-practice/app/modules/vocab/domain/entity"
)

// ErrGeneratorNotConfigured: no API key, so the item fails with a clear error code.
var ErrGeneratorNotConfigured = errors.New("content generator is not configured")

// ErrGeneratorRefused: the model declined the request (not retried).
var ErrGeneratorRefused = errors.New("content generator refused the request")

// IContentGenerator writes card content (Epic 3 §5.3). Implementations: Claude, fake.
type IContentGenerator interface {
	Configured() bool
	// Generate returns one result per request that the model answered; missing inputs are retried by the caller.
	Generate(ctx context.Context, words []entity.WordRequest) ([]entity.WordResult, entity.Usage, error)
}
