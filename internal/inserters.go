package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ═══════════════════════════════════════════════════════════════════════════
// INSERTER INTERFACE (Strategy Pattern)
// ═══════════════════════════════════════════════════════════════════════════

// Inserter defines the strategy interface for inserting batches of rows
type Inserter interface {
	InsertBatch(ctx context.Context, pool *pgxpool.Pool, rows []GenericRow, tableName string) error
}

// ═══════════════════════════════════════════════════════════════════════════
// RAW MODE INSERTER
// ═══════════════════════════════════════════════════════════════════════════

// RawInserter implements the Inserter interface for RAW mode
type RawInserter struct {
	PKColumns []string // Primary key column names
}

// InsertBatch inserts rows into nhk.<table> with all CSV columns + JSONB data column
func (r *RawInserter) InsertBatch(ctx context.Context, pool *pgxpool.Pool, rows []GenericRow, tableName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if len(rows) == 0 {
		return nil
	}

	// Get all CSV column names from the first row (excluding 'data' column)
	var csvColNames []string
	dataColName := ""
	for colName := range rows[0].CSVColumns {
		// Find the data column (could be 'data', 'announcement', 'blacklist', 'site', 'whitelist')
		if colName == "data" || colName == "announcement" || colName == "blacklist" || colName == "site" || colName == "whitelist" {
			dataColName = colName
		} else {
			csvColNames = append(csvColNames, colName)
		}
	}

	if dataColName == "" {
		return fmt.Errorf("no data column found in CSV")
	}

	// Build column list and placeholders for SQL
	allColumns := append(csvColNames, dataColName) // Use actual column name (data/announcement/blacklist/site/whitelist)
	columnList := strings.Join(allColumns, ", ")
	
	// Build placeholders: $1, $2, ..., $N
	placeholders := make([]string, len(allColumns))
	for i := range allColumns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholderList := strings.Join(placeholders, ", ")

	// Build primary key list for ON CONFLICT
	pkCols := strings.Join(r.PKColumns, ", ")

	sql := fmt.Sprintf(`
		INSERT INTO nhk.%s (%s)
		VALUES (%s)
		ON CONFLICT (%s) DO NOTHING
	`, tableName, columnList, placeholderList, pkCols)

	for _, row := range rows {
		// Build args array: all CSV column values + JSON data
		args := make([]interface{}, len(allColumns))
		
		// Add CSV column values (keep as strings, let PostgreSQL handle type conversion)
		for i, colName := range csvColNames {
			val := row.CSVColumns[colName]
			if val == "" {
				args[i] = nil // NULL for empty values
			} else {
				args[i] = val // Keep as string
			}
		}
		
		// Add data column (JSONB)
		args[len(csvColNames)] = row.RawDataStr

		_, err := tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("exec insert (raw mode): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// ═══════════════════════════════════════════════════════════════════════════
// STRUCTURED MODE INSERTERS (one per table type)
// ═══════════════════════════════════════════════════════════════════════════

// VPNInserter handles structured inserts for VPN table
type VPNInserter struct{}

func (v *VPNInserter) InsertBatch(ctx context.Context, pool *pgxpool.Pool, rows []GenericRow, tableName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const sql = `
		INSERT INTO public.vpn (
			vpn_code, id, name, vpn_id_sip, customer_reference, forced_onnet,
			private_prefix_length, offnet_via_bcr_allowed, it_key,
			customer_index_for_billing, hk_master, hk_redirect_percentage,
			partition_right, state
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (it_key)
		DO UPDATE SET
			vpn_code = EXCLUDED.vpn_code,
			name = EXCLUDED.name,
			vpn_id_sip = EXCLUDED.vpn_id_sip,
			customer_reference = EXCLUDED.customer_reference,
			forced_onnet = EXCLUDED.forced_onnet,
			private_prefix_length = EXCLUDED.private_prefix_length,
			offnet_via_bcr_allowed = EXCLUDED.offnet_via_bcr_allowed,
			customer_index_for_billing = EXCLUDED.customer_index_for_billing,
			hk_master = EXCLUDED.hk_master,
			hk_redirect_percentage = EXCLUDED.hk_redirect_percentage,
			partition_right = EXCLUDED.partition_right,
			state = EXCLUDED.state,
			last_modified_date = NOW()
	`

	for _, row := range rows {
		data := row.JSONData.(*VPNData)
		_, err := tx.Exec(ctx, sql,
			data.VpnCode, data.ID, data.Name, data.VpnIDSip, data.CustomerReference,
			data.ForcedOnnet, data.PrivatePrefixLength, data.OffnetViaBcrAllowed,
			data.ITKey, data.CustomerIndexForBilling, data.HKMaster,
			data.HKRedirectPercentage, data.PartitionRight, data.State,
		)
		if err != nil {
			return fmt.Errorf("exec insert (vpn): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// AccessInserter handles structured inserts for ACCESS table
type AccessInserter struct{}

func (a *AccessInserter) InsertBatch(ctx context.Context, pool *pgxpool.Pool, rows []GenericRow, tableName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const sql = `
		INSERT INTO public.access (
			id_site, rank, service_logic_leg, access_type, t1t7, rn,
			id, equipment_id, site_id, load, cdr_info, called_prefix,
			called_number_in_public_format, flexible_called, flexible_calling,
			calling_number_in_public_format, full_prefix, calling_format, called_cli_policy
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		ON CONFLICT (id_site, rank) DO UPDATE SET
			service_logic_leg = EXCLUDED.service_logic_leg,
			access_type = EXCLUDED.access_type,
			t1t7 = EXCLUDED.t1t7,
			rn = EXCLUDED.rn,
			id = EXCLUDED.id,
			equipment_id = EXCLUDED.equipment_id,
			site_id = EXCLUDED.site_id,
			load = EXCLUDED.load,
			cdr_info = EXCLUDED.cdr_info,
			called_prefix = EXCLUDED.called_prefix,
			called_number_in_public_format = EXCLUDED.called_number_in_public_format,
			flexible_called = EXCLUDED.flexible_called,
			flexible_calling = EXCLUDED.flexible_calling,
			calling_number_in_public_format = EXCLUDED.calling_number_in_public_format,
			full_prefix = EXCLUDED.full_prefix,
			calling_format = EXCLUDED.calling_format,
			called_cli_policy = EXCLUDED.called_cli_policy,
			last_modified_date = NOW()
	`

	for _, row := range rows {
		data := row.JSONData.(*AccessData)
		csvCols := row.CSVColumns
		
		_, err := tx.Exec(ctx, sql,
			csvCols["id_site"], csvCols["rank"], csvCols["service_logic_leg"],
			csvCols["access_type"], csvCols["t1t7"], csvCols["rn"],
			data.ID, data.EquipmentID, data.SiteID, data.Load, data.CdrInfo,
			data.CalledPrefix, data.CalledNumberInPublicFormat, data.FlexibleCalled,
			data.FlexibleCalling, data.CallingNumberInPublicFormat, data.FullPrefix,
			data.CallingFormat, data.CalledCliPolicy,
		)
		if err != nil {
			return fmt.Errorf("exec insert (access): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// EquipmentInserter handles structured inserts for EQUIPMENT table
type EquipmentInserter struct{}

func (e *EquipmentInserter) InsertBatch(ctx context.Context, pool *pgxpool.Pool, rows []GenericRow, tableName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const sql = `
		INSERT INTO public.equipment (
			vpn_code, id_equip, id, name, t1t7, vpn_id, id_default_site,
			csp_customer_identifier, global_call_limiter, scl_id,
			number_pre_analysis_name, trusted_cli_policy, calling_cli_match_all_sites,
			equipment_type, calling_site_cli_algo
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (vpn_code, id_equip) DO UPDATE SET
			id = EXCLUDED.id,
			name = EXCLUDED.name,
			t1t7 = EXCLUDED.t1t7,
			vpn_id = EXCLUDED.vpn_id,
			id_default_site = EXCLUDED.id_default_site,
			csp_customer_identifier = EXCLUDED.csp_customer_identifier,
			global_call_limiter = EXCLUDED.global_call_limiter,
			scl_id = EXCLUDED.scl_id,
			number_pre_analysis_name = EXCLUDED.number_pre_analysis_name,
			trusted_cli_policy = EXCLUDED.trusted_cli_policy,
			calling_cli_match_all_sites = EXCLUDED.calling_cli_match_all_sites,
			equipment_type = EXCLUDED.equipment_type,
			calling_site_cli_algo = EXCLUDED.calling_site_cli_algo,
			last_modified_date = NOW()
	`

	for _, row := range rows {
		data := row.JSONData.(*EquipmentData)
		csvCols := row.CSVColumns
		
		_, err := tx.Exec(ctx, sql,
			csvCols["vpn_code"], csvCols["id_equip"],
			data.ID, data.Name, data.T1T7, data.VpnID, data.IDDefaultSite,
			data.CspCustomerIdentifier, data.GlobalCallLimiter, data.SclID,
			data.NumberPreAnalysisName, data.TrustedCliPolicy, data.CallingCliMatchAllSites,
			data.EquipmentType, data.CallingSiteCliAlgo,
		)
		if err != nil {
			return fmt.Errorf("exec insert (equipment): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// PortabilityInserter handles structured inserts for PORTABILITY table
type PortabilityInserter struct{}

func (p *PortabilityInserter) InsertBatch(ctx context.Context, pool *pgxpool.Pool, rows []GenericRow, tableName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const sql = `
		INSERT INTO public.portability (
			vpn_code, id, input_prefix, routing_prefix, output_prefix,
			portability_type, req_uri_additional_parameters, comment,
			bcr_product, bcr_region
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (vpn_code, id) DO UPDATE SET
			input_prefix = EXCLUDED.input_prefix,
			routing_prefix = EXCLUDED.routing_prefix,
			output_prefix = EXCLUDED.output_prefix,
			portability_type = EXCLUDED.portability_type,
			req_uri_additional_parameters = EXCLUDED.req_uri_additional_parameters,
			comment = EXCLUDED.comment,
			bcr_product = EXCLUDED.bcr_product,
			bcr_region = EXCLUDED.bcr_region,
			last_modified_date = NOW()
	`

	for _, row := range rows {
		data := row.JSONData.(*PortabilityData)
		csvCols := row.CSVColumns
		
		_, err := tx.Exec(ctx, sql,
			csvCols["vpn_code"], csvCols["id"],
			data.InputPrefix, data.RoutingPrefix, data.OutputPrefix,
			data.PortabilityType, data.ReqURIAdditionalParameters, data.Comment,
			data.BcrProduct, data.BcrRegion,
		)
		if err != nil {
			return fmt.Errorf("exec insert (portability): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// SCLInserter handles structured inserts for SCL table
type SCLInserter struct{}

func (s *SCLInserter) InsertBatch(ctx context.Context, pool *pgxpool.Pool, rows []GenericRow, tableName string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const sql = `
		INSERT INTO public.scl (
			id, vpn_code, name, global_call_limiter, outgoing_call_limiter,
			incoming_call_limiter, vpn_id, resource_type
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id, vpn_code) DO UPDATE SET
			name = EXCLUDED.name,
			global_call_limiter = EXCLUDED.global_call_limiter,
			outgoing_call_limiter = EXCLUDED.outgoing_call_limiter,
			incoming_call_limiter = EXCLUDED.incoming_call_limiter,
			vpn_id = EXCLUDED.vpn_id,
			resource_type = EXCLUDED.resource_type,
			last_modified_date = NOW()
	`

	for _, row := range rows {
		data := row.JSONData.(*SCLData)
		csvCols := row.CSVColumns
		
		_, err := tx.Exec(ctx, sql,
			csvCols["id"], csvCols["vpn_code"],
			data.Name, data.GlobalCallLimiter, data.OutgoingCallLimiter,
			data.IncomingCallLimiter, data.VpnID, data.ResourceType,
		)
		if err != nil {
			return fmt.Errorf("exec insert (scl): %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
