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
	Assets      []Asset `gorm:"foreignKey:ID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

type User struct {
	gorm.Model
	ReferenceID string `gorm:"uniqueIndex"`
}

type UserPortfolio struct {
	gorm.Model
	UserID      uint      `gorm:"uniqueIndex:idx_user_portfolio"`
	User        User      `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PortfolioID uint      `gorm:"uniqueIndex:idx_user_portfolio"`
	Portfolio   Portfolio `gorm:"foreignKey:PortfolioID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Fund        float64
}

type UserDepositPlan struct {
	gorm.Model
	Type        configs.PlanType `gorm:"uniqueIndex:idx_user_plan"`
	UserID      uint             `gorm:"uniqueIndex:idx_user_plan"`
	User        User             `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PortfolioID uint             `gorm:"uniqueIndex:idx_user_plan"`
	Portfolio   Portfolio        `gorm:"foreignKey:PortfolioID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Amount      float64
}

type Transaction struct {
	gorm.Model
	ReferenceID string `gorm:"uniqueIndex"`
	UserID      uint
	User        User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Type        configs.TransactionType
	Amount      float64
	Processed   bool
}

type Deposit struct {
	gorm.Model
	TransactionID uint            `gorm:"uniqueIndex:idx_deposit"`
	Transaction   Transaction     `gorm:"foreignKey:TransactionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	PlanID        uint            `gorm:"uniqueIndex:idx_deposit"`
	Plan          UserDepositPlan `gorm:"foreignKey:PlanID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Amount        float64
}
