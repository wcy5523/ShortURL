package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	Snowflake SnowflakeConfig
	RateLimit RateLimitConfig
	Bloom     BloomConfig
	Stats     StatsConfig
	JWT       JWTConfig
	Email     EmailConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type MySQLConfig struct {
	DSN      string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type SnowflakeConfig struct {
	WorkerID int64
}

type RateLimitConfig struct {
	WindowSeconds int64
	MaxRequests   int64
}

type BloomConfig struct {
	Key          string
	ExpectedSize uint
	FalseRate    float64
}

type StatsConfig struct {
	BatchSize   int
	WorkerCount int
	ChannelSize int
}

type JWTConfig struct {
	Secret string
}

type EmailConfig struct {
	Enabled       bool
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPassword  string
	SMTPFrom      string
}

var AppConfig *Config

func Load() {
	AppConfig = &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "release"),
		},
		MySQL: MySQLConfig{
			DSN:      getEnv("MYSQL_DSN", ""),
			Host:     getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:     getEnv("MYSQL_PORT", "3306"),
			User:     getEnv("MYSQL_USER", "root"),
			Password: getEnv("MYSQL_PASSWORD", "123456"),
			DBName:   getEnv("MYSQL_DB_NAME", "shorturl"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 100),
		},
		Snowflake: SnowflakeConfig{
			WorkerID: int64(getEnvInt("SNOWFLAKE_WORKER_ID", 1)),
		},
		RateLimit: RateLimitConfig{
			WindowSeconds: int64(getEnvInt("RATE_LIMIT_WINDOW", 60)),
			MaxRequests:   int64(getEnvInt("RATE_LIMIT_MAX", 100)),
		},
		Bloom: BloomConfig{
			Key:          getEnv("BLOOM_KEY", "shorturl:bloom"),
			ExpectedSize: uint(getEnvInt("BLOOM_EXPECTED_SIZE", 10000000)),
			FalseRate:    0.001,
		},
		Stats: StatsConfig{
			BatchSize:   getEnvInt("STATS_BATCH_SIZE", 100),
			WorkerCount: getEnvInt("STATS_WORKER_COUNT", 4),
			ChannelSize: getEnvInt("STATS_CHANNEL_SIZE", 10000),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "shorturl_jwt_secret_key"),
		},
		Email: EmailConfig{
			Enabled:       getEnvBool("EMAIL_ENABLED", false),
			SMTPHost:      getEnv("SMTP_HOST", "smtp.qq.com"),
			SMTPPort:      getEnv("SMTP_PORT", "587"),
			SMTPUser:      getEnv("SMTP_USER", "your_email@qq.com"),
			SMTPPassword:  getEnv("SMTP_PASSWORD", "your_email_authorization_code"),
			SMTPFrom:      getEnv("SMTP_FROM", "your_email@qq.com"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}
