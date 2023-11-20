package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/MaxBlaushild/poltergeist/migrate/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	up   = "up"
	down = "down"
)

type verboseLogger struct{}

func (vl *verboseLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (vl *verboseLogger) Verbose() bool {
	return true
}

func makeDsn(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Public.DbUser,
		cfg.Secret.DbPassword,
		cfg.Public.DbHost,
		cfg.Public.DbPort,
		cfg.Public.DbName,
	)
}

func main() {
	steps := flag.Int("steps", -1, "The amount of migrations to run")
	direction := flag.String("direction", "up", "The direction to migrate")

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", makeDsn(cfg))
	if err != nil {
		panic(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/migrations",
		"postgres", driver)
	if err != nil {
		panic(err)
	}

	m.Log = &verboseLogger{}

	if *direction == up {
		if *steps == -1 {
			if err := m.Up(); err != nil {
				panic(err)
			}
		} else {
			if err := m.Steps(*steps); err != nil {
				panic(err)
			}
		}

		fmt.Println("up migrations successfully ran")
		return
	}

	if *direction == down {
		if *steps == -1 {
			if err := m.Up(); err != nil {
				panic(err)
			}
		} else {
			if err := m.Steps(*steps); err != nil {
				panic(err)
			}
		}

		fmt.Println("down migrations successfully ran")
		return
	}

	panic(fmt.Errorf("invalid direction: %s", *direction))
}
