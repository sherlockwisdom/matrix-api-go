package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var username = "sherlock_" + strconv.FormatInt(time.Now().UnixMilli(), 10)

func TestApiCreate_ValidRequest(t *testing.T) {
	router := setupRouter()

	payload := ClientJsonRequest{
		Username: username,
		Password: "testpass",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/create", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, username, response["username"])
	assert.Equal(t, "created", response["status"])
}

func TestApiCreate_InvalidPayload(t *testing.T) {
	router := setupRouter()

	// Missing required fields
	req, _ := http.NewRequest("POST", "/create", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Username and password are required")
}

func TestApiLogin_ValidCredentials(t *testing.T) {
	router := setupRouter()

	// Pre-store user in DB if needed, or assume test DB state
	payload := ClientJsonRequest{
		Username: username,
		Password: "testpass",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]string
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, username, response["username"])
	assert.Equal(t, "logged in", response["status"])
}

func TestApiLogin_MissingFields(t *testing.T) {
	router := setupRouter()

	body := []byte(`{"username": ""}`) // missing password
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Username and password are required")
}

func TestApiLogin_InvalidCredentials(t *testing.T) {
	router := setupRouter()

	payload := ClientJsonRequest{
		Username: "wronguser",
		Password: "wrongpass",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Login failed", response["error"])
}

// Extract router setup to reuse across tests
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/create", ApiCreate)
	router.POST("/login", ApiLogin)
	return router
}
