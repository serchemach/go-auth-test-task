package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

type User struct {
	Id    string `db:"id"`
	Email string `db:"email"`
	Name  string `db:"name"`
}

func createConn() (*pgx.Conn, error) {
	connString := fmt.Sprintf("postgres://%s:%s@postgres:5432/sample", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"))
	conn, err := pgx.Connect(context.Background(), connString)

	return conn, err
}

func getUser(db *pgx.Conn, id string) (User, error) {
	rows, _ := db.Query(context.Background(),
		fmt.Sprintf("select * from auth_scheme.user where id = '%s'", id))
	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
	}

	for _, p := range products {
		fmt.Printf("%s: %s, %s\n", p.Id, p.Email, p.Name)
	}

	if len(products) > 0 {
		return products[0], nil
	}

	return User{Id: "", Name: "", Email: ""}, err
}
