package appcli

import (
	"context"
	"log/slog"
)

type AppCLI struct{}

func (app *AppCLI) Run(ctx context.Context) error {
	slog.InfoContext(ctx, "App CLI Run")
	return nil
}
