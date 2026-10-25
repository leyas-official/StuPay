package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"acm.stupay.com/internal/database"
)

type BorrowBookArgs struct {
	SenderID   int
	RecieverID int
	Amount     float32
	LibraryID  int
	ReturnTime time.Time
}

func (w *WalletService) BorrowBook(ctx context.Context, args BorrowBookArgs) (bool, error) {
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

	insertBookReservation, err := w.dataService.InsertBookReservation(ctx)
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

	tx.StmtContext(ctx, insertBookReservation).ExecContext(
		ctx,
		sql.Named(
			"account_id", args.SenderID,
		),
		sql.Named(
			"branch_id", args.LibraryID,
		),
		sql.Named(
			"return_time", args.ReturnTime,
		),
		sql.Named(
			"jounrnal_id", journalID,
		),
	)

	err = tx.Commit()
	if err != nil {
		w.logger.Error(err.Error())
		return false, err
	}

	return true, nil
}
