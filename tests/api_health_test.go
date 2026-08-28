package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHealthEndpoint 測試健康檢查端點
func TestHealthEndpoint(t *testing.T) {
	w := makeRequest("GET", "/health", nil, "")

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "ok", response["status"])
}
