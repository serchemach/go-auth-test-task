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
		fmt.Println("Error loading .env file: ", err)
	}

	param, isPresent := os.LookupEnv("API_PORT")
	apiPort, err := strconv.Atoi(param)
	if !isPresent || err != nil {
		apiPort = defaultAPIPort
	}

	router := gin.Default()
	router.GET("/auth", authHandler)

	router.Run(fmt.Sprintf("0.0.0.0:%d", apiPort))
}
