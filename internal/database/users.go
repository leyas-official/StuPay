package database

import (
	"context"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword string
}

type InsertUserArgs struct {
	Name     string `validate:"required"`
	Email    string `validate:"required"`
	Password string `validate:"required"`
}

func (d *DataService) InsertUser(ctx context.Context, args InsertUserArgs) (User, error) {
	stmt := `
		INSERT INTO users (
			name, email, hashed_password
		) VALUES (
			$1, $2, $3
		)
		RETURNING id, name, email;
	`
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(args.Password), bcrypt.DefaultCost)
	if err != nil {
		d.logger.Error(err.Error())
		return User{}, err
	}
	fmt.Print("i am retarded-2")
	var newUser User

	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		d.logger.Error(err.Error())
		return User{}, err
	}

	fmt.Print("i am retarded")

	err = tx.QueryRowContext(
		ctx,
		stmt,
		args.Name,
		args.Email,
		hashedPassword,
	).Scan(
		&newUser.ID,
		&newUser.Name,
		&newUser.Email,
	)
	if err != nil {
		d.logger.Error(err.Error())
		return User{}, err
	}

	err = tx.Commit()
	if err != nil {
		d.logger.Error(err.Error())
		return User{}, err
	}

	return newUser, err
}

type LoginUserArgs struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

func (d *DataService) LoginUser(ctx context.Context, args LoginUserArgs) (bool, error) {
	stmt := `
		SELECT hashed_password
		FROM users
		WHERE email = $1;
	`

	var storedHash string

	tx, err := d.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		d.logger.Error(err.Error())
		return false, err
	}

	err = tx.QueryRowContext(
		ctx,
		stmt,
		args.Email,
	).Scan(&storedHash)
	if err != nil {
		d.logger.Error(err.Error())
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		d.logger.Error(err.Error())
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(args.Password))
	if err != nil {
		return false, err
	}

	return true, nil
}
