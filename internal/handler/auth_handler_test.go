package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRegister(t *testing.T) {
	ts, _ := setup(t)

	t.Run("valid", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodPost, "/api/register",
			`{"username":"regtest1","password":"secret123"}`)
		if status != http.StatusCreated {
			t.Fatalf("status = %d, body = %s", status, raw)
		}

		var got struct {
			Status int `json:"status"`
			Data   struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Data.ID == 0 || got.Data.Username != "regtest1" {
			t.Errorf("unexpected: %+v", got.Data)
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/register",
			`{"username":"regtest1","password":"secret123"}`)
		if status != http.StatusConflict {
			t.Errorf("status = %d, want 409", status)
		}
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
			if status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", status)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	ts, _ := setup(t)

	t.Run("valid", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodPost, "/api/login",
			`{"username":"admin","password":"admin123"}`)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %s", status, raw)
		}

		var got struct {
			Status int `json:"status"`
			Data   struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Data.Token == "" {
			t.Error("token is empty")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/login",
			`{"username":"admin","password":"wrong"}`)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/login",
			`{"username":"nobody","password":"secret123"}`)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/login", `{`)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/login", `{}`)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})
}

func TestLogout(t *testing.T) {
	ts, _ := setup(t)
	token := mustLogin(t, ts)

	t.Run("valid", func(t *testing.T) {
		status, raw := doReq(t, ts, http.MethodPost, "/api/logout", "", token)
		if status != http.StatusOK {
			t.Fatalf("status = %d, body = %s", status, raw)
		}
	})

	t.Run("no token", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPost, "/api/logout", "")
		if status != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", status)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		status, _ := doReq(t, ts, http.MethodPost, "/api/logout", "", "not.a.real.token")
		if status != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", status)
		}
	})

	t.Run("already revoked", func(t *testing.T) {
		status, _ := doReq(t, ts, http.MethodPost, "/api/logout", "", token)
		if status != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", status)
		}
	})
}
