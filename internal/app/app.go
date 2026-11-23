package app

import (
	"context"
	"fmt"
)

// Run is the composition root. UI and provider wiring will be added after
// backend contracts and provider behaviour are verified.
func Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil { return fmt.Errorf("application startup: %w", err) }
	return nil
}
