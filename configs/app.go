package configs

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

func GetEnv(key string) string {
	return os.Getenv(key)
}

func SetEnv(key, value string) {
	err := os.Setenv(key, value)
	if err != nil {
		log.Fatalf("Failed to set environment variable %s: %v", key, err)
	}
}

var (
	once      sync.Once
	appConfig *AppConfig
)

func GetAppConfigs() *AppConfig {
	// Singleton App Config
	// This function will be executed exactly once, even with concurrent calls.
	once.Do(func() {
		var dsn string
		var dbType DBType

		// Load environment variables from .env file if it exists
		godotenv.Load()

		// Parse DSN
		if dsn = GetEnv("DB_FILE_PATH"); dsn != "" {
			dsn = fmt.Sprintf("file:%s%s%s?cache=shared&mode=rwc", os.TempDir(), string(os.PathSeparator), dsn)
		} else {
			dsn = GetEnv("DB_DSN")
		}

		// Parse Database Type
		if dbTypeStr := GetEnv("DB_TYPE"); dbTypeStr != "" {
			switch dbTypeStr {
			case "sqlite":
				dbType = SQLite
			case "mysql":
				dbType = MySQL
			case "postgres":
				dbType = Postgres
			default:
				log.Fatalf("Unsupported database type: %s", dbTypeStr)
			}
		}

		appConfig = &AppConfig{
			DatabaseDSN:         dsn,
			DatabaseType:        dbType,
			DatabaseDryrun:      GetEnv("DB_DRYRUN") == "true" || GetEnv("DB_DRYRUN") == "1",
			DatabaseAutoMigrate: GetEnv("DB_AUTO_MIGRATE") == "true" || GetEnv("DB_AUTO_MIGRATE") == "1",
			DatabaseAutoSeed:    GetEnv("DB_AUTO_SEED") == "true" || GetEnv("DB_AUTO_SEED") == "1",
		}
	})
	return appConfig
}
