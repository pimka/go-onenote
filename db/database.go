package db

import (
	"context"
	pgx "github.com/jackc/pgx/v4"
	"log"
)

type Database struct {
	*pgx.Conn
}

func Connect(url_conn string) (*Database, error) {
	conn, err := pgx.Connect(context.Background(), url_conn)
	if err != nil {
		log.Println("Cannot connect to DB")
		return nil, err
	}

	return &Database{conn}, nil
}
