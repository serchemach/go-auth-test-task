package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingRoute(t *testing.T) {
	router, db := assembleRouter()
	defer db.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/auth?userId=462a75d9-96a4-4ff4-81c8-54b7fd06fbb2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	fmt.Printf("We got the following string: %s", w.Body.String())
}
