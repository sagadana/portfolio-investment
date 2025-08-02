package database

import (
	"context"
	"fmt"
	"portfolio-investment/configs"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	once       sync.Once
	dbInstance *gorm.DB
)

func Migrate(db *gorm.DB) {
	db.AutoMigrate(
		&Asset{},
		&Portfolio{},
		&User{},
		&UserPortfolio{},
		&UserDepositPlan{},
		&Transaction{},
		&Deposit{},
	)
}

func Seed(db *gorm.DB) {

	// Seed Portfolios
	portfolios := SeedPortfolios(db)

	// Seed User
	user := SeedUser(db)

	// Seed User Portfolios & Deposit Plans
	var userPortfolios []UserPortfolio
	var userDepositPlans []UserDepositPlan

	for _, portfolio := range portfolios {
		switch portfolio.Name {
		case "Retirement":
			userPortfolio := UserPortfolio{
				User:      user,
				Portfolio: portfolio,
				Fund:      0,
			}
			oneTimePlan := UserDepositPlan{
				User:      user,
				Type:      configs.PlanTypeOnceTime,
				Portfolio: portfolio,
				Amount:    1000.0, // Example amount
			}
			monthlyPlan := UserDepositPlan{
				User:      user,
				Type:      configs.PlanTypeMonthly,
				Portfolio: portfolio,
				Amount:    100.0, // Example amount
			}
			userPortfolios = append(userPortfolios, userPortfolio)
			userDepositPlans = append(userDepositPlans, oneTimePlan, monthlyPlan)
		case "High Risk":
			userPortfolio := UserPortfolio{
				User:      user,
				Portfolio: portfolio,
				Fund:      0,
			}
			oneTimePlan := UserDepositPlan{
				User:      user,
				Type:      configs.PlanTypeOnceTime,
				Portfolio: portfolio,
				Amount:    500.0, // Example amount
			}
			monthlyPlan := UserDepositPlan{
				User:      user,
				Type:      configs.PlanTypeMonthly,
				Portfolio: portfolio,
				Amount:    100.0, // Example amount
			}
			userPortfolios = append(userPortfolios, userPortfolio)
			userDepositPlans = append(userDepositPlans, oneTimePlan, monthlyPlan)
		}
	}
	SeedUserPortfolios(db, &userPortfolios)
	SeedUserDepositPlans(db, &userDepositPlans)
}

func Connect() *gorm.DB {
	// Singleton DB Connection
	// This function will be executed exactly once, even with concurrent calls.
	once.Do(func() {
		config := configs.GetAppConfigs()
		switch config.DatabaseType {
		case configs.SQLite:
			db, err := gorm.Open(sqlite.Open(config.DatabaseDSN), &gorm.Config{
				DryRun: config.DatabaseDryrun,
			})
			if err != nil {
				panic(fmt.Errorf("failed to connect to database: %w", err))
			}
			db.Exec("PRAGMA foreign_keys = ON")    // Enable foreign key support for SQLite
			db.Exec("PRAGMA journal_mode = WAL")   // Enable Write-Ahead Logging for better concurrency
			db.Exec("PRAGMA synchronous = NORMAL") // Set synchronous mode to NORMAL for better performance
			db.Exec("PRAGMA cache_size = 10000")   // Set cache size for better performance
			dbInstance = db
		case configs.MySQL:
			// MySQL connection logic here
			panic("MySQL connection not implemented")
		case configs.Postgres:
			// Postgres connection logic here
			panic("Postgres connection not implemented")
		default:
			panic(fmt.Sprintf("Unsupported database type: %s", config.DatabaseType))
		}

		// Auto Migrate DB Schemas
		if config.DatabaseAutoMigrate {
			Migrate(dbInstance)
		}

		// Auto Seed
		if config.DatabaseAutoSeed {
			ctx := context.Background()
			Seed(dbInstance.WithContext(ctx))
		}
	})

	return dbInstance
}

func WithContext(ctx *context.Context) *gorm.DB {
	db := Connect()
	return db.WithContext(*ctx)
}

func WithTransaction(ctx *context.Context, handler func(tx *gorm.DB) error) error {
	db := Connect()
	return db.WithContext(*ctx).Transaction(handler)
}
