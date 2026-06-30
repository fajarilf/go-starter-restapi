package handler_test

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
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roomData struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func TestHealthz(t *testing.T) {
	ts, _ := setup(t)

	status, raw := do(t, ts, http.MethodGet, "/api/healthz", "")
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "ok", string(raw))
}

func TestCreate(t *testing.T) {
	ts, _ := setup(t)

	t.Run("valid", func(t *testing.T) {
		got := mustCreate(t, ts, "Conference A", "Third floor")
		assert.NotZero(t, got.ID, "id not assigned")
		assert.Equal(t, "Conference A", got.Name)
		assert.Equal(t, "Third floor", got.Description)
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
			status, _ := do(t, ts, http.MethodPost, "/api/rooms", tc.body)
			assert.Equal(t, http.StatusBadRequest, status, "body: %s", tc.body)
		})
	}
}

func TestGetById(t *testing.T) {
	ts, _ := setup(t)
	created := mustCreate(t, ts, "Room A", "desc")

	t.Run("found", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodGet, path(created.ID), "")
		require.Equal(t, http.StatusOK, status)

		var got successResp
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, created.ID, got.Data.ID)
		assert.Equal(t, "Room A", got.Data.Name)
	})

	t.Run("not found", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodGet, "/api/rooms/999999", "")
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("non-numeric id", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodGet, "/api/rooms/abc", "")
		assert.Equal(t, http.StatusBadRequest, status)
	})
}

func TestUpdate(t *testing.T) {
	ts, _ := setup(t)
	created := mustCreate(t, ts, "Old Name", "old desc")

	t.Run("valid", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodPut, path(created.ID),
			`{"name":"New Name","description":"new desc"}`)
		require.Equal(t, http.StatusOK, status)

		var got successResp
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, "New Name", got.Data.Name)
		assert.Equal(t, "new desc", got.Data.Description)
	})

	t.Run("not found", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPut, "/api/rooms/999999",
			`{"name":"Whatever","description":"x"}`)
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("non-numeric id", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPut, "/api/rooms/abc",
			`{"name":"Whatever","description":"x"}`)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("validation fail", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPut, path(created.ID),
			`{"name":"ab","description":"x"}`)
		assert.Equal(t, http.StatusBadRequest, status)
	})

	t.Run("malformed json", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodPut, path(created.ID), `{`)
		assert.Equal(t, http.StatusBadRequest, status)
	})
}

func TestDelete(t *testing.T) {
	ts, _ := setup(t)
	created := mustCreate(t, ts, "To Delete", "desc")

	t.Run("existing", func(t *testing.T) {
		status, raw := do(t, ts, http.MethodDelete, path(created.ID), "")
		require.Equal(t, http.StatusOK, status)
		assert.Contains(t, string(raw), "room deleted")

		// gone afterwards
		status, _ = do(t, ts, http.MethodGet, path(created.ID), "")
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("not found", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodDelete, "/api/rooms/999999", "")
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("non-numeric id", func(t *testing.T) {
		status, _ := do(t, ts, http.MethodDelete, "/api/rooms/abc", "")
		assert.Equal(t, http.StatusBadRequest, status)
	})
}

func TestList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		ts, _ := setup(t)
		status, raw := do(t, ts, http.MethodGet, "/api/rooms", "")
		require.Equal(t, http.StatusOK, status)

		var got listResp
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Empty(t, got.Data)
		assert.Zero(t, got.Pagination.Total)
		assert.False(t, got.Pagination.HasPrev)
		assert.False(t, got.Pagination.HasNext)
	})

	t.Run("pagination", func(t *testing.T) {
		ts, _ := setup(t)
		for i := 0; i < 5; i++ {
			mustCreate(t, ts, "Room "+string(rune('A'+i)), "d")
		}

		// page 1, limit 2 -> 2 items, has_next true, has_prev false, 3 pages
		status, raw := do(t, ts, http.MethodGet, "/api/rooms?page=1&limit=2", "")
		require.Equal(t, http.StatusOK, status)

		var p1 listResp
		require.NoError(t, json.Unmarshal(raw, &p1))
		assert.Len(t, p1.Data, 2)
		assert.Equal(t, 5, p1.Pagination.Total)
		assert.Equal(t, 3, p1.Pagination.TotalPages)
		assert.False(t, p1.Pagination.HasPrev)
		assert.True(t, p1.Pagination.HasNext)

		// last page -> 1 item, has_next false, has_prev true
		_, raw = do(t, ts, http.MethodGet, "/api/rooms?page=3&limit=2", "")
		var p3 listResp
		require.NoError(t, json.Unmarshal(raw, &p3))
		assert.Len(t, p3.Data, 1)
		assert.True(t, p3.Pagination.HasPrev)
		assert.False(t, p3.Pagination.HasNext)
	})

	t.Run("invalid query falls back to defaults", func(t *testing.T) {
		ts, _ := setup(t)
		mustCreate(t, ts, "Only Room", "d")

		_, raw := do(t, ts, http.MethodGet, "/api/rooms?page=abc&limit=-5", "")
		var got listResp
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, 1, got.Pagination.Page)
		assert.Equal(t, 10, got.Pagination.Limit)
	})
}

func TestRecover(t *testing.T) {
	ts, _ := setup(t)
	token := mustLogin(t, ts)

	t.Run("soft-deleted is recovered", func(t *testing.T) {
		created := mustCreate(t, ts, "Recover Me", "desc")

		status, _ := do(t, ts, http.MethodDelete, path(created.ID), "")
		require.Equal(t, http.StatusOK, status)

		status, raw := doReq(t, ts, http.MethodPost, path(created.ID)+"/recover", "", token)
		require.Equal(t, http.StatusOK, status)

		var got successResp
		require.NoError(t, json.Unmarshal(raw, &got))
		assert.Equal(t, created.ID, got.Data.ID)
		assert.Equal(t, "Recover Me", got.Data.Name)

		status, _ = do(t, ts, http.MethodGet, path(created.ID), "")
		assert.Equal(t, http.StatusOK, status)
	})

	t.Run("already active is conflict", func(t *testing.T) {
		created := mustCreate(t, ts, "Still Active", "desc")

		status, _ := doReq(t, ts, http.MethodPost, path(created.ID)+"/recover", "", token)
		assert.Equal(t, http.StatusConflict, status)
	})

	t.Run("not found", func(t *testing.T) {
		status, _ := doReq(t, ts, http.MethodPost, "/api/rooms/999999/recover", "", token)
		assert.Equal(t, http.StatusNotFound, status)
	})

	t.Run("non-numeric id", func(t *testing.T) {
		status, _ := doReq(t, ts, http.MethodPost, "/api/rooms/abc/recover", "", token)
		assert.Equal(t, http.StatusBadRequest, status)
	})
}
