package configs

import "time"

type ServerConfig struct {
	Port           string
	Env            string // development | test | production
	FrontendURL    string
	TrustedProxies []string
	// RateLimitDisabled is honoured only outside production (used by integration tests).
	RateLimitDisabled bool
}

type DatabaseConfig struct {
	URL      string // takes precedence over the individual fields (Railway provides DATABASE_URL)
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// StorageConfig: S3 (or an S3-compatible server such as MinIO/RustFS when Endpoint is set).
// Credentials come from the standard AWS chain (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY).
type StorageConfig struct {
	Bucket       string
	Region       string
	Endpoint     string
	UsePathStyle bool
}

// LLMConfig: the language models shared by every module (card content, word lists, writing practice).
// Env names keep their historical VOCAB_ prefix. Keys stay on the backend only.
type LLMConfig struct {
	AnthropicAPIKey string
	Model           string // VOCAB_LLM_MODEL: the primary (Claude) model
	Effort          string // VOCAB_LLM_EFFORT: low | medium | high
	// FallbackModel (Gemini) answers when the primary fails; empty disables it.
	FallbackModel string
	GeminiAPIKey  string
}

// VocabConfig: card generation (Epic 3) and words by topic (Epic 4).
type VocabConfig struct {
	Workers    int
	DailyQuota int // uncached words per user per 24 h
	// DailySuggestions: AI word-list requests (Epic 4: a new topic list or "Suggest more") per user per 24 h.
	DailySuggestions int
	ImageProviders   []string // search order, e.g. pixabay,unsplash
	PixabayAPIKey    string
	UnsplashKey      string // Unsplash access key
}

// WritingConfig: guided writing practice (Epic 5). It uses the LLM models and keys.
type WritingConfig struct {
	DailyChecks      int // paragraph checks per user per 24 h
	DailySubmissions int // full-answer feedbacks per user per 24 h
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Google   GoogleOAuthConfig
	Storage  StorageConfig
	LLM      LLMConfig
	Vocab    VocabConfig
	Writing  WritingConfig
	// FakeProviders replaces the LLMs and image providers with offline fakes (VOCAB_FAKE_PROVIDERS;
	// tests and e2e only, never in production).
	FakeProviders bool
}

func (c *Config) IsProduction() bool { return c.Server.Env == "production" }
func (c *Config) IsTest() bool       { return c.Server.Env == "test" }
