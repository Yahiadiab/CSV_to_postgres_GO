package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Row struct {
	VpnCode int64
	ID      int64
	Data    []byte // JSON bytes
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run . <path-to-csv>")
	}
	csvPath := os.Args[1]

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()

	// Create a connection pool
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	// Open the CSV and build a reader
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("open csv: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	// Read header + map columns
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("read header: %v", err)
	}
	col := buildColumnIndex(header)

	// Ensure required columns exist
	for _, name := range []string{"vpn_code", "id", "data"} {
		if _, ok := col[name]; !ok {
			log.Fatalf("missing required column %q in header: %v", name, header)
		}
	}

	// Batch config
	const batchSize = 1000
	batch := make([]Row, 0, batchSize)

	inserted := 0
	lineNum := 1 // header is line 1

	// Read rows in a loop until EOF
	for {
		record, err := reader.Read()
		lineNum++

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("read row %d: %v", lineNum, err)
		}

		// Parse + validate row
		row, err := parseRecord(record, col)
		if err != nil {
			log.Fatalf("parse error at line %d: %v\nrecord=%v", lineNum, err, record)
		}

		// Add to batch
		batch = append(batch, row)

		// Insert each full batch
		if len(batch) >= batchSize {
			if err := insertBatch(ctx, pool, batch); err != nil {
				log.Fatalf("insert batch failed: %v", err)
			}
			inserted += len(batch)
			log.Printf("inserted %d rows...", inserted)
			batch = batch[:0]
		}
	}

	// Insert remaining rows
	if len(batch) > 0 {
		if err := insertBatch(ctx, pool, batch); err != nil {
			log.Fatalf("insert final batch failed: %v", err)
		}
		inserted += len(batch)
	}

	log.Printf("done. inserted %d rows", inserted)
}

func buildColumnIndex(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

func parseRecord(record []string, col map[string]int) (Row, error) {
	get := func(name string) string {
		i := col[name]
		if i < 0 || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	vpnStr := get("vpn_code")
	vpn, err := strconv.ParseInt(vpnStr, 10, 64)
	if err != nil {
		return Row{}, fmt.Errorf("invalid vpn_code %q: %w", vpnStr, err)
	}

	idStr := get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return Row{}, fmt.Errorf("invalid id %q: %w", idStr, err)
	}

	dataStr := get("data")
	if dataStr == "" {
		return Row{}, fmt.Errorf("data is empty")
	}

	dataBytes := []byte(dataStr)
	if !json.Valid(dataBytes) {
		return Row{}, fmt.Errorf("data is not valid JSON (starts with): %.120s", dataStr)
	}

	return Row{VpnCode: vpn, ID: id, Data: dataBytes}, nil
}

func insertBatch(ctx context.Context, pool *pgxpool.Pool, rows []Row) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const sql = `
		INSERT INTO vpn_records (vpn_code, id, data)
		VALUES ($1, $2, $3::jsonb)
		ON CONFLICT (vpn_code, id)
		DO UPDATE SET data = EXCLUDED.data
	`

	for _, r := range rows {
		if _, err := tx.Exec(ctx, sql, r.VpnCode, r.ID, r.Data); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
