package main

import (
	// "crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type tokenPair struct {
	Access  string `json:"Access"`
	Refresh string `json:"Refresh"`
}

const defaultAPIPort = 8000

func accessFromParams(id string, ip string, jwtKey []byte) (string, error) {
	ttl := 10 * time.Second
	claims := jwt.MapClaims{
		"userIp": ip,
		"exp":    time.Now().UTC().Add(ttl).Unix(),
		"userID": id,
	}
	fmt.Println("CLAIMS:", claims)
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	return t.SignedString(jwtKey)
}

func refreshFromAccess(accessToken string, refreshKey string) string {
	hasher := sha512.New()
	hasher.Write([]byte(accessToken + refreshKey))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func authHandler(c *gin.Context, db *pgxpool.Pool, jwtKey []byte, refreshKey string) {
	id := c.Query("userId")
	// Make sure that user id is a valid uuid to prevent injections
	err := uuid.Validate(id)
	if err != nil {
		c.String(http.StatusInternalServerError, "The user id is not a valid uuid")
		return
	}

	_, err = getUser(db, id)
	if err != nil {
		c.String(http.StatusInternalServerError, "User with given id does not exist")
		return
	}

	accessToken, err := accessFromParams(id, c.ClientIP(), jwtKey)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error while generating the token: %v", err))
		return
	}

	refreshToken := refreshFromAccess(accessToken, refreshKey)

	c.IndentedJSON(http.StatusOK, tokenPair{
		Access:  accessToken,
		Refresh: refreshToken,
	})
}

func refreshHandler(c *gin.Context, jwtKey []byte, refreshKey string, db *pgxpool.Pool, salt string) {
	accessToken := c.Query("Access")
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	fmt.Println("CLAIMS WHEN EXPIRED", token.Claims)
	fmt.Printf("Error while parsing the token: %v\n", err)
	// Since expiration verification is done last, the token should still be valid even if it's expired
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error while parsing the token: %v", err))
		return
	}

	refreshToken := c.Query("Refresh")
	if refreshFromAccess(accessToken, refreshKey) != refreshToken {
		c.String(http.StatusInternalServerError, "Invalid refresh token for the access token")
		return
	}

	isRefreshUsed, err := isRefreshTokenExpired(db, refreshToken, salt)
	fmt.Printf("Refresh token %v: %s\n", isRefreshUsed, refreshToken)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error while checking the refresh token: %v", err))
		return
	} else if isRefreshUsed {
		c.String(http.StatusInternalServerError, "The refresh token has already been used")
		return
	}

	hexHash, err := tokenToHash(refreshToken, salt)
	fmt.Printf("Refresh token hash in hex: %s\n", hexHash)
	err = addNewExpiredRefreshToken(db, refreshToken, salt)

	claims, _ := token.Claims.(jwt.MapClaims)
	id, _ := claims["userID"]
	newAccessToken, err := accessFromParams(fmt.Sprint(id), c.ClientIP(), jwtKey)

	newRefreshToken := refreshFromAccess(newAccessToken, refreshKey)

	c.IndentedJSON(http.StatusOK, tokenPair{
		Access:  newAccessToken,
		Refresh: newRefreshToken,
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		os.Exit(1)
	}

	salt := os.Getenv("SALT")
	if salt == "" {
		fmt.Printf("The salt can't be empty")
		os.Exit(1)
	}

	// privateKeyBytes, err := os.ReadFile("jwtHS512.key")
	// if err != nil {
	// 	fmt.Printf("Error loading the keys: %v", err)
	// 	os.Exit(1)
	// }

	// privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	// if err != nil {
	// 	fmt.Printf("Error loading the keys: %v", err)
	// 	os.Exit(1)
	// }

	// publicKeyBytes, err := os.ReadFile("jwtHS512.key.pub")
	// if err != nil {
	// 	fmt.Printf("Error loading the keys: %v", err)
	// 	os.Exit(1)
	// }

	// publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	// if err != nil {
	// 	fmt.Printf("Error loading the keys: %v", err)
	// 	os.Exit(1)
	// }

	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		fmt.Printf("JWT secret cannot be empty")
		os.Exit(1)
	}

	refreshKey := os.Getenv("REFRESH_SECRET")
	if jwtKey == "" {
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
	getUser(db, "462a75d9-96a4-4ff4-81c8-54b7fd06fbb2")

	param, isPresent := os.LookupEnv("API_PORT")
	apiPort, err := strconv.Atoi(param)
	if !isPresent || err != nil {
		apiPort = defaultAPIPort
	}

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.TrustedPlatform = "X-Forwarded-For"

	router.GET("/auth", func(c *gin.Context) {
		// fmt.Printf("Request: %+v", c)
		authHandler(c, db, []byte(jwtKey), refreshKey)
	})

	router.GET("/refresh", func(c *gin.Context) {
		// fmt.Printf("Request: %+v", c)
		refreshHandler(c, []byte(jwtKey), refreshKey, db, salt)
	})

	// not localhost because docker
	router.Run(fmt.Sprintf("0.0.0.0:%d", apiPort))
}
