package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/fajarilf/go-starter-api/internal/domain"
	"github.com/fajarilf/go-starter-api/internal/handler"
	"github.com/fajarilf/go-starter-api/internal/repository"
	"github.com/fajarilf/go-starter-api/internal/server"
	"github.com/fajarilf/go-starter-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type successResp struct {
	Status int      `json:"status"`
	Data   roomData `json:"data"`
}

type listResp struct {
	Status     int               `json:"status"`
	Data       []roomData        `json:"data"`
	Pagination domain.Pagination `json:"pagination"`
}

// setup spins up an httptest server wired to a freshly-truncated test DB.
// It skips the whole test when TEST_DATABASE_URL is not set.
func setup(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	loadEnv() // go test cwd is the package dir; .env lives at the project root

	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration tests")
	}

	if err := repository.Migrate(dbURL); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	if _, err := pool.Exec(context.Background(), "TRUNCATE rooms, users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	if _, err := pool.Exec(context.Background(),
		`INSERT INTO users (username, password_hash) VALUES ('admin', '$2a$10$3yQmNm0U93S.tWlm3nf7NuPki4JMX9ZN3zo.EE4hpbL.edqg57Cta')`); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	repo := repository.NewRoomRepository(pool)
	svc := service.NewRoomService(repo, validator.New())
	h := handler.NewRoomHandler(svc)
	userRepo := repository.NewUserRepository(pool)
	authSvc := service.NewAuthService(userRepo, validator.New(), config.Config{JWTSecret: "test-secret", JWTExpiryHours: 1})
	authH := handler.NewAuthHandler(authSvc)
	srv := server.New(config.Config{Port: "0"}, h, authH, authSvc)

	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		ts.Close()
		pool.Close()
	})

	return ts, pool
}

// do performs an HTTP request and returns the status code and raw body.
func do(t *testing.T, ts *httptest.Server, method, path, body string) (int, []byte) {
	return doReq(t, ts, method, path, body, "")
}

func doReq(t *testing.T, ts *httptest.Server, method, path, body, token string) (int, []byte) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, ts.URL+path, reader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return resp.StatusCode, raw
}

// mustLogin authenticates as admin and returns a JWT token.
func mustLogin(t *testing.T, ts *httptest.Server) string {
	t.Helper()

	status, raw := do(t, ts, http.MethodPost, "/api/login",
		`{"username":"admin","password":"admin123"}`)
	if status != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", status, raw)
	}

	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("login unmarshal: %v (body %s)", err, raw)
	}
	return resp.Data.Token
}

// mustCreate creates a valid room and returns its decoded data.
func mustCreate(t *testing.T, ts *httptest.Server, name, desc string) roomData {
	t.Helper()

	body := `{"name":"` + name + `","description":"` + desc + `"}`
	status, raw := do(t, ts, http.MethodPost, "/api/rooms", body)
	if status != http.StatusCreated {
		t.Fatalf("create %q: status = %d, body = %s", name, status, raw)
	}

	var got successResp
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("create unmarshal: %v (body %s)", err, raw)
	}
	return got.Data
}

// loadEnv walks up from the test's working directory to the module root
// (where go.mod lives) and loads the .env beside it, if any. Real env vars
// already set in the shell take precedence (godotenv does not overwrite them).
func loadEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			_ = godotenv.Load(filepath.Join(dir, ".env"))
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return // reached filesystem root
		}
		dir = parent
	}
}

// path builds the /api/rooms/{id} URL.
func path(id int) string { return "/api/rooms/" + strconv.Itoa(id) }

// str returns a string of n 'x' characters, for length-validation cases.
func str(n int) string { return string(bytes.Repeat([]byte("x"), n)) }
