package llm

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/llm/infrastructure/chain"
)

// NewLLMModule provides the language-model chain (llm/ports/in.IJSONModel) to every module.
// Shared infrastructure module (like database and storage): no routes, no use cases.
func NewLLMModule() fx.Option {
	return fx.Provide(chain.NewJSONModel)
}
