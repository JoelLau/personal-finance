package main_test

import (
	"log/slog"
	main "personal-finance/apps/ingest-ocbc"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	sb := new(strings.Builder)
	slogger := slog.New(slog.NewTextHandler(sb, &slog.HandlerOptions{}))
	args := []string{"ingest", "--file", "../../tests/testdata/ocbc.csv"}

	cmd := main.NewIngestOCBCAccountStatemtnCSVCommand(slogger)
	err := cmd.Run(ctx, args)
	require.NoError(t, err)
}
