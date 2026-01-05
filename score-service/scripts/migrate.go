// scripts/migrate.go
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run scripts/migrate.go <command>")
	}

	// Получаем URL БД из переменных окружения или используем по умолчанию
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Формат: postgres://username:password@host:port/database?sslmode=disable
		dbURL = "postgresql://archy:12345678@localhost:5430/archy?sslmode=disable"
	}

	m, err := migrate.New(
		"file://internal/database/migrations",
		dbURL,
	)
	if err != nil {
		log.Fatal("Failed to create migrate instance:", err)
	}

	cmd := os.Args[1]

	switch cmd {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Migration up failed:", err)
		}
		fmt.Println("✅ Migrations applied successfully")

	case "down":
		steps := 1
		if len(os.Args) > 2 {
			if s, err := fmt.Sscanf(os.Args[2], "%d", &steps); err != nil || s != 1 {
				steps = 1
			}
		}
		if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Migration down failed:", err)
		}
		fmt.Printf("✅ Rolled back %d migration(s)\n", steps)

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate force <version>")
		}
		version := os.Args[2]
		ver, err := strconv.Atoi(version)
		if err != nil {
			log.Fatal("Failed to parse version:", err)
		}
		if err := m.Force(ver); err != nil {
			log.Fatal("Force version failed:", err)
		}
		fmt.Printf("✅ Forced migration to version %s\n", version)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal("Failed to get version:", err)
		}
		fmt.Printf("Version: %d, Dirty: %v\n", version, dirty)

	case "drop":
		if err := m.Drop(); err != nil {
			log.Fatal("Drop failed:", err)
		}
		fmt.Println("✅ Database dropped")

	default:
		log.Fatal("Unknown command. Available: up, down, force, version, drop")
	}
}
