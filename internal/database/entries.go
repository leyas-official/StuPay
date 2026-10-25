package database

import (
	"context"
	"database/sql"
)

type InsertEntryArgs struct {
	Amount    float64
	CreditID  int
	DebitID   int
	JournalId JournalId
}

func (d *DataService) InsertEntry(ctx context.Context) (*sql.Stmt, error) {
	stmt := `
		INSERT INTO entries (
			amount, credit, debit, journal_id
		) VALUES (
			$1, $2, $3, $4
		)
	`

	prepStmt, err := d.db.PrepareContext(ctx, stmt)
	if err != nil {
		d.logger.Error(err.Error())
		return nil, err
	}

	return prepStmt, nil
}
