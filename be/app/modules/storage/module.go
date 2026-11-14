package storage

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/storage/infrastructure"
)

func NewStorageModule() fx.Option {
	return fx.Options(
		fx.Provide(infrastructure.NewS3Storage),
	)
}
