package database

import (
	"portfolio-investment/configs"

	"gorm.io/gorm"
)

func SeedPortfolios(db *gorm.DB) []Portfolio {
	portfolios := []Portfolio{
		{
			ReferenceID: configs.DefaultPortfolioRetirement,
			Name:        "Retirement",
			Assets: []Asset{
				{Name: "Apple Inc.", Class: "Stock"},
				{Name: "Tesla Inc.", Class: "Stock"},
			},
		},
		{
			ReferenceID: configs.DefaultPortfolioHighRisk,
			Name:        "High Risk",
			Assets: []Asset{
				{Name: "Bitcoin", Class: "Cryptocurrency"},
				{Name: "Ethereum", Class: "Cryptocurrency"},
			},
		},
		{
			ReferenceID: "portfolio-low-risk",
			Name:        "Low Risk",
			Assets: []Asset{
				{Name: "US Treasury Bonds", Class: "Bond"},
				{Name: "Gold", Class: "Commodity"},
			},
		},
	}
	if err := db.Create(&portfolios).Error; err != nil {
		panic("Failed to seed portfolios: " + err.Error())
	}

	return portfolios
}

func SeedUser(db *gorm.DB) User {
	user := User{ReferenceID: "user-123"}
	if err := db.Create(&user).Error; err != nil {
		panic("Failed to seed user: " + err.Error())
	}

	return user
}

func SeedUserPortfolios(db *gorm.DB, userPortfolios *[]UserPortfolio) {
	if err := db.Create(&userPortfolios).Error; err != nil {
		panic("Failed to seed user portfolios: " + err.Error())
	}
}

func SeedUserDepositPlans(db *gorm.DB, plans *[]UserDepositPlan) {
	if err := db.Create(&plans).Error; err != nil {
		panic("Failed to seed user: " + err.Error())
	}
}
