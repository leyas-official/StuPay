package database

import (
	"context"
	"database/sql"
)

func (d *DataService) InsertBookReservation(ctx context.Context) (*sql.Stmt, error) {
	stmt := `
		INSERT INTO book_reservation (
			account_id, branch_id, return_time, journal_id
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
