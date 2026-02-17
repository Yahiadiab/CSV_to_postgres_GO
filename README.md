# VPN CSV Importer

A flexible Go program that imports CSV data into PostgreSQL with two ingestion modes:
- **RAW mode**: Stores JSON data as-is in a JSONB column (hookah database)
- **STRUCTURED mode**: Unmarshals JSON and stores each field in separate columns (newkah database)

first ensure that we have a csv folder in the same directory as the program containing your CSV file. then ensure the databses are set and up running 

### Running the Program

**Process ALL CSV files:**
```bash
# RAW mode - imports to hookah database
go run . --mode=raw

# STRUCTURED mode - imports to newkah database  
go run . --mode=structured
```

**Process SPECIFIC files:**
```bash
# Process single file
go run . --mode=raw --files=NHK_VPN.csv

# Process multiple files
go run . --mode=structured --files="NHK_VPN.csv,NHK_ACCESS.csv,NHK_EQUIPMENT.csv"
```

**Build and run:**
```bash
# Build binary
go build -o vpn-importer .

# Run binary
./vpn-importer --mode=raw
./vpn-importer --mode=structured --files=NHK_VPN.csv
```



we can't proceed the structured mode because some of them have arrays in the jason and we can't map it to columns...
### 5 files to run with structured mode 
1. NHK_VPN.csv
2. NHK_ACCESS.csv
3. NHK_EQUIPMENT.csv
4. NHK_PORTABILITY.csv
5. NHK_SCL.csv



