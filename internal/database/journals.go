package database

import (
	"context"
	"database/sql"
)

type InsertJournalArgs struct {
	Description string
}

type JournalId int

func (d *DataService) InsertJournal(ctx context.Context) (*sql.Stmt, error) {
	stmt := `
		INSERT INTO journals (
			description
		) VALUES (
			$1
		)
		RETURNING id;
	`

	prepStmt, err := d.db.PrepareContext(ctx, stmt)
	if err != nil {
		d.logger.Error(err.Error())
		return nil, err
	}

	return prepStmt, nil
}
