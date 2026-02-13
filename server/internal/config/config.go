package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	MinIO     MinIOConfig     `mapstructure:"minio"`
	Upload    UploadConfig    `mapstructure:"upload"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	OAuth     OAuthConfig     `mapstructure:"oauth"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
	MaxConns int32  `mapstructure:"max_conns"`
	MinConns int32  `mapstructure:"min_conns"`
}

type RedisConfig struct {
	URL string `mapstructure:"url"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Bucket    string `mapstructure:"bucket"`
}

type UploadConfig struct {
	MaxSize int64 `mapstructure:"max_size"`
}

type RateLimitConfig struct {
	Unauthenticated int `mapstructure:"unauthenticated"`
	Authenticated   int `mapstructure:"authenticated"`
	Webhooks        int `mapstructure:"webhooks"`
	Auth            int `mapstructure:"auth"`
}

type OAuthConfig struct {
	GoogleClientID string `mapstructure:"google_client_id"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("/app")

	v.SetEnvPrefix("FEATHER")
	v.AutomaticEnv()

	// Map nested keys to env vars
	v.BindEnv("database.url", "FEATHER_DATABASE_URL")
	v.BindEnv("redis.url", "FEATHER_REDIS_URL")
	v.BindEnv("jwt.secret", "FEATHER_JWT_SECRET")
	v.BindEnv("server.port", "FEATHER_SERVER_PORT")
	v.BindEnv("minio.endpoint", "FEATHER_MINIO_ENDPOINT")
	v.BindEnv("minio.access_key", "FEATHER_MINIO_ACCESS_KEY")
	v.BindEnv("minio.secret_key", "FEATHER_MINIO_SECRET_KEY")
	v.BindEnv("minio.use_ssl", "FEATHER_MINIO_USE_SSL")
	v.BindEnv("minio.bucket", "FEATHER_MINIO_BUCKET")
	v.BindEnv("database.max_conns", "FEATHER_DATABASE_MAX_CONNS")
	v.BindEnv("database.min_conns", "FEATHER_DATABASE_MIN_CONNS")
	v.BindEnv("oauth.google_client_id", "FEATHER_OAUTH_GOOGLE_CLIENT_ID")

	// Defaults
	v.SetDefault("database.max_conns", 25)
	v.SetDefault("database.min_conns", 5)
	v.SetDefault("server.port", 8080)
	v.SetDefault("jwt.access_ttl", "15m")
	v.SetDefault("jwt.refresh_ttl", "168h")
	v.SetDefault("upload.max_size", 20971520)
	v.SetDefault("rate_limit.unauthenticated", 100)
	v.SetDefault("rate_limit.authenticated", 300)
	v.SetDefault("rate_limit.webhooks", 60)
	v.SetDefault("rate_limit.auth", 10)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
