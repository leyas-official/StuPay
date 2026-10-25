package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"acm.stupay.com/internal/database"
)

type WalletService struct {
	logger      *slog.Logger
	dataService *database.DataService
}

type NewWalletServiceArgs struct {
	Logger   *slog.Logger
	Database *database.DataService
}

type TransferMoneyArgs struct {
	SenderID   int
	RecieverID int
	Amount     float32
}

func NewWalletService(args NewWalletServiceArgs) *WalletService {
	if args.Logger == nil {
		args.Logger = slog.Default()
	}

	if args.Database == nil {
		args.Logger.Error("missing database")
	}

	return &WalletService{
		logger:      args.Logger,
		dataService: args.Database,
	}
}

// TODO: fix this by literally not doing this and using a tx in the database layer instead
func (w *WalletService) TransferMoney(ctx context.Context, args TransferMoneyArgs) (bool, error) {
	// First make sure the sender has enough funds
	balance, err := w.dataService.GetUserBalance(
		ctx,
		database.GetUserBalanceArgs{
			UserID: args.SenderID,
		},
	)
	if err != nil {
		w.logger.Error(err.Error())
		return false, err
	}

	print("\n", balance, "\n")

	if balance < args.Amount {
		return false, errors.New("Not Enough Balance")
	}

	insertJournalStmt, err := w.dataService.InsertJournal(ctx)
	if err != nil {
		w.logger.Error(err.Error())
		return false, err
	}
	var journalID database.JournalId

	insertEntryStmt, err := w.dataService.InsertEntry(ctx)
	if err != nil {
		w.logger.Error(err.Error())
		return false, err
	}

	tx, err := w.dataService.BeginTransaction(ctx, &sql.TxOptions{})
	if err != nil {
		w.logger.Error(err.Error())
		return false, err
	}

	tx.StmtContext(ctx, insertJournalStmt).QueryRowContext(
		ctx,
		fmt.Sprintf(
			"User: %d sent %f LYD to user: %d.",
			args.SenderID,
			args.Amount,
			args.RecieverID,
		),
	).Scan(
		&journalID,
	)

	tx.StmtContext(ctx, insertEntryStmt).ExecContext(
		ctx,
		sql.Named(
			"amount", args.Amount,
		),
		sql.Named(
			"credit", args.RecieverID,
		),
		sql.Named(
			"debit", args.SenderID,
		),
		sql.Named(
			"journal_id", journalID,
		),
	)

	err = tx.Commit()
	if err != nil {
		w.logger.Error(err.Error())
		return false, err
	}

	return true, nil
}
