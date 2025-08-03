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

func SeedUser(db *gorm.DB, userReferenceID string, portfolios []Portfolio) (*User, *[]UserPortfolio, *[]UserDepositPlan) {
	user := User{ReferenceID: userReferenceID}
	if err := db.Create(&user).Error; err != nil {
		panic("Failed to seed user: " + err.Error())
	}

	// Seed User Portfolios & Deposit Plans
	var userPortfolios []UserPortfolio
	var userDepositPlans []UserDepositPlan

	for _, portfolio := range portfolios {
		switch portfolio.ReferenceID {
		case configs.DefaultPortfolioRetirement:
			userPortfolio := UserPortfolio{
				User:      user,
				Portfolio: portfolio,
				Fund:      0,
			}
			oneTimePlan := UserDepositPlan{
				User:      user,
				Type:      configs.PlanTypeOnceTime,
				Portfolio: portfolio,
				Amount:    500.0,
			}
			monthlyPlan := UserDepositPlan{
				User:      user,
				Type:      configs.PlanTypeMonthly,
				Portfolio: portfolio,
				Amount:    100.0,
			}
			userPortfolios = append(userPortfolios, userPortfolio)
			userDepositPlans = append(userDepositPlans, oneTimePlan, monthlyPlan)
		case configs.DefaultPortfolioHighRisk:
			userPortfolio := UserPortfolio{
				User:      user,
				Portfolio: portfolio,
				Fund:      0,
			}
			oneTimePlan := UserDepositPlan{
				User:      user,
				Type:      configs.PlanTypeOnceTime,
				Portfolio: portfolio,
				Amount:    10000.0,
			}
			monthlyPlan := UserDepositPlan{
				User:      user,
				Type:      configs.PlanTypeMonthly,
				Portfolio: portfolio,
				Amount:    0.0,
			}
			userPortfolios = append(userPortfolios, userPortfolio)
			userDepositPlans = append(userDepositPlans, oneTimePlan, monthlyPlan)
		}
	}
	SeedUserPortfolios(db, &userPortfolios)
	SeedUserDepositPlans(db, &userDepositPlans)

	return &user, &userPortfolios, &userDepositPlans
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
