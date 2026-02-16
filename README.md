# VPN CSV Importer

A flexible Go program that imports CSV data into PostgreSQL with two ingestion modes:
- **RAW mode**: Stores JSON data as-is in a JSONB column
- **STRUCTURED mode**: Unmarshals JSON and stores each field in separate columns

## Features

✅ Single CSV parsing pipeline  
✅ Mode selection via CLI flag  
✅ Batched inserts with transactions  
✅ Connection pooling with pgx  
✅ Schema-qualified table names  
✅ Production-grade error handling  
✅ Comprehensive comments  

## Architecture

The program uses the **Strategy Pattern** to cleanly separate the two insertion modes:

```
┌─────────────────────────────────────────────────────────┐
│                      Main Pipeline                       │
│  1. Parse CLI flags                                      │
│  2. Select database connection (based on mode)           │
│  3. Create inserter strategy (RawInserter or             │
│     StructuredInserter)                                  │
│  4. Parse CSV (mode-agnostic)                            │
│  5. Batch insert via selected strategy                   │
└─────────────────────────────────────────────────────────┘
```

### Components

- **`Inserter` interface**: Strategy pattern for insertion logic
- **`RawInserter`**: Implements raw JSON storage
- **`StructuredInserter`**: Implements structured column storage
- **`parseRecord()`**: Mode-agnostic CSV parsing
- **`buildColumnIndex()`**: Column name mapping helper

## CSV Format

The CSV must contain three columns:

| Column    | Type   | Description                    |
|-----------|--------|--------------------------------|
| vpn_code  | bigint | VPN code identifier            |
| id        | bigint | Record ID                      |
| data      | JSON   | JSON string with VPN data      |

### Example CSV

```csv
vpn_code,id,data
12345,1,"{""name"":""VPN-A"",""vpn_id_sip"":""sip123"",""it_key"":""key1"",""state"":""active""}"
12346,2,"{""name"":""VPN-B"",""vpn_id_sip"":""sip456"",""it_key"":""key2"",""state"":""inactive""}"
```

## Database Setup

### RAW Mode

**Database**: `hookah`  
**Schema**: `nhk`  
**Table**: `raw_records`

```bash
# Connect to hookah database
psql -U postgres -d hookah

# Run schema
\i schema_raw.sql
```

**Table Structure**:
- `vpn_code` (bigint) - part of composite PK
- `id` (bigint) - part of composite PK
- `data` (jsonb) - JSON stored as-is
- `created_at` (timestamp)

**Primary Key**: `(vpn_code, id)`

### STRUCTURED Mode

**Database**: `newkah`  
**Schema**: `public`  
**Table**: `vpn`

```bash
# Connect to newkah database
psql -U postgres -d newkah

# Run schema
\i schema_structured.sql
```

**Table Structure**:
- `vpn_code` (bigint)
- `id` (bigint)
- `name` (text) - required
- `vpn_id_sip` (text) - required
- `customer_reference` (text)
- `forced_onnet` (boolean)
- `private_prefix_length` (integer)
- `offnet_via_bcr_allowed` (boolean)
- `it_key` (text) - **primary key**
- `customer_index_for_billing` (text) - nullable
- `hk_master` (text)
- `hk_redirect_percentage` (integer)
- `partition_right` (text)
- `state` (text)
- `created_date` (timestamp)
- `last_modified_date` (timestamp)

**Primary Key**: `it_key`  
**Conflict Resolution**: Upsert on `it_key` conflict

## Environment Variables

Set the appropriate database connection string based on the mode you'll use:

```bash
# For RAW mode (connects to hookah database)
export DATABASE_URL_RAW="postgres://user:password@localhost:5432/hookah?sslmode=disable"

# For STRUCTURED mode (connects to newkah database)
export DATABASE_URL_STRUCTURED="postgres://user:password@localhost:5432/newkah?sslmode=disable"
```

## Usage

### RAW Mode

Stores JSON data as-is into `hookah.nhk.raw_records`:

```bash
go run . --mode=raw csv/sample.csv
```

**Output**:
```
Mode: RAW → Target: hookah.nhk.raw_records
inserted 1000 rows...
inserted 2000 rows...
✓ Done. Inserted 2500 rows in RAW mode
```

### STRUCTURED Mode

Unmarshals JSON and stores fields in separate columns in `newkah.public.vpn`:

```bash
go run . --mode=structured csv/sample.csv
```

**Output**:
```
Mode: STRUCTURED → Target: newkah.public.vpn
inserted 1000 rows...
inserted 2000 rows...
✓ Done. Inserted 2500 rows in STRUCTURED mode
```

## Build

```bash
# Build binary
go build -o vpn-importer .

# Run with binary
./vpn-importer --mode=raw csv/sample.csv
./vpn-importer --mode=structured csv/sample.csv
```

## Configuration

### Batch Size

Default batch size is **1000 rows**. To change it, modify the constant in `main.go`:

```go
const batchSize = 1000  // Adjust as needed
```

### Timeout

Default transaction timeout is **60 seconds**. Adjust in the inserter methods:

```go
ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
```

## Error Handling

The program will exit with an error if:

- ❌ Invalid mode flag (must be `raw` or `structured`)
- ❌ Missing CSV file path
- ❌ Missing required environment variable
- ❌ Database connection failure
- ❌ Missing required CSV columns (`vpn_code`, `id`, `data`)
- ❌ Invalid CSV data (malformed integers, invalid JSON)
- ❌ Missing required JSON fields (in STRUCTURED mode only)
- ❌ Database insertion failure

## Validation

### RAW Mode Validation

- ✅ `vpn_code` is a valid bigint
- ✅ `id` is a valid bigint
- ✅ `data` is valid JSON syntax

### STRUCTURED Mode Validation

All RAW mode validations **plus**:

- ✅ `name` field is non-empty
- ✅ `vpn_id_sip` field is non-empty
- ✅ `it_key` field is non-empty

## Examples

### Example 1: Import to RAW mode

```bash
export DATABASE_URL_RAW="postgres://admin:secret@localhost:5432/hookah"
go run . --mode=raw csv/vpn_data.csv
```

### Example 2: Import to STRUCTURED mode

```bash
export DATABASE_URL_STRUCTURED="postgres://admin:secret@localhost:5432/newkah"
go run . --mode=structured csv/vpn_data.csv
```

### Example 3: Verify RAW mode data

```bash
psql -U postgres -d hookah -c "SELECT vpn_code, id, data->>'name' as name FROM nhk.raw_records LIMIT 5;"
```

### Example 4: Verify STRUCTURED mode data

```bash
psql -U postgres -d newkah -c "SELECT vpn_code, id, name, it_key, state FROM public.vpn LIMIT 5;"
```

## Dependencies

- Go 1.21+
- [pgx/v5](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit

```bash
go mod download
```

## License

MIT