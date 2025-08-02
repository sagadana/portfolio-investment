package repositories

import (
	"context"
	"portfolio-investment/database"
)

// PUBLIC: Get user's record by reference ID
func GetUser(ctx *context.Context, referenceID string) (*database.User, error) {
	var user database.User
	err := database.WithContext(ctx).Where(&database.User{ReferenceID: referenceID}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// PUBLIC: Get user's deposit plan record(s) by reference ID
func GetUserDepositPlans(ctx *context.Context, referenceID string) ([]database.UserDepositPlan, error) {
	var userDepositPlans []database.UserDepositPlan
	err := database.WithContext(ctx).Where(&database.UserDepositPlan{User: database.User{ReferenceID: referenceID}}).Find(&userDepositPlans).Error
	if err != nil {
		return nil, err
	}
	return userDepositPlans, nil
}
