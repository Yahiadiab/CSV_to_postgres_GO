package internal

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcessFile processes a single CSV file
func ProcessFile(ctx context.Context, pool *pgxpool.Pool, csvPath string, config *FileConfig, mode string) (int, error) {
	// Open CSV file
	f, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	// Read header
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}

	// Build column index
	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Create inserter based on mode and table
	var inserter Inserter
	pkColumns := config.GetPKColumns(mode)
	if mode == "raw" {
		inserter = &RawInserter{PKColumns: pkColumns}
	} else {
		// Create appropriate structured inserter
		switch config.TableName {
		case "vpn":
			inserter = &VPNInserter{}
		case "access":
			inserter = &AccessInserter{}
		case "equipment":
			inserter = &EquipmentInserter{}
		case "portability":
			inserter = &PortabilityInserter{}
		case "scl":
			inserter = &SCLInserter{}
		default:
			return 0, fmt.Errorf("no structured inserter for table %s", config.TableName)
		}
	}

	// Parse and batch insert rows
	const batchSize = 1000
	batch := make([]GenericRow, 0, batchSize)
	inserted := 0
	lineNum := 1

	for {
		record, err := reader.Read()
		lineNum++

		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("read row %d: %w", lineNum, err)
		}

		// Parse record
		row, skip, err := parseGenericRecord(record, colIndex, config, mode)
		if err != nil {
			return 0, fmt.Errorf("parse error at line %d: %w", lineNum, err)
		}
		if skip {
			// Skip rows with empty primary key values
			continue
		}

		batch = append(batch, row)

		// Insert batch when full
		if len(batch) >= batchSize {
			if err := inserter.InsertBatch(ctx, pool, batch, config.TableName); err != nil {
				return 0, fmt.Errorf("insert batch failed: %w", err)
			}
			inserted += len(batch)
			log.Printf("  → inserted %d rows...", inserted)
			batch = batch[:0]
		}
	}

	// Insert remaining rows
	if len(batch) > 0 {
		if err := inserter.InsertBatch(ctx, pool, batch, config.TableName); err != nil {
			return 0, fmt.Errorf("insert final batch failed: %w", err)
		}
		inserted += len(batch)
	}

	return inserted, nil
}


func parseGenericRecord(record []string, colIndex map[string]int, config *FileConfig, mode string) (GenericRow, bool, error) {
	// Helper to get column value
	get := func(name string) string {
		i, ok := colIndex[name]
		if !ok || i < 0 || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	// Build CSV columns map
	csvCols := make(map[string]string)
	for name, idx := range colIndex {
		if idx >= 0 && idx < len(record) {
			csvCols[name] = strings.TrimSpace(record[idx])
		}
	}

	// Check if any primary key columns are empty or have obviously invalid values
	pkColumns := config.GetPKColumns(mode)
	for _, pkCol := range pkColumns {
		val := csvCols[pkCol]
		if val == "" {
			// Skip rows with empty primary keys
			return GenericRow{}, true, nil
		}
	}
	
	// Get JSON data column (usually "data", but some files use different names)
	dataColName := "data"
	if _, ok := colIndex["data"]; !ok {
		// Try alternative column names
		for _, alt := range []string{"announcement", "blacklist", "site", "whitelist"} {
			if _, ok := colIndex[alt]; ok {
				dataColName = alt
				break
			}
		}
	}

	dataStr := get(dataColName)
	if dataStr == "" {
		return GenericRow{}, false, fmt.Errorf("%s column is empty", dataColName)
	}

	// Validate JSON
	dataBytes := []byte(dataStr)
	if !json.Valid(dataBytes) {
		return GenericRow{}, false, fmt.Errorf("invalid JSON in %s column", dataColName)
	}

	row := GenericRow{
		CSVColumns: csvCols,
		RawDataStr: dataStr,
	}

	// Unmarshal JSON based on table type (only for structured mode)
	if mode == "structured" {
		switch config.TableName {
		case "vpn":
			var data VPNData
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				return GenericRow{}, false, fmt.Errorf("unmarshal VPNData: %w", err)
			}
			row.JSONData = &data
		case "access":
			var data AccessData
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				return GenericRow{}, false, fmt.Errorf("unmarshal AccessData: %w", err)
			}
			row.JSONData = &data
		case "equipment":
			var data EquipmentData
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				return GenericRow{}, false, fmt.Errorf("unmarshal EquipmentData: %w", err)
			}
			row.JSONData = &data
		case "portability":
			var data PortabilityData
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				return GenericRow{}, false, fmt.Errorf("unmarshal PortabilityData: %w", err)
			}
			row.JSONData = &data
		case "scl":
			var data SCLData
			if err := json.Unmarshal(dataBytes, &data); err != nil {
				return GenericRow{}, false, fmt.Errorf("unmarshal SCLData: %w", err)
			}
			row.JSONData = &data
		}
	}

	return row, false, nil
}
