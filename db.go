package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/scrypt"
	"os"
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
	query := fmt.Sprintf("select * from auth_scheme.user where id = '%s'", id)
	rows, _ := db.Query(context.Background(), query)
	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		return User{Id: "", Name: "", Email: ""}, err
	}

	// for _, p := range products {
	// 	fmt.Printf("%s: %s, %s\n", p.Id, p.Email, p.Name)
	// }

	if len(products) > 0 {
		return products[0], nil
	}

	return User{Id: "", Name: "", Email: ""}, &NoUserError{}
}

// bcrypt library is not used because you can't generate hashes with fixed salt, which makes hashes different each time you generate them
// in order to match hashed refresh tokens in the database, we need to keep the salt fixed
func tokenToHash(token string, salt string) (string, error) {
	bytes, err := scrypt.Key([]byte(token), []byte(salt), 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func addNewExpiredRefreshToken(db *pgxpool.Pool, refreshToken string, salt string) error {
	bytesHex, err := tokenToHash(refreshToken, salt)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("insert into auth_scheme.expired_refresh values (decode('%s', 'hex'));", bytesHex)
	// fmt.Printf(query)
	_, err = db.Query(context.Background(), query)

	return err
}

func isRefreshTokenExpired(db *pgxpool.Pool, refreshToken string, salt string) (bool, error) {
	bytesHex, err := tokenToHash(refreshToken, salt)
	if err != nil {
		return false, err
	}

	query := fmt.Sprintf("select count(*) from auth_scheme.expired_refresh where token = decode('%s', 'hex');", bytesHex)
	// fmt.Printf(query)
	var numberOfVals int
	err = db.QueryRow(context.Background(), query).Scan(&numberOfVals)
	if err != nil {
		return false, err
	}

	if numberOfVals > 0 {
		return true, nil
	}

	return false, nil
}
