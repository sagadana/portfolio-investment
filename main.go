package main

import (
	"context"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.TODO(), 15000*time.Millisecond)
	defer cancel()

	ProcessFunds(&ctx, "user-123", 12000)
}
