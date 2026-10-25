package database

import (
	"context"
)

type GetUserBalanceArgs struct {
	UserID int
}

func (d *DataService) GetUserBalance(ctx context.Context, args GetUserBalanceArgs) (float32, error) {
	stmt := `
		SELECT balance
		FROM account_balances
		WHERE id = (
			SELECT id
			FROM accounts
			WHERE owner_user_id = $1
		);
	`

	var balance float32

	err := d.db.QueryRowContext(
		ctx,
		stmt,
		args.UserID,
	).Scan(&balance)
	if err != nil {
		d.logger.Error(err.Error())
		return 0, err
	}

	return balance, nil
}
