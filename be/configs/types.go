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

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Google   GoogleOAuthConfig
	Storage  StorageConfig
}

func (c *Config) IsProduction() bool { return c.Server.Env == "production" }
func (c *Config) IsTest() bool       { return c.Server.Env == "test" }
