package database

import (
	"portfolio-investment/configs"

	"gorm.io/gorm"
)

type Asset struct {
	gorm.Model
	Name  string
	Class string
}

type Portfolio struct {
	gorm.Model
	ReferenceID string `gorm:"uniqueIndex"`
	Name        string
	Assets      []Asset
}

type User struct {
	gorm.Model
	ReferenceID string `gorm:"uniqueIndex"`
}

type UserPortfolio struct {
	gorm.Model
	User      User      `gorm:"uniqueIndex:idx_user_portfolio"`
	Portfolio Portfolio `gorm:"uniqueIndex:idx_user_portfolio"`
	Fund      float64
}

type UserDepositPlan struct {
	gorm.Model
	User      User             `gorm:"uniqueIndex:idx_user_plan"`
	Type      configs.PlanType `gorm:"uniqueIndex:idx_user_plan"`
	Portfolio Portfolio        `gorm:"uniqueIndex:idx_user_plan"`
	Amount    float64
}

type Transaction struct {
	gorm.Model
	ReferenceID string `gorm:"uniqueIndex"`
	User        User
	Type        configs.TransactionType
	Amount      float64
	Processed   bool
}

type Deposit struct {
	gorm.Model
	Transaction Transaction     `gorm:"uniqueIndex:idx_deposit"`
	Plan        UserDepositPlan `gorm:"uniqueIndex:idx_deposit"`
	Amount      float64
}
