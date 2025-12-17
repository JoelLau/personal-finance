package main_test

import (
	main "personal-finance"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDBSCreditCardDate_UnmarshalCSV(t *testing.T) {
	t.Parallel()

	want := time.Date(2025, 12, 13, 0, 0, 0, 0, time.UTC)
	d := main.DBSCreditCardDate{}

	err := d.UnmarshalCSV([]byte("13-Dec-2025"))
	require.NoError(t, err)

	if d.Compare(want) != 0 {
		t.Fatalf("time mismatch - want %+v, have %+v", d.Time, want)
	}
}

func TestDBSCreditCardDate_UnmarshalCSVError(t *testing.T) {
	t.Parallel()

	d := main.DBSCreditCardDate{}

	err := d.UnmarshalCSV([]byte("13/12/2025")) // wrong format
	require.Error(t, err)
}
