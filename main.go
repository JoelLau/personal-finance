package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))
	slogger.InfoContext(ctx, "initializing...")

	cmd := NewIngestDBSCreditCardCSVCommand(slogger)
	if err := cmd.Run(ctx, os.Args); err != nil {
		slogger.ErrorContext(ctx, "error running command", slog.Any("error", err))
		return
	}

}

func NewIngestDBSCreditCardCSVCommand(slogger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:  "ingest",
		Usage: "parses file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "file",
				Aliases:   []string{"f"},
				Usage:     "path to csv `FILE`, not directory (e.g. path/to/dbs_creditcard.csv)",
				TakesFile: true,
				Required:  true,
			},
			&cli.StringFlag{
				Name:    "month",
				Aliases: []string{"m"},
				Usage: fmt.Sprintf("month filter in `yyyy-mm` - ignores all records in file that are not in month (e.g. %s)",
					GetLastMonthYYYYMM(DefaultNower),
				),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			slogger.InfoContext(ctx, "running ingest dbs credit card csv command", slog.String("args.file", c.String("file")), slog.String("args.month", c.String("month")))

			return fmt.Errorf("not implemented")
		},
	}
}

var DefaultNower = TimeNower{}

type TimeNower struct{}

func (n TimeNower) Now() time.Time {
	return time.Now()
}

// returns last month as string in format yyyy-mm
// e.g. if its 2025-12-13, return 2025-11 (M - 1)
func GetLastMonthYYYYMM(nower Nower) string {
	now := nower.Now()
	lastMonthTime := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())

	return lastMonthTime.Format("2006-01")
}

type Nower interface {
	Now() time.Time
}
