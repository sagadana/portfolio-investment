package main

import (
	"context"
	"fmt"
	"portfolio-investment/repositories"
)

func GetUserTotalFunds(ctx *context.Context, userReferenceID string) (float64, error) {
	// Get user portfolios
	userPortfolios, err := repositories.GetUserPortfolios(ctx, userReferenceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user portfolios: %w", err)
	}

	// Calculate total funds
	total := 0.0
	for _, userPortfolio := range userPortfolios {
		total += userPortfolio.Fund
	}

	return total, nil
}

func GetPortfolioTotalFunds(ctx *context.Context, userReferenceID string) (map[string]float64, error) {
	// Get user portfolios
	userPortfolios, err := repositories.GetUserPortfolios(ctx, userReferenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user portfolios: %w", err)
	}

	totals := make(map[string]float64)

	// Calculate total funds
	for _, userPortfolio := range userPortfolios {
		portfolioReferenceID := userPortfolio.Portfolio.ReferenceID
		if _, exists := totals[portfolioReferenceID]; !exists {
			totals[portfolioReferenceID] = 0
		}
		totals[portfolioReferenceID] += userPortfolio.Fund
	}

	return totals, nil
}

func ProcessFunds(
	ctx *context.Context,
	userReferenceID string,
	funds []float64,
) (map[string]float64, error) {

	validFunds := []float64{}
	for _, fund := range funds {
		if fund > 0 {
			validFunds = append(validFunds, fund)
		}
	}
	if len(validFunds) == 0 {
		fmt.Println("No valid fund(s)")
		return make(map[string]float64), nil
	}

	// Create a transaction for the deposit
	transactions, err := repositories.CreateDepositTransactions(ctx, userReferenceID, validFunds)
	if err != nil {
		return nil, fmt.Errorf("failed to create deposit transaction: %w", err)
	}

	fmt.Printf("Processing funds for user (%s). Funds: %v\n", userReferenceID, funds)

	// Get user deposit plans
	plans, err := repositories.GetUserDepositPlans(ctx, userReferenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user one-time deposit plans: %w", err)
	}

	// Deposit funds into the plans
	results, err := repositories.DepositFunds(ctx, transactions, plans)
	if err != nil {
		return nil, fmt.Errorf("failed to deposit funds: %w", err)
	}

	fmt.Printf("Completed depositing funds to user (%s). Funds: %v\n", userReferenceID, funds)
	return results, nil
}
