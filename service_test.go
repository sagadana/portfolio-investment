package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestProcessFunds(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		time.Sleep(5 * time.Second) // Wait for logs
		cancel()
	}()

	userReferenceID := "user-123"

	var tests = []struct {
		name  string
		funds []float64
	}{
		{"Test with valid happy case whole number funds", []float64{10500.0, 100.0}},
		{"Test with valid decimal funds", []float64{10500.50, 100.25}},
		{"Test with valid mixed funds", []float64{10500.0, 100.25, 200.75}},
		{"Test with large fund", []float64{100000.0}},
		{"Test with zero fund", []float64{0.0}},
		{"Test with negative fund", []float64{-100.0}},
		{"Test with mixed valid and invalid funds", []float64{10500.0, -100.0, 200.0, 0.0, 300.75}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funds := tt.funds
			totalFunds := 0.0
			for _, fund := range funds {
				totalFunds += fund
			}

			fmt.Println("\n----------------------------------------------------------------------------------------------------")

			// Get funds before processing
			oldTotals, err := GetPortfolioTotalFunds(&ctx, userReferenceID)
			if err != nil {
				t.Fatalf("GetUserTotalFunds failed: %v", err)
			}
			fmt.Printf("📌 Old total funds: %v\n", oldTotals)

			// Process funds
			resultTotals, err := ProcessFunds(&ctx, userReferenceID, funds)
			if err != nil {
				t.Fatalf("ProcessFunds failed: %v", err)
			}
			fmt.Printf("📌 Result total funds: %v\n", resultTotals)

			// Get funds after processing
			newTotals, err := GetPortfolioTotalFunds(&ctx, userReferenceID)
			if err != nil {
				t.Fatalf("ProcessFunds failed: %v", err)
			}
			fmt.Printf("📌 New total funds: %v\n", newTotals)

			// Check if the new totals are greater than the old totals
			for portfolioReferenceID, resultTotal := range resultTotals {
				oldTotal, oldTotalExists := oldTotals[portfolioReferenceID]
				newTotal, newTotalExists := newTotals[portfolioReferenceID]
				if !oldTotalExists {
					t.Errorf("❌ Portfolio ID %s does not exist in old totals", portfolioReferenceID)
				} else if !newTotalExists {
					t.Errorf("❌ Portfolio ID %s does not exist in new totals", portfolioReferenceID)
				} else if resultTotal < oldTotal {
					t.Errorf("❌ Expected total funds for portfolio '%s' to increase, but got %.2f (old: %.2f)", portfolioReferenceID, resultTotal, oldTotal)
				} else if resultTotal < oldTotal {
					t.Errorf("❌ Expected total funds for portfolio '%s' to increase, but got %.2f (old: %.2f)", portfolioReferenceID, resultTotal, oldTotal)
				} else if resultTotal != newTotal {
					t.Errorf("❌ Mismatch in total funds for portfolio '%s': expected %.2f, got %.2f", portfolioReferenceID, resultTotal, newTotal)
				} else {
					t.Logf("✅ Portfolio '%s' total funds increased as expected: %.2f (old: %.2f)",
						portfolioReferenceID, resultTotal, oldTotal)
				}
			}

			fmt.Print("----------------------------------------------------------------------------------------------------\n\n")
		})
	}

}
