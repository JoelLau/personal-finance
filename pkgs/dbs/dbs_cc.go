package dbs

import (
	"fmt"
	"time"

	gocsv "github.com/JoelLau/go-csv"
)

// Represents a single line in DBS's credit card statement (csv)
// NOTE: Row's DebitAmount OR CreditAmount must be 0. One of them MUST have a value.
type CreditCardItem struct {
	TransactionDate        DBSCreditCardDate `csv:"Transaction Date"`         // e.g. "22-Oct-25"
	TransactionPostingDate DBSCreditCardDate `csv:"Transaction Posting Date"` // e.g. "23-Oct-25"
	TransactionDescription string            `csv:"Transaction Description"`  // e.g. "SUPER SIMPLE           SINGAPORE     SG"
	PaymentType            string            `csv:"Payment Type"`             // e.g. ""Contactless", "Online/In-App Payment"
	TransactionStatus      string            `csv:"Transaction Status"`       // e.g. "Settled"
	DebitAmount            string            `csv:"Debit Amount"`             // e.g. "2.94"
	CreditAmount           string            `csv:"Credit Amount"`            // e.g. ""
}

const DBSCreditCardDateLayout = "02 Jan 2006"

type DBSCreditCardDate struct{ time.Time }

var _ gocsv.CSVUnmarshaller = &DBSCreditCardDate{}

func (d *DBSCreditCardDate) UnmarshalCSV(data []byte) (err error) {
	d.Time, err = time.Parse(DBSCreditCardDateLayout, string(data))
	if err != nil {
		err = fmt.Errorf("failed to parse dbs date: %v", err)
		return
	}

	return
}
