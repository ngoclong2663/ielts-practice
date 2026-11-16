package notify

import (
	"github.com/rs/zerolog"
	"go.uber.org/fx"

	"ielts-practice/app/modules/notify/infrastructure/noop"
	"ielts-practice/app/modules/notify/infrastructure/telegram"
	notifyin "ielts-practice/app/modules/notify/ports/in"
	"ielts-practice/configs"
)

// NewNotifyModule provides notify/ports/in.IAdminNotifier to every module.
// Shared infrastructure module (like storage and llm): no routes, no use cases.
func NewNotifyModule() fx.Option {
	return fx.Provide(NewAdminNotifier)
}

// NewAdminNotifier picks Telegram when it is configured, otherwise the no-op adapter.
// It never sends in tests or with fake providers. The bot token is never logged.
func NewAdminNotifier(cfg *configs.Config, logger zerolog.Logger) notifyin.IAdminNotifier {
	tg := cfg.Telegram
	switch {
	case cfg.IsTest() || cfg.FakeProviders:
		logger.Info().Msg("Admin notifications: off (test environment or fake providers)")
		return noop.Notifier{}
	case tg.BotToken == "" || tg.ChatID == "":
		logger.Info().Msg("Admin notifications: off (TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is empty)")
		return noop.Notifier{}
	}
	logger.Info().Str("chat_id", tg.ChatID).Int64("thread_id", tg.ThreadID).Msg("Admin notifications: Telegram")
	return telegram.NewNotifier(tg.BotToken, tg.ChatID, tg.ThreadID, telegram.DefaultBaseURL)
}
