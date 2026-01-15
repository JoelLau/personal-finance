package domains

import (
	"io"
	"time"
)

type AccountingDomain interface {
	ParseExpenseCSV(io.Reader) ([]Expense, error)
}

type Expense struct {
	TransactedOn           time.Time
	Description            string
	CategoryID             int64
	DebitAmountInMicroSGD  int64
	CreditAmountInMicroSGD int64
}
