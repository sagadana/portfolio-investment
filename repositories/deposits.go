package repositories

import (
	"context"
	"fmt"
	"portfolio-investment/configs"
	"portfolio-investment/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

		fmt.Printf("\t\t- Allocated %f to '%s' deposit plan. Remainder: %f\n", deposit.Amount, plan.Portfolio.ReferenceID, remainder)

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

	results := make(map[string]float64)

	// Use a transaction to ensure atomicity
	// Calculate total amount to be allocated
	// Iterate over plans and allocate funds based on the plan amounts
	err := database.WithTransaction(ctx, func(tx *gorm.DB) error {

		for _, transaction := range transactions {

			// Strategy:
			// 1. First allocate funds to 'one-time' plan portfolios till planned amount is met
			// 2. Then allocate funds to 'monthly' plan portfolios till planned amount is met
			// 3. If both both 'one-time' & 'monthly' planned amount are met, distribute funds equally to both

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
			if oneTimeAllocation > 0 || (len(monthlyPlans) == 0 && remainingOneTimeFund == 0) {
				fmt.Printf("\t- Depositing %.2f to one-time plans for user %s\n", oneTimeAllocation, transaction.User.ReferenceID)
				oneTimeResult, err := allocateFunds(tx, oneTimePlans, &transaction, oneTimeAllocation)
				if err != nil {
					return fmt.Errorf("failed to deposit funds: %w", err)
				}

				// Populate resulting funds for each portfolio
				for portfolioReferenceID, funds := range oneTimeResult {
					if _, exists := results[portfolioReferenceID]; !exists {
						results[portfolioReferenceID] = 0
					}
					results[portfolioReferenceID] += funds
					fmt.Printf("\t\t - Increased portfolio '%s' total funds by %f to %f \n", portfolioReferenceID, funds, results[portfolioReferenceID])
				}

				// Deduct allocated amount from remaining fund
				remainingFund -= oneTimeAllocation
			}

			// Deposit remaining funds into the 'monthly' plan portfolios
			if remainingFund > 0 {
				fmt.Printf("\t- Depositing %.2f to monthly plans for user %s\n", remainingFund, transaction.User.ReferenceID)
				monthlyResult, err := allocateFunds(tx, monthlyPlans, &transaction, remainingFund)

				// Populate resulting funds for each portfolio
				for portfolioReferenceID, funds := range monthlyResult {
					if _, exists := results[portfolioReferenceID]; !exists {
						results[portfolioReferenceID] = 0
					}
					results[portfolioReferenceID] += funds
					fmt.Printf("\t\t - Increased portfolio '%s' total funds by %f to %f \n", portfolioReferenceID, funds, results[portfolioReferenceID])
				}

				if err != nil {
					return fmt.Errorf("failed to deposit funds: %w", err)
				}
			}

			// Update user portfolio funds
			for portfolioReferenceID, funds := range results {
				err := tx.Model(&database.UserPortfolio{}).Where(&database.UserPortfolio{
					UserID:      transaction.ID,
					PortfolioID: database.Portfolio{ReferenceID: portfolioReferenceID}.ID,
				}).Updates(&database.UserPortfolio{Fund: funds}).Error
				if err != nil {
					return fmt.Errorf("failed to update portfolio funds: %w", err)
				}
			}

			// Update the transaction as processed
			transaction.Processed = true
			err := tx.Save(&transaction).Error
			if err != nil {
				return fmt.Errorf("failed to transaction status: %w", err)
			}
		}

		return nil

	})

	return results, err
}
