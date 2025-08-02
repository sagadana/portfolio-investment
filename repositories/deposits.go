package repositories

import (
	"context"
	"fmt"
	"portfolio-investment/configs"
	"portfolio-investment/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PUBLIC: Create transaction records for deposits
func CreateDepositTransaction(
	ctx *context.Context,
	userReferenceID string,
	amount float64,
) (*database.Transaction, error) {
	db := database.WithContext(ctx)

	user, err := GetUser(ctx, userReferenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for reference ID (%s): %w", userReferenceID, err)
	}

	trxnReferenceID := uuid.New().String()
	transaction := &database.Transaction{
		ReferenceID: trxnReferenceID,
		User:        *user,
		Type:        configs.TrxnTypeDeposit,
		Amount:      amount,
		Processed:   false,
	}
	err = db.Create(transaction).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create transaction for user reference ID (%s) and amount (%f): %w",
			userReferenceID, amount, err,
		)
	}

	return transaction, nil
}

// PRIVATE: Allocate funds to each plan based on the plan amounts
func allocateFunds(
	tx *gorm.DB,
	plans []database.UserDepositPlan,
	transaction *database.Transaction,
	fund float64,
) error {
	total := 0.0
	for _, plan := range plans {
		total += plan.Amount
	}

	remainder := total
	for i, plan := range plans {
		allocatedAmount := 0.0

		// Calculate ration to use for allocation
		ratio := plan.Amount / total

		// Last portfolio gets the remaining amount
		if i == len(plans)-1 {
			allocatedAmount = remainder
		} else {
			allocatedAmount = fund * ratio
		}

		// Create  deposit
		deposit := &database.Deposit{
			Transaction: *transaction,
			Plan:        plan,
			Amount:      allocatedAmount,
		}
		err := tx.Create(deposit).Error
		if err != nil {
			return err
		}

		// Update the user portfolio fund
		// 1. Get user portfolio
		var userPortfolio database.UserPortfolio
		err = tx.Where(&database.UserPortfolio{
			User:      database.User{ReferenceID: plan.User.ReferenceID},
			Portfolio: database.Portfolio{ReferenceID: plan.Portfolio.ReferenceID},
		}).First(&userPortfolio).Error
		if err != nil {
			return err
		}
		// 2. Update Fund
		userPortfolio.Fund += deposit.Amount
		err = tx.Save(&userPortfolio).Error
		if err != nil {
			return err
		}

		// Deduct allocated amount from allocation pool
		remainder -= deposit.Amount

		fmt.Printf("Allocated %f to plan %s (Transaction ID: %s). Remainder: %f\n",
			deposit.Amount, plan.Portfolio.Name, deposit.Transaction.ReferenceID, deposit.Amount)

	}

	return nil
}

// PUBLIC: Deposit funds fairly to user's deposit plan portfolios
func DepositFunds(
	ctx *context.Context,
	transaction *database.Transaction,
	plans []database.UserDepositPlan,
) error {

	// Use a transaction to ensure atomicity
	// Calculate total amount to be allocated
	// Iterate over plans and allocate funds based on the plan amounts
	return database.WithTransaction(ctx, func(tx *gorm.DB) error {

		// Separate plans into one-time and monthly
		var oneTimePlans []database.UserDepositPlan
		var monthlyPlans []database.UserDepositPlan
		for _, plan := range plans {
			switch plan.Type {
			case configs.PlanTypeOnceTime:
				oneTimePlans = append(oneTimePlans, plan)
			case configs.PlanTypeMonthly:
				monthlyPlans = append(monthlyPlans, plan)
			}
		}
		if len(monthlyPlans) == 0 && len(oneTimePlans) == 0 {
			return fmt.Errorf("no deposit plans found for user %s", transaction.User.ReferenceID)
		}

		// Get total funds and plan amount for 'one-time' plans
		totalOneTimeFund := 0.0
		totalOneTimePlanAmount := 0.0
		for _, plan := range oneTimePlans {
			var userPortfolio database.UserPortfolio
			err := tx.Where(&database.UserPortfolio{
				User:      database.User{ReferenceID: transaction.User.ReferenceID},
				Portfolio: database.Portfolio{ReferenceID: plan.Portfolio.ReferenceID},
			}).First(&userPortfolio).Error
			if err != nil {
				return fmt.Errorf("failed to get user portfolio for plan %s: %w", plan.Portfolio.Name, err)
			}
			totalOneTimeFund += userPortfolio.Fund
			totalOneTimePlanAmount += plan.Amount
		}

		// If monthly plans are empty, allocate funds to one-time plans
		// Else, first check if sum of total funds for 'one-time' plan portfolios is less than the planned amount,
		// - If true, then deposit to 'one-time' plan first before allocating the remaining funds to 'monthly' plan

		remainingFund := transaction.Amount
		remainingOneTimeFund := totalOneTimePlanAmount - totalOneTimeFund

		oneTimeAllocation := remainingFund
		if len(monthlyPlans) > 0 && remainingOneTimeFund < oneTimeAllocation {
			oneTimeAllocation = remainingOneTimeFund
		}

		// Deposit allocated funds into the 'one-time' plan portfolios
		if oneTimeAllocation > 0 {
			err := allocateFunds(tx, oneTimePlans, transaction, oneTimeAllocation)
			if err != nil {
				return fmt.Errorf("failed to deposit funds: %w", err)
			}

			// Deduct allocated amount from remaining fund
			remainingFund -= oneTimeAllocation
			fmt.Printf("Deposited %.2f to one-time plans for user %s\n", oneTimeAllocation, transaction.User.ReferenceID)
		}

		// Deposit remaining funds into the 'monthly' plan portfolios
		if remainingFund > 0 {
			err := allocateFunds(tx, monthlyPlans, transaction, remainingFund)
			if err != nil {
				return fmt.Errorf("failed to deposit funds: %w", err)
			}
			fmt.Printf("Deposited %.2f to monthly plans for user %s\n", remainingFund, transaction.User.ReferenceID)
		}

		// Update the transaction as processed
		transaction.Processed = true
		err := tx.Save(transaction).Error
		if err != nil {
			return err
		}
		return nil

	})
}
