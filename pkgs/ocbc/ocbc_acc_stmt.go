package ocbc

import (
	"fmt"
	"time"

	gocsv "github.com/JoelLau/go-csv"
)

// Represents a single line in OCBC's account transactions statement (csv)
type OCBCAccountTransactionItem struct {
	TransactionDate OCBCAccountTransactionsDateLayout `csv:"Transaction date"` // e.g. "22/12/25"
	ValueDate       OCBCAccountTransactionsDateLayout `csv:"Value date"`       // e.g. "23/12/25"
	Description     string                            `csv:"Description"`      // e.g. "SUPER SIMPLE           SINGAPORE     SG"
	WithdrawalsSGD  string                            `csv:"Withdrawals(SGD)"` // e.g. "6,002.94"
	DepositsSGD     string                            `csv:"Deposits(SGD)"`    // e.g. ""
}

const OCBCAccountStatementDateLayout = "2/1/2006"

type OCBCAccountTransactionsDateLayout struct{ time.Time }

var _ gocsv.CSVUnmarshaller = &OCBCAccountTransactionsDateLayout{}

func (d *OCBCAccountTransactionsDateLayout) UnmarshalCSV(data []byte) (err error) {
	d.Time, err = time.Parse(OCBCAccountStatementDateLayout, string(data))
	if err != nil {
		err = fmt.Errorf("failed to parse ocbc date: %v", err)
		return
	}

	return
}
