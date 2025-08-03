package main

import (
	"context"
	"fmt"
	"testing"
)

func TestProcessFunds(t *testing.T) {
	ctx := context.Background()
	userReferenceID := "user-123"

	var tests = []struct {
		name  string
		funds []float64
	}{
		{"Test with valid happy case whole number funds", []float64{10500.0, 100.0}},
		{"Test with valid decimal funds", []float64{10500.50, 100.25}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funds := tt.funds
			totalFunds := 0.0
			for _, fund := range funds {
				totalFunds += fund
			}

			fmt.Println("\n----------------------------------------------------------------------------------------------------")

			// Get user portfolios and calculate total funds
			oldTotals, err := GetPortfolioTotalFunds(&ctx, userReferenceID)
			if err != nil {
				t.Fatalf("GetTotalFunds failed: %v", err)
			}
			fmt.Printf("Old total funds: %v\n", oldTotals)

			newTotals, err := ProcessFunds(&ctx, userReferenceID, funds)
			if err != nil {
				t.Fatalf("ProcessFunds failed: %v", err)
			}
			fmt.Printf("New total funds: %v\n", newTotals)

			// Check if the new totals are greater than the old totals
			for portfolioReferenceID, newTotal := range newTotals {
				oldTotal, oldTotalExists := oldTotals[portfolioReferenceID]
				if !oldTotalExists {
					t.Errorf("❌ Portfolio ID %s does not exist in old totals", portfolioReferenceID)
				} else if newTotal < oldTotal {
					t.Errorf("❌ Expected total funds for portfolio '%s' to increase, but got %.2f (old: %.2f)", portfolioReferenceID, newTotal, oldTotal)
				} else {
					t.Logf("✅ Portfolio '%s' total funds increased as expected: %.2f (old: %.2f)",
						portfolioReferenceID, newTotal, oldTotal)
				}
			}

			fmt.Print("----------------------------------------------------------------------------------------------------\n\n")
		})
	}

}
