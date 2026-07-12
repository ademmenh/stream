package infrastructure

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-starter/internal/config/domain"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOrDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBoolOrDefault(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		switch strings.ToLower(v) {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}
	return def
}

type ConfigAdapter struct{}

func NewConfigAdapter() domain.IConfig {
	return &ConfigAdapter{}
}

func (c *ConfigAdapter) Env() string         { return envOrDefault("ENV", "dev") }
func (c *ConfigAdapter) AppName() string     { return envOrDefault("APP_NAME", "waslini") }
func (c *ConfigAdapter) APIVersion() string  { return envOrDefault("API_VERSION", "1") }
func (c *ConfigAdapter) Port() string        { return envOrDefault("PORT", "8000") }
func (c *ConfigAdapter) DBHost() string      { return envOrDefault("DB_HOST", "localhost") }
func (c *ConfigAdapter) DBPort() string      { return envOrDefault("DB_PORT", "5432") }
func (c *ConfigAdapter) DBUser() string      { return envOrDefault("DB_USER", "postgres") }
func (c *ConfigAdapter) DBPassword() string  { return envOrDefault("DB_PASSWORD", "postgres") }
func (c *ConfigAdapter) DBName() string      { return envOrDefault("DB_NAME", "waslini") }
func (c *ConfigAdapter) LogsDirname() string { return envOrDefault("LOGS_DIRNAME", "logs") }
func (c *ConfigAdapter) RetentionDays() int  { 	return envIntOrDefault("RETENTION_DAYS", 30) }
func (c *ConfigAdapter) JWTAccessTokenSecret() string {
	return envOrDefault("JWT_ACCESS_TOKEN_SECRET", "your_access_token_secret")
}
func (c *ConfigAdapter) JWTRefreshTokenSecret() string {
	return envOrDefault("JWT_REFRESH_TOKEN_SECRET", "your_refresh_token_secret")
}
func (c *ConfigAdapter) JWTAccessTokenExpiry() int {
	return envIntOrDefault("JWT_ACCESS_TOKEN_EXPIRY", 3600)
}
func (c *ConfigAdapter) JWTRefreshTokenExpiry() int {
	return envIntOrDefault("JWT_REFRESH_TOKEN_EXPIRY", 604800)
}
func (c *ConfigAdapter) JWTAlgo() string         { return envOrDefault("JWT_ALGO", "HS256") }
func (c *ConfigAdapter) CookiesSecure() bool     { return envBoolOrDefault("COOKIES_SECURE", true) }
func (c *ConfigAdapter) CookiesSameSite() string { return envOrDefault("COOKIES_SAME_SITE", "lax") }
func (c *ConfigAdapter) CORSOrigins() []string {
	raw := envOrDefault("CORS_ORIGINS", `["*"]`)
	raw = strings.Trim(raw, "[]")
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(strings.Trim(p, `"`))
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}
func (c *ConfigAdapter) CORSCredentials() bool { return envBoolOrDefault("CORS_CREDENTIALS", true) }
func (c *ConfigAdapter) DatabaseURL() string {
	ssl := c.SSLMode()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost(), c.DBPort(), c.DBUser(), c.DBPassword(), c.DBName(), ssl)
	return dsn
}
func (c *ConfigAdapter) SSLMode() string   { return envOrDefault("DB_SSLMODE", "disable") }
func (c *ConfigAdapter) JWTSecret() string { return c.JWTAccessTokenSecret() }
func (c *ConfigAdapter) Debug() bool       { return c.Env() == "dev" }

func (c *ConfigAdapter) S3Host() string     { return envOrDefault("S3_HOST", "s3") }
func (c *ConfigAdapter) S3Port() string     { return envOrDefault("S3_PORT", "9000") }
func (c *ConfigAdapter) S3Region() string   { return envOrDefault("S3_REGION", "us-east-1") }
func (c *ConfigAdapter) S3AccessKey() string { return envOrDefault("S3_ACCESS_KEY", "minioadmin") }
func (c *ConfigAdapter) S3SecretKey() string { return envOrDefault("S3_SECRET_KEY", "minioadmin") }
func (c *ConfigAdapter) S3Bucket() string    { return envOrDefault("S3_BUCKET", "starter") }
func (c *ConfigAdapter) S3PublicEndpoint() string {
	return envOrDefault("S3_PUBLIC_ENDPOINT", "http://s3:9000")
}

func (c *ConfigAdapter) AdminEmail() string {
	return envOrDefault("ADMIN_EMAIL", "admin@gmail.com")
}

func (c *ConfigAdapter) AdminPassword() string {
	return envOrDefault("ADMIN_PASSWORD", "admin123")
}
