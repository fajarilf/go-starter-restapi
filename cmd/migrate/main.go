package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/fajarilf/go-starter-api/internal/config"
	"github.com/fajarilf/go-starter-api/internal/repository"
	"github.com/golang-migrate/migrate/v4"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: migarte <up|down|version|force> [n]")
	}

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error load config: %v", err)
	}

	dbURL := cfg.DatabaseURL

	m, err := repository.NewMigrator(dbURL)
	if err != nil {
		log.Fatalf("Error migrator: %v", err)
	}
	defer m.Close()

	cmd := os.Args[1]
	switch cmd {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "version":
		printVersion(m)
	case "force":
		if len(os.Args) < 3 {
			log.Fatalf("usage: migrate force <version>")
		}
		v, convErr := strconv.Atoi(os.Args[2])
		if convErr != nil {
			log.Fatalf("invalid version %q: %v", os.Args[2], convErr)
		}
		err = m.Force(v)
	default:
		log.Fatalf("unknown command %q (want up|down|version|force)", cmd)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("%s: %v", cmd, err)
	}

	fmt.Printf("migrate %s: ok\n", cmd)
	printVersion(m)
}

func printVersion(m *migrate.Migrate) {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("version: none (no migration applied)")
	}
	if err != nil {
		log.Fatalf("Error version: %v", err)
	}

	fmt.Printf("version: %d (dirty=%t)\n", version, dirty)
}
