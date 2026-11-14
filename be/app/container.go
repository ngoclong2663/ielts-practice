package app

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/auth"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/decks"
	"ielts-practice/app/modules/health"
	"ielts-practice/app/modules/storage"
	"ielts-practice/app/modules/study"
	"ielts-practice/app/modules/users"
	"ielts-practice/app/modules/vocab"
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
		),
		database.NewDatabaseModule(),
		storage.NewStorageModule(),
		health.NewHealthModule(),
		users.NewUserModule(),
		auth.NewAuthModule(),
		decks.NewDecksModule(),
		study.NewStudyModule(),
		vocab.NewVocabModule(),
	)
}
