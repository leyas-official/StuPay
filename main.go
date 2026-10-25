package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"acm.stupay.com/server"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load(".env", ".envrc")
	db := openDB()
	defer db.Close()

	handler := log.New(os.Stderr)
	log.SetReportCaller(true)
	logger := slog.New(handler)

	s := server.InitServer(server.NewServerArgs{
		Logger: logger,
		DB:     db,
	})

	s.RegisterRoutes()
	s.Run()
}

func openDB() *sql.DB {
	dbString := os.Getenv("DB_STRING")

	db, err := sql.Open("pgx", dbString)
	if err != nil {
		panic("Could not open db")
	}

	err = db.Ping()
	if err != nil {
		panic("Could not ping db")
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	return db
}
