package ocbc_test

import (
	"personal-finance/pkgs/dbs"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDBSCreditCardDate_UnmarshalCSV(t *testing.T) {
	t.Parallel()

	want := time.Date(2025, 12, 13, 0, 0, 0, 0, time.UTC)
	d := dbs.DBSCreditCardDate{}

	err := d.UnmarshalCSV([]byte("13 Dec 2025"))
	require.NoError(t, err)

	if d.Compare(want) != 0 {
		t.Fatalf("time mismatch - want %+v, have %+v", d.Time, want)
	}
}

func TestDBSCreditCardDate_UnmarshalCSVError(t *testing.T) {
	t.Parallel()

	d := dbs.DBSCreditCardDate{}

	err := d.UnmarshalCSV([]byte("13/12/2025")) // wrong format
	require.Error(t, err)
}
