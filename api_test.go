package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"time"
)

func TestAuth(t *testing.T) {
	router, _ := assembleRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth?userId=462a75d9-96a4-4ff4-81c8-54b7fd06fbb2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	// fmt.Printf("We got the following string: %s", w.Body.String())
}

func TestAuthAndRefresh(t *testing.T) {
	router, _ := assembleRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth?userId=462a75d9-96a4-4ff4-81c8-54b7fd06fbb2", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	tokens := tokenPair{}
	err := json.Unmarshal([]byte(w.Body.String()), &tokens)
	assert.Nil(t, err)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/refresh?Access=%s&Refresh=%s", tokens.Access, tokens.Refresh), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAuthAndRefreshFail(t *testing.T) {
	router, _ := assembleRouter()

	// To make access tokens different between tests
	time.Sleep(time.Second)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth?userId=462a75d9-96a4-4ff4-81c8-54b7fd06fbb2", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	tokens := tokenPair{}
	err := json.Unmarshal([]byte(w.Body.String()), &tokens)
	assert.Nil(t, err)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/refresh?Access=%s&Refresh=%s", tokens.Access, tokens.Refresh), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", fmt.Sprintf("/refresh?Access=%s&Refresh=%s", tokens.Access, tokens.Refresh), nil)
	router.ServeHTTP(w, req)
	assert.NotEqual(t, 200, w.Code)
}
