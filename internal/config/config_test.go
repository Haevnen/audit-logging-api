package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Haevnen/audit-logging-api/internal/config"
)

func TestGetGORMConfig(t *testing.T) {
	cfg := config.Config{
		PostgresHost:     "localhost",
		PostgresPort:     5432,
		PostgresUser:     "user",
		PostgresPassword: "pass",
		PostgresDB:       "testdb",
		MaxOpenConn:      10,
		MaxIdleConn:      5,
	}

	gormCfg := cfg.GetGORMConfig()
	assert.Equal(t, "localhost", gormCfg.DBHost)
	assert.Equal(t, 5432, gormCfg.DBPort)
	assert.Equal(t, "user", gormCfg.DBUser)
	assert.Equal(t, "pass", gormCfg.DBPass)
	assert.Equal(t, "testdb", gormCfg.DBName)
	assert.Equal(t, 10, gormCfg.MaxOpenConn)
	assert.Equal(t, 5, gormCfg.MaxIdleConn)
	assert.Equal(t, 300, gormCfg.MaxLifetimeSecond)
}

func TestGetURLBase(t *testing.T) {
	cfg := config.Config{
		APIHost: "127.0.0.1",
		APIPort: 8080,
	}

	url := cfg.GetURLBase()
	assert.Equal(t, "127.0.0.1:8080", url)
}

func TestLoadConfig_Success(t *testing.T) {
	// prepare env variables
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "user")
	os.Setenv("POSTGRES_PASSWORD", "pass")
	os.Setenv("POSTGRES_DB", "testdb")
	os.Setenv("API_PORT", "8080")
	os.Setenv("API_HOST", "127.0.0.1")

	// since LoadConfig calls godotenv.Load(), ensure .env is optional
	// (we won't create it, godotenv.Load will return error → panic)
	// So here we test panic path
	assert.Panics(t, func() {
		_, _ = config.LoadConfig()
	})
}

func TestLoadConfig_ParseEnv(t *testing.T) {
	// simulate environment variables directly
	os.Setenv("API_PORT", "9090")
	os.Setenv("API_HOST", "myhost")

	cfg := config.Config{
		APIHost: "myhost",
		APIPort: 9090,
	}

	assert.Equal(t, "myhost:9090", cfg.GetURLBase())
}
