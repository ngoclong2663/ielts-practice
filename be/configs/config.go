package configs

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

// LoadConfig reads configuration from the environment. Local `.env` files are optional:
// `be/.env` is loaded first, then the shared project-root `.env` (existing variables are never overridden).
// On Railway, variables come from the service settings and no file is needed.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	env := getEnv("APP_ENV", getEnv("ENV", "development"))

	cfg := &Config{
		Server: ServerConfig{
			Port:              getEnv("PORT", "8080"),
			Env:               env,
			FrontendURL:       strings.TrimRight(getEnv("FRONTEND_URL", "http://localhost:3000"), "/"),
			TrustedProxies:    parseTrustedProxies(getEnv("TRUSTED_PROXIES", "")),
			RateLimitDisabled: getEnv("RATE_LIMIT_DISABLED", "false") == "true",
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", getEnv("DB_URL", "")),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "ielts_practice"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Auth: AuthConfig{
			JWTSecret:       getEnv("JWT_SECRET", ""),
			AccessTokenTTL:  getDuration("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL: getDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		},
		Google: GoogleOAuthConfig{
			ClientID:     getEnv("OAUTH_CLIENT_ID", ""),
			ClientSecret: getEnv("OAUTH_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("OAUTH_REDIRECT_URL", "http://localhost:3000/api/auth/google/callback"),
		},
		Storage: StorageConfig{
			Bucket:       getEnv("S3_BUCKET_NAME", ""),
			Region:       getEnv("AWS_REGION", getEnv("AWS_DEFAULT_REGION", "ap-southeast-1")),
			Endpoint:     getEnv("S3_ENDPOINT", ""),
			UsePathStyle: getEnv("S3_FORCE_PATH_STYLE", "false") == "true",
		},
		Vocab: VocabConfig{
			AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
			Model:           getEnv("VOCAB_LLM_MODEL", "claude-haiku-5-5"),
			Effort:          getEnv("VOCAB_LLM_EFFORT", "low"),
			FallbackModel:   os.Getenv("VOCAB_FALLBACK_MODEL"),
			GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
			Workers:         getInt("VOCAB_WORKERS", 4),
			DailyQuota:      getInt("VOCAB_DAILY_QUOTA", 50),
			ImageProviders:  parseList(getEnv("VOCAB_IMAGE_PROVIDERS", "pixabay,unsplash")),
			PixabayAPIKey:   getEnv("PIXABAY_API_KEY", ""),
			UnsplashKey:     getEnv("UNSPLASH_ACCESS_KEY", ""),
			FakeProviders:   env != "production" && getEnv("VOCAB_FAKE_PROVIDERS", "false") == "true",
		},
	}

	// Unset means the default fallback; set to an empty value to disable it.
	if _, set := os.LookupEnv("VOCAB_FALLBACK_MODEL"); !set {
		cfg.Vocab.FallbackModel = "gemini-3.1-flash-lite"
	}
	cfg.Vocab.FallbackModel = strings.TrimSpace(cfg.Vocab.FallbackModel)

	if cfg.Auth.JWTSecret == "" {
		if cfg.IsProduction() {
			return nil, fmt.Errorf("JWT_SECRET is required in production")
		}
		// Development convenience: a random secret means sessions are invalidated on restart.
		cfg.Auth.JWTSecret = randomHex(32)
	}
	if cfg.IsProduction() && len(cfg.Auth.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.IsProduction() && (cfg.Google.ClientID == "" || cfg.Google.ClientSecret == "") {
		return nil, fmt.Errorf("OAUTH_CLIENT_ID and OAUTH_CLIENT_SECRET are required in production")
	}

	return cfg, nil
}

func InitLoggerConfig(cfg *Config) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	if cfg.IsProduction() {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}).With().Timestamp().Logger()
}

func getEnv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getInt(key string, fallback int) int {
	var n int
	if _, err := fmt.Sscanf(getEnv(key, ""), "%d", &n); err == nil && n > 0 {
		return n
	}
	return fallback
}

func parseList(raw string) []string {
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(getEnv(key, "")); err == nil && d > 0 {
		return d
	}
	return fallback
}

// parseTrustedProxies turns "*" into "trust every proxy" (Vercel → Railway chain),
// otherwise a comma separated list of CIDRs/IPs. Empty means trust none.
func parseTrustedProxies(raw string) []string {
	if raw == "" {
		return nil
	}
	if raw == "*" {
		return []string{"0.0.0.0/0", "::/0"}
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
