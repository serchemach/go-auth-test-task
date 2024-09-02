package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	// "github.com/google/uuid"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"strconv"
)

type tokenPair struct {
	Access  string `json:"Access"`
	Refresh string `json:"Refresh"`
}

const defaultAPIPort = 8000

func authHandler(c *gin.Context) {

	c.IndentedJSON(http.StatusOK, tokenPair{
		Access:  "123123",
		Refresh: "dsioaoidjf",
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	db, err := createConn()
	if err != nil {
		fmt.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	getUser(db, "462a75d9-96a4-4ff4-81c8-54b7fd06fbb2")

	param, isPresent := os.LookupEnv("API_PORT")
	apiPort, err := strconv.Atoi(param)
	if !isPresent || err != nil {
		apiPort = defaultAPIPort
	}

	router := gin.Default()
	router.GET("/auth", authHandler)

	// not localhost because docker
	router.Run(fmt.Sprintf("0.0.0.0:%d", apiPort))
}
