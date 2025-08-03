package configs

type DBType string

const (
	SQLite   DBType = "sqlite"
	MySQL    DBType = "mysql"
	Postgres DBType = "postgres"
)

type AppConfig struct {
	DatabaseType        DBType
	DatabaseDSN         string
	DatabaseDryrun      bool
	DatabaseAutoMigrate bool
	DatabaseAutoSeed    bool
}

type PlanType string

const (
	PlanTypeOnceTime PlanType = "onetime"
	PlanTypeMonthly  PlanType = "monthly"
)

type TransactionType string

const (
	TrxnTypeDeposit    TransactionType = "deposit"
	TrxnTypeWithdrawal TransactionType = "withdrawal"
)

const (
	DefaultPortfolioRetirement string = "portfolio-retirement"
	DefaultPortfolioHighRisk   string = "portfolio-high-risk"
)
