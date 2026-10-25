package database

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
)

type DataService struct {
	db        *sql.DB
	logger    *slog.Logger
	validator *validator.Validate
}

type NewDataServiceArgs struct {
	Db        *sql.DB
	Logger    *slog.Logger
	Validator *validator.Validate
}

func NewDataService(args NewDataServiceArgs) *DataService {
	if args.Logger == nil {
		args.Logger = slog.Default()
	}
	if args.Db == nil {
		args.Logger.Error("Did not provide valid db connection.")
		os.Exit(1)
	}
	if args.Validator == nil {
		args.Logger.Error("Did not provide valid validator.")
		os.Exit(1)
	}
	return &DataService{
		db:        args.Db,
		logger:    args.Logger,
		validator: args.Validator,
	}
}

func (d *DataService) BeginTransaction(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}
