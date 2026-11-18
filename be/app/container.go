package app

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/admin"
	"ielts-practice/app/modules/auth"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/decks"
	"ielts-practice/app/modules/feedback"
	"ielts-practice/app/modules/guides"
	"ielts-practice/app/modules/health"
	"ielts-practice/app/modules/llm"
	"ielts-practice/app/modules/notify"
	"ielts-practice/app/modules/storage"
	"ielts-practice/app/modules/study"
	"ielts-practice/app/modules/users"
	"ielts-practice/app/modules/vocab"
	"ielts-practice/app/modules/writing"
	"ielts-practice/configs"
)

func BuildContainer() *fx.App {
	return fx.New(
		CoreOptions(),
		fx.Invoke(
			RunHTTPServer,
		),
		fx.NopLogger,
	)
}

// CoreOptions wires every module without starting the HTTP listener (reused by integration tests).
func CoreOptions() fx.Option {
	return fx.Options(
		fx.Provide(
			configs.LoadConfig,
			configs.InitLoggerConfig,
			NewGinEngine,
			NewHTTPServer,
			NewJWTManager,
			NewAdminJWT,
		),
		database.NewDatabaseModule(),
		storage.NewStorageModule(),
		llm.NewLLMModule(),
		notify.NewNotifyModule(),
		health.NewHealthModule(),
		users.NewUserModule(),
		auth.NewAuthModule(),
		decks.NewDecksModule(),
		study.NewStudyModule(),
		vocab.NewVocabModule(),
		writing.NewWritingModule(),
		feedback.NewFeedbackModule(),
		admin.NewAdminModule(),
		guides.NewGuidesModule(),
	)
}
