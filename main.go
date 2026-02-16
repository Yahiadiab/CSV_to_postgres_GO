package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"
	"strings"

	"vpn-importer/internal"

	"github.com/jackc/pgx/v5/pgxpool"
)


func main() {
	// Parse CLI flags
	var mode string
	flag.StringVar(&mode, "mode", "", "Ingestion mode: 'raw' or 'structured'")
	flag.Parse()

	if mode != "raw" && mode != "structured" {
		log.Fatal("usage: go run . --mode=<raw|structured>")
	}

	// Get database configurations
	dbConfigs := internal.GetDatabaseConfigs()

	// Automatically select database based on mode
	var dbURL string
	if mode == "raw" {
		config := dbConfigs["hookah"]
		dbURL = config.URL
		log.Printf("Mode: RAW → Database: %s (%s)", config.Name, config.Description)
	} else {
		config := dbConfigs["newkah"]
		dbURL = config.URL
		log.Printf("Mode: STRUCTURED → Database: %s (%s)", config.Name, config.Description)
	}

	// Create database connection pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	// Get all file configurations
	fileConfigs := internal.GetFileConfigs()

	// Discover CSV files in csv/ directory
	csvFiles, err := filepath.Glob("csv/*.csv")
	if err != nil {
		log.Fatalf("failed to list CSV files: %v", err)
	}

	if len(csvFiles) == 0 {
		log.Fatal("no CSV files found in csv/ directory")
	}

	log.Printf("Found %d CSV files to process", len(csvFiles))

	// Process each CSV file
	totalInserted := 0
	processedCount := 0
	skippedCount := 0

	for _, csvPath := range csvFiles {
		fileName := filepath.Base(csvPath)
		
		// Find configuration for this file
		var config *internal.FileConfig
		for i := range fileConfigs {
			if fileConfigs[i].FileName == fileName {
				config = &fileConfigs[i]
				break
			}
		}

		if config == nil {
			log.Printf("⚠️  Skipping %s (no configuration found)", fileName)
			skippedCount++
			continue
		}

		// Skip files not compatible with structured mode
		if mode == "structured" && !config.SupportsStructured {
			log.Printf("⏭️  Skipping %s (not compatible with structured mode)", fileName)
			skippedCount++
			continue
		}

		log.Printf("\n📄 Processing %s → %s.%s", fileName, mode, config.TableName)

		// Process the file
		inserted, err := internal.ProcessFile(ctx, pool, csvPath, config, mode)
		if err != nil {
			log.Fatalf("❌ Failed to process %s: %v", fileName, err)
		}

		log.Printf("✅ Completed %s: inserted %d rows", fileName, inserted)
		totalInserted += inserted
		processedCount++
	}

	log.Printf("\n" + strings.Repeat("═", 80))
	log.Printf("✓ DONE")
	log.Printf("  Processed: %d files", processedCount)
	log.Printf("  Skipped:   %d files", skippedCount)
	log.Printf("  Inserted:  %d total rows", totalInserted)
	log.Printf("  Mode:      %s", strings.ToUpper(mode))
	log.Printf(strings.Repeat("═", 80))
}
