package app

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"

	"ielts-practice/app/modules/admin"
	adminin "ielts-practice/app/modules/admin/ports/in"
	"ielts-practice/app/modules/database"
	"ielts-practice/app/modules/guides"
	guidesin "ielts-practice/app/modules/guides/ports/in"
	"ielts-practice/app/modules/library"
	libraryin "ielts-practice/app/modules/library/ports/in"
	"ielts-practice/app/modules/storage"
	"ielts-practice/app/modules/writing"
	writingin "ielts-practice/app/modules/writing/ports/in"
	"ielts-practice/configs"
	"ielts-practice/seeds"
)

// RunSeed applies the reference data compiled into the binary and the admin account from
// ADMIN_USERNAME / ADMIN_PASSWORD (`api seed`). It is idempotent: the container entrypoint runs it
// on every deploy, right after the migrations.
func RunSeed() error {
	var (
		seeder      writingin.ITaskSeeder
		guideSeeder guidesin.IGuideSeeder
		libSeeder   libraryin.ILibrarySeeder
		adminSeeder adminin.IAdminSeeder
		cfg         *configs.Config
	)
	fxApp := fx.New(
		fx.Provide(configs.LoadConfig, configs.InitLoggerConfig),
		database.NewDatabaseModule(),
		storage.NewStorageModule(),
		writing.NewWritingSeedModule(),
		guides.NewGuidesSeedModule(),
		library.NewLibrarySeedModule(),
		admin.NewAdminSeedModule(),
		fx.Populate(&seeder, &guideSeeder, &libSeeder, &adminSeeder, &cfg),
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

	if _, err := libSeeder.SeedLibrary(ctx, seeds.Library, "library"); err != nil {
		return fmt.Errorf("seed library: %w", err)
	}
	// After the library: seeded writing vocabulary reuses library cards.
	if _, err := seeder.SeedTasks(ctx, seeds.Writing, "writing"); err != nil {
		return fmt.Errorf("seed writing tasks: %w", err)
	}
	if _, err := guideSeeder.SeedGuides(ctx, seeds.Guides, "guides/videos.json"); err != nil {
		return fmt.Errorf("seed guide videos: %w", err)
	}
	if _, err := adminSeeder.SeedAdmin(ctx, cfg.Admin.Username, cfg.Admin.Password); err != nil {
		return fmt.Errorf("seed admin account: %w", err)
	}
	return nil
}
