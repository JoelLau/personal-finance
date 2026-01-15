package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"personal-finance/internal/usecases"
	"personal-finance/packages/nower"

	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	cmd := NewRootCmd()
	if err := cmd.Run(ctx, os.Args); err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)

		return
	}
}

// checks sub-commands
func NewRootCmd() *cli.Command {
	return &cli.Command{
		Name:        "pf",
		Usage:       "e.g. fi ingest dbs_credit_card.csv --month=2026-01",
		Description: "personal finance CLI",
		Commands: []*cli.Command{
			NewIngestCmd(),
		},
	}
}

var ErrInvalidParam = errors.New("invalid parameter")

// parses flags, calls use cases
func NewIngestCmd() *cli.Command {
	uc := usecases.NewIngestFileUseCase()
	lastMonth := nower.GetLastMonthYYYYMM(nower.DefaultNower)

	return &cli.Command{
		Name:        "ingest",
		Usage:       fmt.Sprintf("e.g. pf ingest dbs_credit_card.csv --month=%s", lastMonth),
		Description: "parses a file, adding them to the database",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "month",
				Aliases: []string{"m"},
				Usage: fmt.Sprintf("month filter in `yyyy-mm` - take only rows that are in the defined month (e.g. %s)",
					lastMonth,
				),
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:      "filepath",
				UsageText: "e.g. usage text",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			slog.InfoContext(ctx, "ingest")

			filePath := cmd.StringArg("filepath")
			if filePath == "" {
				return fmt.Errorf("%w: filepath is required - %s", ErrInvalidParam, cmd.Usage)
			}

			monthFilter := cmd.String("month")

			if err := uc.Execute(ctx, filePath, monthFilter); err != nil {
				return err
			}

			return nil
		},
	}
}
