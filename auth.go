package main

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"time"
)

type tokenPair struct {
	Access  string `json:"Access"`
	Refresh string `json:"Refresh"`
}

const JWT_TTL = 10 * time.Second

func accessFromParams(id string, ip string, jwtKey []byte) (string, error) {
	claims := jwt.MapClaims{
		"userIp": ip,
		"exp":    time.Now().UTC().Add(JWT_TTL).Unix(),
		"userID": id,
	}
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

func refreshHandler(c *gin.Context, jwtKey []byte, refreshKey string, db *pgxpool.Pool, salt string, creds Sender) {
	accessToken := c.Query("Access")
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
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
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error while checking the refresh token: %v", err))
		return
	} else if isRefreshUsed {
		c.String(http.StatusInternalServerError, "The refresh token has already been used")
		return
	}

	err = addNewExpiredRefreshToken(db, refreshToken, salt)

	claims, _ := token.Claims.(jwt.MapClaims)
	id, _ := claims["userID"]
	newAccessToken, err := accessFromParams(fmt.Sprint(id), c.ClientIP(), jwtKey)

	prevIp := fmt.Sprint(claims["userIp"])
	if c.ClientIP() != prevIp {
		user, err := getUser(db, fmt.Sprint(id))
		if err != nil {
			fmt.Printf("Error while fetching the user for sending mail : %s\n", err)
		}

		err = sendMail(user.Name+" refresh operation notification", "Your access token was refreshed from IP ("+c.ClientIP()+") differing the one it was given to ("+prevIp+")", user.Email, creds)
		if err != nil {
			fmt.Printf("Error while sending mail: %s\n", err)
		}
	}

	newRefreshToken := refreshFromAccess(newAccessToken, refreshKey)

	c.IndentedJSON(http.StatusOK, tokenPair{
		Access:  newAccessToken,
		Refresh: newRefreshToken,
	})
}
