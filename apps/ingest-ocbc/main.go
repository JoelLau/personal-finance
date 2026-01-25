package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	domain "personal-finance/pkgs/domains"
	"personal-finance/pkgs/ocbc"
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

	cmd := NewIngestOCBCAccountStatemtnCSVCommand(slogger)
	if err := cmd.Run(ctx, os.Args); err != nil {
		slogger.ErrorContext(ctx, "error running command", slog.Any("error", err))
		return
	}

}

func NewIngestOCBCAccountStatemtnCSVCommand(slogger *slog.Logger) *cli.Command {
	return &cli.Command{
		Name:  "ingest",
		Usage: "parses file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "file",
				Aliases:   []string{"f"},
				Usage:     "path to csv `FILE`, not directory (e.g. path/to/ocbc_statement.csv)",
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
				"running ingest ocbc account statements csv command",
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
			if len(table) <= OCBCAccountStatementCSVSkipXRows {
				return fmt.Errorf("error parsing file contents - expected more rows in file")
			}

			truncatedTable := table[OCBCAccountStatementCSVSkipXRows:]
			sb := &strings.Builder{}
			w := csv.NewWriter(sb)
			if err := w.WriteAll(truncatedTable); err != nil {
				fmt.Println("Error writing CSV:", err)
			}
			csvStr := sb.String()

			var txRowData []ocbc.OCBCAccountTransactionItem
			slogger.InfoContext(ctx, "unmarshalling...", slog.Any("truncatedTable", truncatedTable))
			err = gocsv.Unmarshal([]byte(csvStr), &txRowData)
			if err != nil {
				return fmt.Errorf("error unmarshalling csv: %+v", err)
			}

			slogger.InfoContext(ctx, "processing data...")
			repo := domain.NewInMemoryAccountingRepository()
			for idx, row := range txRowData {
				slog.DebugContext(ctx, "processing", slog.Int("row #", idx), slog.Any("row", row))

				withdrawalAmount := 0.0
				if row.WithdrawalsSGD != "" {
					withdrawalAmount, err = strconv.ParseFloat(strings.ReplaceAll(row.WithdrawalsSGD, ",", ""), 64)
					if err != nil {
						return fmt.Errorf("error parsing credit amount: %+v", err)
					}
				}

				depositAmount := 0.0
				if row.DepositsSGD != "" {
					depositAmount, err = strconv.ParseFloat(strings.ReplaceAll(row.DepositsSGD, ",", ""), 64)
					if err != nil {
						return fmt.Errorf("error parsing debit amount: %+v", err)
					}
				}

				withdrawalInMicroSGD := int64(withdrawalAmount * 1_000_000)
				depositInMicroSGD := int64(depositAmount * 1_000_000)

				if (withdrawalInMicroSGD == 0) == (depositInMicroSGD == 0) {
					return fmt.Errorf("transaction (%d, %s) has both withdrawal (%s) and deposit (%s)", idx, row.Description, row.DepositsSGD, row.DepositsSGD)
				}

				// assume transaction is an expense if there is a withdrawal
				if withdrawalInMicroSGD > 0 {
					err = repo.CreateExpense(ctx, domain.CreateExpenseParams{
						Name:             row.Description,
						Description:      "",
						TransactedAt:     row.TransactionDate.Time,
						CreditInMicroSGD: depositInMicroSGD,
						DebitInMicroSGD:  withdrawalInMicroSGD,
					})
					if err != nil {
						return fmt.Errorf("error creating expense while processing ocbc account statement row: %+v", err)
					}

					continue
				}

				// assume transaction is income if there is a withdrawal
				err = repo.CreateIncome(ctx, domain.CreateIncomeParams{
					Name:             row.Description,
					Description:      "",
					TransactedAt:     row.TransactionDate.Time,
					CreditInMicroSGD: depositInMicroSGD,
					DebitInMicroSGD:  withdrawalInMicroSGD,
				})
				if err != nil {
					return fmt.Errorf("error creating expense while processing ocbc account statement row: %+v", err)
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

// number of rows to skip in ocbc's account statement csv
const OCBCAccountStatementCSVSkipXRows = 5
