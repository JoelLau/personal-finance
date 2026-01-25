package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"personal-finance/pkgs/dbs"
	domain "personal-finance/pkgs/domains"
	"strconv"
	"strings"
	"time"

	gocsv "github.com/JoelLau/go-csv"
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
				Usage: fmt.Sprintf("month filter in `yyyy-mm` - take only rows that are in the defined month (e.g. %s)",
					GetLastMonthYYYYMM(DefaultNower),
				),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			slogger.InfoContext(ctx,
				"running ingest dbs credit card csv command",
				slog.String("args.file", c.String("file")),
				slog.String("args.month", c.String("month")),
			)

			filepath := c.String("file")
			slogger.InfoContext(ctx, "opening file handle", slog.Any("filepath", filepath))
			file, err := os.Open(filepath)
			if err != nil {
				return fmt.Errorf("error opening file at at '%s': %+v", filepath, err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					err := fmt.Errorf("error closing file: %+v", err)
					slogger.ErrorContext(ctx, "error closing file", slog.Any("error", err))
					return
				}
			}()

			slogger.InfoContext(ctx, "reading file")
			fileBytes, err := io.ReadAll(file)
			if err != nil {
				return fmt.Errorf("error reading bytes from reader: %+v", err)
			}

			table, err := gocsv.ReadAll(fileBytes)
			if err != nil {
				return fmt.Errorf("error reading converting file bytes to string 2D array")
			}

			// the first X rows contain metadata like the credit card and bank account numbers.
			// this is sensitive information that we want nothing to do with.
			if len(table) <= DBSCreditCardCSVSkipXRows {
				return fmt.Errorf("error parsing file contents - expected more rows in file")
			}

			csv := table[DBSCreditCardCSVSkipXRows:]
			rows := make([]string, len(csv))
			for idx, row := range csv {
				rows[idx] = strings.Join(row, ",")
			}
			csvStr := (strings.Join(rows, "\n"))

			var ccRowData []dbs.CreditCardItem
			slogger.InfoContext(ctx, "unmarshalling...", slog.Any("csv", csv), slog.Any("rows", rows))
			err = gocsv.Unmarshal([]byte(csvStr), &ccRowData)
			if err != nil {
				return fmt.Errorf("error unmarshalling csv: %+v", err)
			}

			slogger.InfoContext(ctx, "processing data...")
			repo := domain.NewInMemoryAccountingRepository()
			for idx, row := range ccRowData {
				slog.DebugContext(ctx, "processing", slog.Int("row #", idx), slog.Any("row", row))

				creditAmount := 0.0
				if row.CreditAmount != "" {
					creditAmount, err = strconv.ParseFloat(row.CreditAmount, 64)
					if err != nil {
						return fmt.Errorf("error parsing credit amount: %+v", err)
					}
				}

				debitAmount := 0.0
				if row.DebitAmount != "" {
					debitAmount, err = strconv.ParseFloat(row.DebitAmount, 64)
					if err != nil {
						return fmt.Errorf("error parsing debit amount: %+v", err)
					}
				}

				creditInMicroSGD := int64(creditAmount * 1_000_000)
				debitInMicroSGD := int64(debitAmount * 1_000_000)

				err = repo.CreateExpense(ctx, domain.CreateExpenseParams{
					Name:             row.TransactionDescription,
					Description:      "",
					TransactedAt:     row.TransactionDate.Time,
					CreditInMicroSGD: creditInMicroSGD,
					DebitInMicroSGD:  debitInMicroSGD,
				})
				if err != nil {
					return fmt.Errorf("error creating expense while processing dbs credit card row: %+v", err)
				}
			}

			expenses, err := repo.ListTransactions(ctx)
			if err != nil {
				return fmt.Errorf("error listing expenses: %+v", err)
			}

			slog.InfoContext(ctx, "completed", slog.Any("expenses", expenses))
			return nil
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

const DBSCreditCardCSVSkipXRows = 6
