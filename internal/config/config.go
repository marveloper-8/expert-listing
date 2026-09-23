package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"expertlisting/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Port         string
	DBDriver     string
	DatabaseURL  string
	Environment  string
	AllowOrigins []string
}

func LoadConfig() *Config {
	port := firstNonEmpty(os.Getenv("PORT"), "8082")
	driver := strings.ToLower(firstNonEmpty(os.Getenv("DB_DRIVER"), "sqlite"))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	env := firstNonEmpty(os.Getenv("ENV"), "development")

	allowedOrigins := []string{"*"}
	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		allowedOrigins = strings.Split(origins, ",")
	}

	return &Config{
		Port:         port,
		DBDriver:     driver,
		DatabaseURL:  databaseURL,
		Environment:  env,
		AllowOrigins: allowedOrigins,
	}
}

func InitDB(cfg *Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	logLevel := logger.Warn
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	}

	if cfg.DBDriver == "sqlite" || strings.HasPrefix(cfg.DatabaseURL, "file:") || strings.HasSuffix(cfg.DatabaseURL, ".db") {
		dbPath := cfg.DatabaseURL
		if dbPath == "" {
			dbPath = "expertlisting.db"
		}
		dialector = sqlite.Open(dbPath)
		db, err := gorm.Open(dialector, gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			sqlDB.SetMaxOpenConns(1)
			sqlDB.SetMaxIdleConns(1)
			_ = db.Exec("PRAGMA journal_mode = WAL;").Error
			_ = db.Exec("PRAGMA synchronous = NORMAL;").Error
			_ = db.Exec("PRAGMA cache_size = -64000;").Error
			_ = db.Exec("PRAGMA temp_store = MEMORY;").Error
			_ = db.Exec("PRAGMA mmap_size = 268435456;").Error
		}

		if err := runMigrations(db); err != nil {
			return nil, err
		}
		log.Printf("Connected to SQLite: %s", dbPath)
		return db, nil
	}

	dsn := buildPostgresDSN(cfg)
	var db *gorm.DB
	var err error
	maxRetries := 10

	for i := 1; i <= maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				if err = sqlDB.Ping(); err == nil {
					sqlDB.SetMaxOpenConns(50)
					sqlDB.SetMaxIdleConns(25)
					sqlDB.SetConnMaxLifetime(15 * time.Minute)
					sqlDB.SetConnMaxIdleTime(3 * time.Minute)
					log.Println("Connected to PostgreSQL")

					if err := runMigrations(db); err != nil {
						return nil, err
					}
					return db, nil
				}
			} else {
				err = pingErr
			}
		}

		if i < maxRetries {
			time.Sleep(2 * time.Second)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

func runMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.Listing{}); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_listings_lat_lng ON listings (latitude, longitude);")
	_ = db.Exec("CREATE INDEX IF NOT EXISTS idx_listings_price_bedrooms ON listings (price, bedrooms);")
	return nil
}

func buildPostgresDSN(cfg *Config) string {
	if cfg.DatabaseURL != "" {
		return cfg.DatabaseURL
	}

	host := firstNonEmpty(os.Getenv("DB_HOST"), "localhost")
	port := firstNonEmpty(os.Getenv("DB_PORT"), "5432")
	user := firstNonEmpty(os.Getenv("DB_USER"), "postgres")
	password := os.Getenv("DB_PASSWORD")
	dbName := firstNonEmpty(os.Getenv("DB_NAME"), "expertlisting")
	sslMode := firstNonEmpty(os.Getenv("DB_SSLMODE"), "disable")
	timeZone := firstNonEmpty(os.Getenv("DB_TIMEZONE"), "UTC")

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, user, password, dbName, port, sslMode, timeZone,
	)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
