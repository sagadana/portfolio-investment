package main

import (
	"context"
	"fmt"
	"portfolio-investment/repositories"
)

func ProcessFunds(
	ctx *context.Context,
	userReferenceID string,
	fund float64,
) error {

	// Create a transaction for the deposit
	transaction, err := repositories.CreateDepositTransaction(ctx, userReferenceID, fund)
	if err != nil {
		return fmt.Errorf("failed to create deposit transaction: %w", err)
	}

	// Get user deposit plans
	plans, err := repositories.GetUserDepositPlans(ctx, userReferenceID)
	if err != nil {
		return fmt.Errorf("failed to get user one-time deposit plans: %w", err)
	}

	// Deposit funds into the plans
	err = repositories.DepositFunds(ctx, transaction, plans)
	if err != nil {
		return fmt.Errorf("failed to deposit funds: %w", err)
	}

	fmt.Printf("Compleded depositing %.2f to user's deposit plan(s) portfolios %s\n", transaction.Amount, userReferenceID)
	return nil
}
