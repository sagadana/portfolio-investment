package repositories

import (
	"context"
	"fmt"
	"math"
	"portfolio-investment/configs"
	"portfolio-investment/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PRIVATE: Get current fund for user's portfolio
func getFund(
	tx *gorm.DB,
	userID uint,
	portfolioID uint,
) (float64, error) {
	var userPortfolio database.UserPortfolio
	err := tx.Where(&database.UserPortfolio{
		UserID:      userID,
		PortfolioID: portfolioID,
	}).First(&userPortfolio).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get user portfolio (user: %d, portfolio: %d): %w", userID, portfolioID, err)
	}
	return userPortfolio.Fund, nil
}

// PRIVATE: Allocate funds to each plan based on the plan amounts
// Returns allocated funds per portfolio => { PortfolioReferenceID : Allocated Fund }
func allocateFunds(
	tx *gorm.DB,
	plans []database.UserDepositPlan,
	transaction *database.Transaction,
	fund float64,
) (map[string]float64, error) {

	remainder := fund
	total := 0.0
	results := make(map[string]float64)

	// Calculate total planned amount
	// Use this to calculate the ratio for allocation
	for _, plan := range plans {
		total += plan.Amount
	}

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
		deposit := database.Deposit{
			Transaction: *transaction,
			Plan:        plan,
			Amount:      allocatedAmount,
		}
		err := tx.Create(&deposit).Error
		if err != nil {
			return nil, err
		}

		// Update allocated funds
		if _, exists := results[plan.Portfolio.ReferenceID]; !exists {
			results[plan.Portfolio.ReferenceID] = 0
		}
		results[plan.Portfolio.ReferenceID] += deposit.Amount

		// Deduct allocated amount from allocation pool
		remainder -= deposit.Amount
		if remainder < 0 {
			remainder = 0
		}

		fmt.Printf("\t\t- Allocated %f to '%s' deposit plan. Total: %f, Remainder: %f\n",
			deposit.Amount, plan.Portfolio.ReferenceID, results[plan.Portfolio.ReferenceID], remainder)

		if remainder <= 0 {
			break
		}
	}

	return results, nil
}

// PUBLIC: Create transaction records for deposits
func CreateDepositTransactions(
	ctx *context.Context,
	userReferenceID string,
	amounts []float64,
) ([]database.Transaction, error) {
	db := database.WithContext(ctx)

	user, err := GetUser(ctx, userReferenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for reference ID (%s): %w", userReferenceID, err)
	}

	transactions := make([]database.Transaction, 0, len(amounts))
	for _, amount := range amounts {
		trxnReferenceID := uuid.New().String()
		transaction := database.Transaction{
			ReferenceID: trxnReferenceID,
			User:        *user,
			Type:        configs.TrxnTypeDeposit,
			Amount:      amount,
			Processed:   false,
		}
		transactions = append(transactions, transaction)
	}
	err = db.Create(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create transaction for user reference ID (%s) and amounts (%v): %w",
			userReferenceID, amounts, err,
		)
	}

	return transactions, nil
}

// PUBLIC: Deposit funds fairly to user's deposit plan portfolios
func DepositFunds(
	ctx *context.Context,
	transactions []database.Transaction,
	plans []database.UserDepositPlan,
) (map[string]float64, error) {

	deposits := make(map[string]float64)

	// Use a transaction to ensure atomicity
	// Calculate total amount to be allocated
	// Iterate over plans and allocate funds based on the plan amounts
	err := database.WithTransaction(ctx, func(tx *gorm.DB) error {

		for _, transaction := range transactions {

			// Strategy:
			// 1. First allocate funds to 'one-time' plan portfolios till planned amount is met
			// 2. Then allocate funds to 'monthly' plan portfolios till planned amount is met
			// 3. If both 'one-time' & 'monthly' planned amount are met, distribute funds equally to both

			// Track all plan portfolios
			portfolios := make(map[string]database.Portfolio)

			// Separate plans into one-time and monthly
			var oneTimePlans []database.UserDepositPlan
			var monthlyPlans []database.UserDepositPlan
			for _, plan := range plans {
				portfolios[plan.Portfolio.ReferenceID] = plan.Portfolio
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

			// Get total current funds and total planned amounts for one-time and monthly plans
			totalOneTimeFund, totalOneTimePlanAmount := 0.0, 0.0
			totalMonthlyFund, totalMonthlyPlanAmount := 0.0, 0.0

			// Calculate total funds and planned amounts for both types
			if len(oneTimePlans) > 0 {
				for _, plan := range oneTimePlans {
					totalOneTimePlanAmount += plan.Amount
					fund, err := getFund(tx, plan.User.ID, plan.Portfolio.ID)
					if err != nil {
						return err
					}
					totalOneTimeFund += fund
				}
			}
			if len(monthlyPlans) > 0 {
				for _, plan := range monthlyPlans {
					totalMonthlyPlanAmount += plan.Amount
					fund, err := getFund(tx, plan.User.ID, plan.Portfolio.ID)
					if err != nil {
						return err
					}
					totalMonthlyFund += fund
				}
			}

			// Begin allocation
			remainingFund := transaction.Amount
			results := make(map[string]float64)

			// Step 1: Allocate to 'one-time' plans
			remainingOneTimeAllocation := totalOneTimePlanAmount - totalOneTimeFund
			oneTimeAllocation := math.Min(remainingFund, remainingOneTimeAllocation)

			if oneTimeAllocation > 0 {
				fmt.Printf("\t- Depositing %.2f to one-time plans for user %s\n", oneTimeAllocation, transaction.User.ReferenceID)
				oneTimeResult, err := allocateFunds(tx, oneTimePlans, &transaction, oneTimeAllocation)
				if err != nil {
					return fmt.Errorf("failed to deposit to one-time plans: %w", err)
				}
				for portfolioReferenceID, funds := range oneTimeResult {
					results[portfolioReferenceID] += funds
				}
				remainingFund -= oneTimeAllocation
			}

			// Step 2: Allocate to 'monthly' plans
			remainingMonthlyAllocation := totalMonthlyPlanAmount - totalMonthlyFund
			monthlyAllocation := math.Min(remainingFund, remainingMonthlyAllocation)

			if monthlyAllocation > 0 {
				fmt.Printf("\t- Depositing %.2f to monthly plans for user %s\n", monthlyAllocation, transaction.User.ReferenceID)
				monthlyResult, err := allocateFunds(tx, monthlyPlans, &transaction, monthlyAllocation)
				if err != nil {
					return fmt.Errorf("failed to deposit to monthly plans: %w", err)
				}
				for portfolioReferenceID, funds := range monthlyResult {
					results[portfolioReferenceID] += funds
				}
				remainingFund -= monthlyAllocation
			}

			// Step 3: Split remaining fund equally if both are fully funded
			if remainingFund > 0 && remainingOneTimeAllocation <= 0 && remainingMonthlyAllocation <= 0 {
				splitFund := remainingFund / 2

				if len(oneTimePlans) > 0 {
					fmt.Printf("\t- Equally splitting: Depositing %.2f of %.2f to one-time plans for user %s\n",
						splitFund, remainingFund, transaction.User.ReferenceID)
					oneTimeResult, err := allocateFunds(tx, oneTimePlans, &transaction, splitFund)
					if err != nil {
						return fmt.Errorf("failed to deposit to one-time plans in split: %w", err)
					}
					for portfolioReferenceID, funds := range oneTimeResult {
						results[portfolioReferenceID] += funds
					}
				}

				if len(monthlyPlans) > 0 {
					leftOverFund := remainingFund - splitFund
					fmt.Printf("\t- Equally splitting: Depositing %.2f of %.2f to monthly plans for user %s\n",
						leftOverFund, remainingFund, transaction.User.ReferenceID)
					monthlyResult, err := allocateFunds(tx, monthlyPlans, &transaction, leftOverFund)
					if err != nil {
						return fmt.Errorf("failed to deposit to monthly plans in split: %w", err)
					}
					for portfolioReferenceID, funds := range monthlyResult {
						results[portfolioReferenceID] += funds
					}
				}
			}

			// Update user portfolio funds
			for portfolioReferenceID, funds := range results {
				var userPortfolio database.UserPortfolio
				err := tx.Where(&database.UserPortfolio{
					UserID:      transaction.User.ID,
					PortfolioID: portfolios[portfolioReferenceID].ID,
				}).First(&userPortfolio).Error
				if err != nil {
					return fmt.Errorf("failed to get user portfolio for reference ID (%s): %w", portfolioReferenceID, err)
				}
				userPortfolio.Fund += funds
				err = tx.Save(&userPortfolio).Error
				if err != nil {
					return fmt.Errorf("failed to update user portfolio funds: %w", err)
				}
				deposits[portfolioReferenceID] = userPortfolio.Fund
			}

			// Mark transaction as processed
			transaction.Processed = true
			err := tx.Save(&transaction).Error
			if err != nil {
				return fmt.Errorf("failed to update transaction status: %w", err)
			}

		}

		return nil
	})

	return deposits, err
}
