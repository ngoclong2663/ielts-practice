package llm

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/llm/infrastructure/chain"
	"ielts-practice/app/modules/llm/infrastructure/ledger"
)

// NewLLMModule provides the language-model chain (llm/ports/in.IJSONModel) to every module. Every model
// call is recorded in the usage ledger (table ai_usage) for the admin board.
// Shared infrastructure module (like database and storage): no routes, no use cases.
func NewLLMModule() fx.Option {
	return fx.Provide(chain.NewMeteredJSONModel, ledger.NewUsageLedger)
}
