package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type AccountingRepository interface {
	CreateExpense(context.Context, CreateExpenseParams) error
	ListExpenses(context.Context) ([]Expense, error)
}

// NOTE: use slice index as ID field (hidden from public)
type InMemoryAccountingRepository struct {
	accounts       []LedgerAccount
	journalEntries []CreateJournalEntryParams
	postings       []CreatePostingParams
}

var _ AccountingRepository = &InMemoryAccountingRepository{}

func NewInMemoryAccountingRepository() *InMemoryAccountingRepository {
	return &InMemoryAccountingRepository{
		accounts:       []LedgerAccount{},
		journalEntries: []CreateJournalEntryParams{},
		postings:       []CreatePostingParams{},
	}
}

// naive approach - should be in db transaction
func (repo *InMemoryAccountingRepository) CreateExpense(ctx context.Context, param CreateExpenseParams) error {
	journalEntryID, err := repo.createJournalEntry(ctx, CreateJournalEntryParams{
		Name:        param.Name,
		Description: param.Description,
		Date:        param.TransactedAt,
	})
	if err != nil {
		return fmt.Errorf("error creating journal entry for expense creation: %+v", err)
	}

	_, err = repo.createPosting(ctx, CreatePostingParams{
		Name:             param.Name,
		Description:      param.Description,
		CreditInMicroSGD: param.CreditInMicroSGD,
		DebitInMicroSGD:  param.DebitInMicroSGD,

		JournalEntryID: journalEntryID,
		AccountID:      -1,
	})
	if err != nil {
		return fmt.Errorf("error creating posting for expense creation: %+v", err)
	}

	return nil
}

func (repo *InMemoryAccountingRepository) createJournalEntry(_ context.Context, param CreateJournalEntryParams) (journalEntryID int64, err error) {
	journalEntryID = int64(len(repo.journalEntries))
	repo.journalEntries = append(repo.journalEntries, param)

	return journalEntryID, nil
}

func (repo *InMemoryAccountingRepository) createPosting(_ context.Context, param CreatePostingParams) (postingID int64, err error) {
	postingID = int64(len(repo.postings))
	repo.postings = append(repo.postings, param)

	return postingID, nil
}

func (repo *InMemoryAccountingRepository) ListExpenses(context.Context) ([]Expense, error) {
	expenseMap := make(map[int64]Expense)
	for idx, param := range repo.journalEntries {
		journalEntryID := int64(idx)
		expenseMap[journalEntryID] = Expense{
			ID:               journalEntryID,
			Name:             param.Name,
			Description:      param.Description,
			TransactedAt:     param.Date,
			CreditInMicroSGD: 0, // will be populated / computed later
			DebitInMicroSGD:  0, // will be populated / computed later
			journalEntryID:   journalEntryID,
			postingIDs:       make(map[int64]bool),
		}
	}

	for idx, param := range repo.postings {
		postingID := int64(idx)
		journalEntryID := int64(param.JournalEntryID)

		expense, ok := expenseMap[journalEntryID]
		if !ok {
			return nil, fmt.Errorf("no expense with id %d", journalEntryID)
		}

		expense.postingIDs[postingID] = true
		expense.CreditInMicroSGD += param.CreditInMicroSGD
		expense.DebitInMicroSGD += param.DebitInMicroSGD

		expenseMap[journalEntryID] = expense
		slog.Info("finished addition", slog.Any("updated expense", expenseMap[journalEntryID]))
	}

	expenses := make([]Expense, len(expenseMap))
	idx := 0
	for _, v := range expenseMap {
		expenses[idx] = v
		idx++
	}

	return expenses, nil
}

type CreateExpenseParams struct {
	Name             string
	Description      string
	TransactedAt     time.Time
	CreditInMicroSGD int64
	DebitInMicroSGD  int64
}

type Expense struct {
	ID               int64
	Name             string
	Description      string
	TransactedAt     time.Time
	CreditInMicroSGD int64
	DebitInMicroSGD  int64

	journalEntryID int64
	postingIDs     map[int64]bool
}

type LedgerAccount struct {
	ID          int64
	Name        string
	Description string
	ParentID    string
}

type CreateJournalEntryParams struct {
	Name        string
	Description string
	Date        time.Time
}

type JournalEntry struct {
	ID          int64
	Name        string
	Description string
	Date        time.Time
}

type CreatePostingParams struct {
	Name             string
	Description      string
	CreditInMicroSGD int64
	DebitInMicroSGD  int64

	AccountID      int64
	JournalEntryID int64
}

type Posting struct {
	ID               int64
	Name             string
	Description      string
	CreditInMicroSGD int64
	DebitInMicroSGD  int64

	AccountID      int64
	JournalEntryID int64
}
