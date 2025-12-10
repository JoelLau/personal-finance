package main

import (
	"context"
	"log/slog"
	appcli "personal-finance/internal/app-cli"
)

func main() {
	ctx := context.Background()

	slog.InfoContext(ctx, "Personal Finance CLI")
	app := appcli.AppCLI{}
	err := app.Run(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "app cli encountered a fatal error", slog.Any("error", err))
	}
}
