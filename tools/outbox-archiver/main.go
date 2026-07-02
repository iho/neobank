// outbox-archiver exports published outbox rows older than a retention window
// to newline-delimited JSON for object-storage upload (WORM bucket policy is
// enforced at the storage layer).
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type schema struct {
	name string
	qual string
}

var schemas = []schema{
	{name: "user", qual: `"user"`},
	{name: "payment", qual: "payment"},
	{name: "card", qual: "card"},
}

func main() {
	var (
		months   int
		outDir   string
		dryRun   bool
		database string
	)
	flag.IntVar(&months, "months", 12, "archive publications older than this many months")
	flag.StringVar(&outDir, "out", "./outbox-archive", "output directory for JSONL files")
	flag.BoolVar(&dryRun, "dry-run", false, "report counts only")
	flag.StringVar(&database, "database-url", "", "Postgres URL (or DATABASE_URL env)")
	flag.Parse()

	if database == "" {
		database = os.Getenv("DATABASE_URL")
	}
	if database == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL or -database-url required")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	cutoff := time.Now().UTC().AddDate(0, -months, 0)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
		os.Exit(1)
	}

	var total int
	for _, s := range schemas {
		n, err := archiveSchema(ctx, pool, s, cutoff, outDir, dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", s.name, err)
			os.Exit(1)
		}
		total += n
		fmt.Printf("%s: archived %d rows (cutoff %s)\n", s.name, n, cutoff.Format(time.RFC3339))
	}
	fmt.Printf("done: %d total rows\n", total)
}

type archiveRow struct {
	EventID      string    `json:"event_id"`
	Schema       string    `json:"schema"`
	PublishedAt  time.Time `json:"published_at"`
	EventType    string    `json:"event_type"`
	AggregateID  string    `json:"aggregate_id"`
	AggregateTyp string    `json:"aggregate_type"`
	Payload      []byte    `json:"payload"`
}

func archiveSchema(ctx context.Context, pool *pgxpool.Pool, s schema, cutoff time.Time, outDir string, dryRun bool) (int, error) {
	q := `
SELECT p.event_id, p.published_at, e.event_type, e.aggregate_id, e.aggregate_type, e.payload
FROM ` + s.qual + `.outbox_publications p
JOIN ` + s.qual + `.outbox_events e ON e.id = p.event_id
WHERE p.published_at < $1
ORDER BY p.published_at`

	rows, err := pool.Query(ctx, q, cutoff)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var out *os.File
	if !dryRun {
		filename := filepath.Join(outDir, fmt.Sprintf("%s-%s.jsonl", s.name, cutoff.Format("2006-01")))
		f, err := os.Create(filename)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		out = f
	}

	var count int
	for rows.Next() {
		var row archiveRow
		row.Schema = s.name
		if err := rows.Scan(&row.EventID, &row.PublishedAt, &row.EventType, &row.AggregateID, &row.AggregateTyp, &row.Payload); err != nil {
			return count, err
		}
		count++
		if dryRun || out == nil {
			continue
		}
		b, err := json.Marshal(row)
		if err != nil {
			return count, err
		}
		if _, err := out.Write(append(b, '\n')); err != nil {
			return count, err
		}
	}
	return count, rows.Err()
}