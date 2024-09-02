package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NoUserError struct{}

func (m *NoUserError) Error() string {
	return "No such users found"
}

type User struct {
	Id    string `db:"id"`
	Email string `db:"email"`
	Name  string `db:"name"`
}

func createConn() (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@127.0.0.1:5432/sample", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))
	conn, err := pgxpool.New(context.Background(), connString)

	return conn, err
}

func getUser(db *pgxpool.Pool, id string) (User, error) {
	rows, _ := db.Query(context.Background(),
		fmt.Sprintf("select * from auth_scheme.user where id = '%s'", id))
	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		return User{Id: "", Name: "", Email: ""}, err
	}

	for _, p := range products {
		fmt.Printf("%s: %s, %s\n", p.Id, p.Email, p.Name)
	}

	if len(products) > 0 {
		return products[0], nil
	}

	return User{Id: "", Name: "", Email: ""}, &NoUserError{}
}
