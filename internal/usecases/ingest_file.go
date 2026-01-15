package usecases

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type IngestFileUseCase struct {
	parsers []any // TODO: create parser interface
}

func NewIngestFileUseCase() *IngestFileUseCase {
	return &IngestFileUseCase{}
}

// TODO: wip
func (uc *IngestFileUseCase) Execute(
	ctx context.Context,
	filepath string,
	monthfilter string,
) error {
	slog.Info("Execute", slog.String("filepath", filepath))

	// TODO: open file
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	fileContent := string(fileBytes)
	slog.InfoContext(ctx, "file readed", slog.Any("file contents", fileContent))

	// TODO: loop through parsers, attempt data transformation
	// TODO: store contents somewhere(?)

	return nil
}
