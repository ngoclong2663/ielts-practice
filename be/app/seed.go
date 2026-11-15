package app

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"

	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/storage"
	"ielts-practice/app/modules/writing"
	writingin "ielts-practice/app/modules/writing/ports/in"
	"ielts-practice/configs"
	"ielts-practice/seeds"
)

// RunSeed applies the reference data compiled into the binary (`api seed`). It is idempotent:
// the container entrypoint runs it on every deploy, right after the migrations.
func RunSeed() error {
	var seeder writingin.ITaskSeeder
	fxApp := fx.New(
		fx.Provide(configs.LoadConfig, configs.InitLoggerConfig),
		database.NewDatabaseModule(),
		storage.NewStorageModule(),
		writing.NewWritingSeedModule(),
		fx.Populate(&seeder),
		fx.NopLogger,
	)
	if err := fxApp.Err(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := fxApp.Start(ctx); err != nil {
		return err
	}
	defer func() { _ = fxApp.Stop(context.Background()) }()

	if _, err := seeder.SeedTasks(ctx, seeds.Writing, "writing"); err != nil {
		return fmt.Errorf("seed writing tasks: %w", err)
	}
	return nil
}
