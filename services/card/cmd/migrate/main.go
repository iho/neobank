package main

import (
	"fmt"
	"os"
	"path/filepath"

	dbmigrate "github.com/iho/neobank/pkg/migrate"
	"github.com/iho/neobank/services/card/internal/config"
)

func main() {
	cfg := config.Load()
	dir, err := filepath.Abs("migrations")
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve migrations dir: %v\n", err)
		os.Exit(1)
	}

	if err := dbmigrate.Up(cfg.DatabaseURL, dir, dbmigrate.Config{SchemaName: "card"}); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("card service migrations applied")
}