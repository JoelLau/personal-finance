package main_test

import (
	main "personal-finance"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetLastMonth(t *testing.T) {
	t.Parallel()

	n := NowerStub{val: time.Date(2025, 12, 13, 0, 0, 0, 0, time.UTC)}
	have := main.GetLastMonthYYYYMM(n)

	require.Equal(t, "2025-11", have)
}

type NowerStub struct {
	val time.Time
}

func (n NowerStub) Now() time.Time {
	return n.val
}
