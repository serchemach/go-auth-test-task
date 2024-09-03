package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const defaultAPIPort = 8000

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	sender_address := os.Getenv("EMAIL_ADDRESS")
	if sender_address == "" {
		fmt.Printf("The sender email address can't be empty")
		os.Exit(1)
	}
	// fmt.Println(sender_address)

	sender_password := os.Getenv("EMAIL_PASS")
	if sender_password == "" {
		fmt.Printf("The sender email password can't be empty")
		os.Exit(1)
	}

	salt := os.Getenv("SALT")
	if salt == "" {
		fmt.Printf("The salt can't be empty")
		os.Exit(1)
	}

	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		fmt.Printf("JWT secret cannot be empty")
		os.Exit(1)
	}

	refreshKey := os.Getenv("REFRESH_SECRET")
	if refreshKey == "" {
		fmt.Printf("Refresh secret cannot be empty")
		os.Exit(1)
	}

	db, err := createConn()
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// for temporary testing
	// getUser(db, "462a75d9-96a4-4ff4-81c8-54b7fd06fbb2")

	param, isPresent := os.LookupEnv("API_PORT")
	apiPort, err := strconv.Atoi(param)
	if !isPresent || err != nil {
		apiPort = defaultAPIPort
	}

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.TrustedPlatform = "X-Forwarded-For"

	router.GET("/auth", func(c *gin.Context) {
		authHandler(c, db, []byte(jwtKey), refreshKey)
	})

	router.GET("/refresh", func(c *gin.Context) {
		refreshHandler(c, []byte(jwtKey), refreshKey, db, salt, Sender{email: sender_address, pass: sender_password})
	})

	// not localhost because docker
	router.Run(fmt.Sprintf("0.0.0.0:%d", apiPort))
}
