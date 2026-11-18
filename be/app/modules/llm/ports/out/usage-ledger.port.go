// Package out lists what the llm module needs from outside: a ledger for the tokens every model call uses.
package out

import (
	"context"

	llmin "ielts-practice/app/modules/llm/ports/in"
)

const (
	ProviderAnthropic = "anthropic"
	ProviderGoogle    = "google"
)

// UsageRecord is one model call. Usage.InputTokens is the uncached input (cache reads are counted apart).
type UsageRecord struct {
	Provider string
	Model    string
	Feature  string
	OK       bool
	llmin.Usage
}

// IUsageLedger stores model calls for the admin board (Epic 7). Implemented in llm/infrastructure/ledger.
type IUsageLedger interface {
	Record(ctx context.Context, r UsageRecord) error
}
