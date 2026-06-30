package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	ts, _ := setup(t)

	t.Run("valid", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodPost, "/api/register",
			`{"username":"regtest1","password":"secret123"}`)
		require.Equal(t, http.StatusCreated, status, "body: %s", raw)

		var got struct {
			Status int `json:"status"`
			Data   struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.NotZero(t, got.Data.ID)
		assert.Equal(t, "regtest1", got.Data.Username)
	})

	t.Run("duplicate username", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/register",
			`{"username":"regtest1","password":"secret123"}`)
		assert.Equal(t, http.StatusConflict, status)
	})

	bad := []struct {
		name string
		body string
	}{
		{"malformed json", `{`},
		{"empty body", `{}`},
		{"username too short", `{"username":"ab","password":"secret123"}`},
		{"password too short", `{"username":"validuser","password":"12345"}`},
		{"missing password", `{"username":"validuser"}`},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			status, _ := do(t, ts, http.MethodPost, "/api/register", tc.body)
			assert.Equal(t, http.StatusBadRequest, status)
		})
	}
}

func TestLogin(t *testing.T) {
	ts, _ := setup(t)

	t.Run("valid", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodPost, "/api/login",
			`{"username":"admin","password":"admin123"}`)
		require.Equal(t, http.StatusOK, status, "body: %s", raw)

		var got struct {
			Status int `json:"status"`
			Data   struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.NotEmpty(t, got.Data.Token)
	})

	t.Run("wrong password", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/login",
			`{"username":"admin","password":"wrong"}`)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("non-existent user", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/login",
			`{"username":"nobody","password":"secret123"}`)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("malformed json", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/login", `{`)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("empty body", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/login", `{}`)
		assert.Equal(t, http.StatusBadRequest, status)
	})
}

func TestLogout(t *testing.T) {
	ts, _ := setup(t)
	token := mustLogin(t, ts)

	t.Run("valid", func(t *testing.T) {
		status, raw := doReq(t, ts, http.MethodPost, "/api/logout", "", token)
		assert.Equal(t, http.StatusOK, status, "body: %s", raw)
	})

	t.Run("no token", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/logout", "")
		assert.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("invalid token", func(t *testing.T) {
		status, _ := doReq(t, ts, http.MethodPost, "/api/logout", "", "not.a.real.token")
		assert.Equal(t, http.StatusUnauthorized, status)
	})

	t.Run("already revoked", func(t *testing.T) {
		status, _ := doReq(t, ts, http.MethodPost, "/api/logout", "", token)
		assert.Equal(t, http.StatusUnauthorized, status)
	})
}
