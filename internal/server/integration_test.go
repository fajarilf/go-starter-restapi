package server_test

// Integration tests: they drive the real chi router -> handler -> service ->
// repository -> pgx stack against a real Postgres.
//
// Set TEST_DATABASE_URL to run them, e.g.:
//
//	TEST_DATABASE_URL=postgres://user:pass@localhost:5432/app_test?sslmode=disable go test ./internal/server -v
//
// Without it the suite skips (so `go test ./...` stays green without a DB).
// The target database is migrated and TRUNCATEd between tests, so point it at
// a throwaway DB, never a real one.

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

type roomData struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

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

	if _, err := pool.Exec(context.Background(), "TRUNCATE rooms RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	repo := repository.NewRoomRepository(pool)
	svc := service.NewRoomService(repo, validator.New())
	h := handler.NewRoomHandler(svc)
	srv := server.New(config.Config{Port: "0"}, h)

	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(func() {
		ts.Close()
		pool.Close()
	})

	return ts, pool
}

// do performs an HTTP request and returns the status code and raw body.
func do(t *testing.T, ts *httptest.Server, method, path, body string) (int, []byte) {
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

func TestHealthz(t *testing.T) {
	ts, _ := setup(t)

	status, raw := do(t, ts, http.MethodGet, "/api/healthz", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	if string(raw) != "ok" {
		t.Fatalf("body = %q, want ok", raw)
	}
}

func TestCreate(t *testing.T) {
	ts, _ := setup(t)

	t.Run("valid", func(t *testing.T) {
		got := mustCreate(t, ts, "Conference A", "Third floor")
		if got.ID == 0 {
			t.Errorf("id not assigned: %+v", got)
		}
		if got.Name != "Conference A" || got.Description != "Third floor" {
			t.Errorf("unexpected data: %+v", got)
		}
	})

	bad := []struct {
		name string
		body string
	}{
		{"malformed json", `{`},
		{"unknown field", `{"name":"Room A","description":"x","extra":1}`},
		{"empty body", `{}`},
		{"name too short", `{"name":"ab","description":"x"}`},
		{"name too long", `{"name":"` + str(51) + `","description":"x"}`},
		{"missing description", `{"name":"Valid Name"}`},
		{"empty description", `{"name":"Valid Name","description":""}`},
		{"wrong type", `{"name":123,"description":"x"}`},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			status, raw := do(t, ts, http.MethodPost, "/api/rooms", tc.body)
			if status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400 (body %s)", status, raw)
			}
		})
	}
}

func TestGetById(t *testing.T) {
	ts, _ := setup(t)
	created := mustCreate(t, ts, "Room A", "desc")

	t.Run("found", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodGet, path(created.ID), "")
		if status != http.StatusOK {
			t.Fatalf("status = %d (body %s)", status, raw)
		}
		var got successResp
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Data.ID != created.ID || got.Data.Name != "Room A" {
			t.Errorf("unexpected: %+v", got.Data)
		}
	})

	t.Run("not found", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodGet, "/api/rooms/999999", "")
		if status != http.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})

	t.Run("non-numeric id", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodGet, "/api/rooms/abc", "")
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})
}

func TestUpdate(t *testing.T) {
	ts, _ := setup(t)
	created := mustCreate(t, ts, "Old Name", "old desc")

	t.Run("valid", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodPut, path(created.ID),
			`{"name":"New Name","description":"new desc"}`)
		if status != http.StatusOK {
			t.Fatalf("status = %d (body %s)", status, raw)
		}
		var got successResp
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got.Data.Name != "New Name" || got.Data.Description != "new desc" {
			t.Errorf("not updated: %+v", got.Data)
		}
	})

	t.Run("not found", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPut, "/api/rooms/999999",
			`{"name":"Whatever","description":"x"}`)
		if status != http.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})

	t.Run("non-numeric id", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPut, "/api/rooms/abc",
			`{"name":"Whatever","description":"x"}`)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("validation fail", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPut, path(created.ID),
			`{"name":"ab","description":"x"}`)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPut, path(created.ID), `{`)
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})
}

func TestDelete(t *testing.T) {
	ts, _ := setup(t)
	created := mustCreate(t, ts, "To Delete", "desc")

	t.Run("existing", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodDelete, path(created.ID), "")
		if status != http.StatusOK {
			t.Fatalf("status = %d (body %s)", status, raw)
		}
		if !bytes.Contains(raw, []byte("room deleted")) {
			t.Errorf("body = %s, want message", raw)
		}
		// gone afterwards
		status, _ = do(t, ts, http.MethodGet, path(created.ID), "")
		if status != http.StatusNotFound {
			t.Errorf("after delete status = %d, want 404", status)
		}
	})

	t.Run("not found", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodDelete, "/api/rooms/999999", "")
		if status != http.StatusNotFound {
			t.Errorf("status = %d, want 404", status)
		}
	})

	t.Run("non-numeric id", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodDelete, "/api/rooms/abc", "")
		if status != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", status)
		}
	})
}

func TestList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		ts, _ := setup(t)
		status, raw := do(t, ts, http.MethodGet, "/api/rooms", "")
		if status != http.StatusOK {
			t.Fatalf("status = %d (body %s)", status, raw)
		}
		var got listResp
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(got.Data) != 0 || got.Pagination.Total != 0 {
			t.Errorf("expected empty, got %+v", got)
		}
		if got.Pagination.HasPrev || got.Pagination.HasNext {
			t.Errorf("expected no prev/next, got %+v", got.Pagination)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		ts, _ := setup(t)
		for i := 0; i < 5; i++ {
			mustCreate(t, ts, "Room "+string(rune('A'+i)), "d")
		}

		// page 1, limit 2 -> 2 items, has_next true, has_prev false, 3 pages
		status, raw := do(t, ts, http.MethodGet, "/api/rooms?page=1&limit=2", "")
		if status != http.StatusOK {
			t.Fatalf("status = %d (body %s)", status, raw)
		}
		var p1 listResp
		json.Unmarshal(raw, &p1)
		if len(p1.Data) != 2 {
			t.Errorf("page1 len = %d, want 2", len(p1.Data))
		}
		if p1.Pagination.Total != 5 || p1.Pagination.TotalPages != 3 {
			t.Errorf("page1 pagination = %+v, want total 5 pages 3", p1.Pagination)
		}
		if p1.Pagination.HasPrev || !p1.Pagination.HasNext {
			t.Errorf("page1 prev/next = %+v", p1.Pagination)
		}

		// last page -> 1 item, has_next false, has_prev true
		_, raw = do(t, ts, http.MethodGet, "/api/rooms?page=3&limit=2", "")
		var p3 listResp
		json.Unmarshal(raw, &p3)
		if len(p3.Data) != 1 {
			t.Errorf("page3 len = %d, want 1", len(p3.Data))
		}
		if !p3.Pagination.HasPrev || p3.Pagination.HasNext {
			t.Errorf("page3 prev/next = %+v", p3.Pagination)
		}
	})

	t.Run("invalid query falls back to defaults", func(t *testing.T) {
		ts, _ := setup(t)
		mustCreate(t, ts, "Only Room", "d")

		_, raw := do(t, ts, http.MethodGet, "/api/rooms?page=abc&limit=-5", "")
		var got listResp
		json.Unmarshal(raw, &got)
		if got.Pagination.Page != 1 || got.Pagination.Limit != 10 {
			t.Errorf("expected defaults page1 limit10, got %+v", got.Pagination)
		}
	})
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
