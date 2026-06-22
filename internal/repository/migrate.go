package repository

import (
	"errors"
	"strings"

	"github.com/fajarilf/go-starter-api/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
)

func toPgxURL(dbURL string) string {
	for _, prefix := range []string{"postgresql://", "postgres://"} {
		if after, ok := strings.CutPrefix(dbURL, prefix); ok {
			return "pgx5://" + after
		}
	}

	return dbURL
}

func NewMigrator(dbURL string) (*migrate.Migrate, error) {
	src, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, toPgxURL(dbURL))
	if err != nil {
		return nil, err
	}

	return m, nil
}

func Migrate(dbURL string) error {
	m, err := NewMigrator(dbURL)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
