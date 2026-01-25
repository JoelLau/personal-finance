package domain

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type AccountingRepository interface {
	CreateExpense(context.Context, CreateExpenseParams) error
	CreateIncome(context.Context, CreateIncomeParams) error
	ListTransactions(context.Context) ([]Expense, error)
}

const (
	// ---
	// Asset Accounts (What you OWN)
	// ---
	// These have a Normal Debit balance.
	// When you receive money, you Debit these.
	AccountID_Asset_BankAccount        = 1000 // Your primary spending account.
	AccountID_Asset_CashOnHand         = 1100 // The physical cash in your wallet.
	AccountID_Asset_Investments        = 1200 // Brokerage accounts, 401k, or stocks.
	AccountID_Asset_AccountsReceivable = 1300 // Money people owe you (e.g., a friend you lent $20 to)

	// ---
	// 2. Liability Accounts (What you OWE)
	// ---
	// These have a Normal Credit balance.
	// When you borrow, you Credit these.
	AccountID_Liability_CreditCard      = 2100 // Your outstanding balance on a specific card.
	AccountID_Liability_StudentLoan     = 2200 // Long-term education debt.
	AccountID_Liability_MortgageCarLoan = 2300 // Large installment loans.
	AccountID_Liability_PersonalLoans   = 2400 // Money you owe to friends or family.

	// ---
	// 3. Income Accounts (Where money COMES FROM)
	// ---
	// These have a Normal Credit balance.
	// When you earn, you Credit these
	AccountID_Income_SalaryWages      = 3100 // Your primary paycheck.
	AccountID_Income_InterestIncome   = 3200 // Dividends or interest from bank accounts.
	AccountID_Income_GiftsReceived    = 3300 // Money received for birthdays or holidays.
	AccountID_Income_SideHustleIncome = 3400 // Freelance or gig economy earnings.
	AccountID_Income_TaxRefunds       = 3500 // Money returned from the government.

	// ---
	// 4. Expense Accounts (Where money GOES)
	// ---
	// These have a Normal Debit balance.
	// When you spend, you Debit these.
	AccountID_Expense_Housing        = 4100 // Rent or mortgage interest.
	AccountID_Expense_Groceries      = 4200 // Food for home.
	AccountID_Expense_DiningOut      = 4300 // Restaurants, coffee, and takeout.
	AccountID_Expense_Utilities      = 4400 // Electricity, water, internet, and phone.
	AccountID_Expense_Transportation = 4500 // Gas, public transit, or car maintenance.
	AccountID_Expense_Subscriptions  = 4600 // Netflix, Spotify, gym memberships.
	AccountID_Expense_PersonalCare   = 4700 // Haircuts, toiletries, and clothing.

	// ---
	// 5. Equity Accounts (Your "Net Worth")
	// ---
	// These have a Normal Credit balance.
	AccountID_Equity_OpeningBalanceEquity = 5100 // A special account used only when you first start your books to record the initial money you have in your accounts.
	AccountID_Equity_RetainedEarnings     = 5200 // (Automated by most software) This represents the total "profit" or savings you’ve accumulated over time.
)

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

// WARN: NOT atomic
func (repo *InMemoryAccountingRepository) CreateIncome(ctx context.Context, param CreateIncomeParams) error {
	journalEntryID, err := repo.createJournalEntry(ctx, CreateJournalEntryParams{
		Name:        param.Name,
		Description: param.Description,
		Date:        param.TransactedAt,
	})
	if err != nil {
		return fmt.Errorf("error creating journal entry for income creation: %+v", err)
	}

	_, err = repo.createPosting(ctx, CreatePostingParams{
		Name:             param.Name,
		Description:      param.Description,
		CreditInMicroSGD: param.CreditInMicroSGD,
		DebitInMicroSGD:  param.DebitInMicroSGD,

		JournalEntryID: journalEntryID,
		AccountID:      AccountID_Income_SalaryWages,
	})
	if err != nil {
		return fmt.Errorf("error creating posting for income creation: %+v", err)
	}

	return nil
}

// WARN: NOT atomic
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
		AccountID:      AccountID_Asset_BankAccount,
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

func (repo *InMemoryAccountingRepository) ListTransactions(context.Context) ([]Expense, error) {
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

type CreateIncomeParams = CreateExpenseParams
type Income = Expense

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
