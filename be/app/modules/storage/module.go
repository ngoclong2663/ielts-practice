package storage

import (
	"go.uber.org/fx"

	"ielts-practice/app/modules/storage/infrastructure/s3"
)

func NewStorageModule() fx.Option {
	return fx.Options(
		fx.Provide(s3.NewS3Storage),
	)
}
