-- ═══════════════════════════════════════════════════════════════════════════
-- STRUCTURED MODE SCHEMAS FOR COMPATIBLE CSV FILES
-- ═══════════════════════════════════════════════════════════════════════════
-- Database: newkah
-- Schema: public
--
-- This file contains structured table schemas for the 5 CSV files compatible
-- with structured mode (flat JSON structures):
-- 1. NHK_ACCESS.csv -> public.access
-- 2. NHK_EQUIPMENT.csv -> public.equipment
-- 3. NHK_PORTABILITY.csv -> public.portability
-- 4. NHK_SCL.csv -> public.scl
-- 5. NHK_VPN.csv -> public.vpn (already in schema_structured.sql)
-- ═══════════════════════════════════════════════════════════════════════════

-- ═══════════════════════════════════════════════════════════════════════════
-- 1. ACCESS TABLE (from NHK_ACCESS.csv)
-- ═══════════════════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS public.access (
    -- CSV columns
    id_site BIGINT NOT NULL,
    rank BIGINT NOT NULL,
    service_logic_leg INTEGER,
    access_type INTEGER,
    t1t7 TEXT,
    rn INTEGER,
    
    -- JSON data fields (extracted from 'data' column)
    id BIGINT,
    equipment_id BIGINT,
    site_id BIGINT,
    load INTEGER,
    cdr_info TEXT,
    called_prefix TEXT,
    called_number_in_public_format BOOLEAN,
    flexible_called TEXT,  -- nullable
    flexible_calling TEXT,  -- nullable
    calling_number_in_public_format BOOLEAN,
    full_prefix TEXT,
    calling_format TEXT,
    called_cli_policy TEXT,
    
    -- Metadata columns
    created_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Composite primary key
    PRIMARY KEY (id_site, rank)
);

-- Indexes for access table
CREATE INDEX IF NOT EXISTS idx_access_equipment_id ON public.access (equipment_id);
CREATE INDEX IF NOT EXISTS idx_access_site_id ON public.access (site_id);
CREATE INDEX IF NOT EXISTS idx_access_called_prefix ON public.access (called_prefix);

-- ═══════════════════════════════════════════════════════════════════════════
-- 2. EQUIPMENT TABLE (from NHK_EQUIPMENT.csv)
-- ═══════════════════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS public.equipment (
    -- CSV columns
    vpn_code BIGINT NOT NULL,
    id_equip BIGINT NOT NULL,
    
    -- JSON data fields (extracted from 'data' column)
    id BIGINT,
    name TEXT,
    t1t7 TEXT,
    vpn_id TEXT,
    id_default_site BIGINT,
    csp_customer_identifier TEXT,  -- nullable
    global_call_limiter INTEGER,  -- nullable
    scl_id INTEGER,  -- nullable
    number_pre_analysis_name TEXT,
    trusted_cli_policy BOOLEAN,
    calling_cli_match_all_sites BOOLEAN,
    equipment_type TEXT,
    calling_site_cli_algo TEXT,
    
    -- Metadata columns
    created_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Composite primary key
    PRIMARY KEY (vpn_code, id_equip)
);

-- Indexes for equipment table
CREATE INDEX IF NOT EXISTS idx_equipment_id ON public.equipment (id);
CREATE INDEX IF NOT EXISTS idx_equipment_name ON public.equipment (name);
CREATE INDEX IF NOT EXISTS idx_equipment_vpn_id ON public.equipment (vpn_id);
CREATE INDEX IF NOT EXISTS idx_equipment_type ON public.equipment (equipment_type);

-- ═══════════════════════════════════════════════════════════════════════════
-- 3. PORTABILITY TABLE (from NHK_PORTABILITY.csv)
-- ═══════════════════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS public.portability (
    -- CSV columns
    vpn_code BIGINT NOT NULL,
    id BIGINT NOT NULL,
    
    -- JSON data fields (extracted from 'data' column)
    input_prefix TEXT,
    routing_prefix TEXT,
    output_prefix TEXT,
    portability_type TEXT,
    req_uri_additional_parameters TEXT,
    comment TEXT,  -- nullable
    bcr_product TEXT,
    bcr_region TEXT,  -- nullable
    
    -- Metadata columns
    created_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Composite primary key
    PRIMARY KEY (vpn_code, id)
);


-- Indexes for portability table
CREATE INDEX IF NOT EXISTS idx_portability_input_prefix ON public.portability (input_prefix);
CREATE INDEX IF NOT EXISTS idx_portability_routing_prefix ON public.portability (routing_prefix);
CREATE INDEX IF NOT EXISTS idx_portability_type ON public.portability (portability_type);









CREATE TABLE IF NOT EXISTS public.vpn (
    -- Columns from CSV
    vpn_code BIGINT NOT NULL,
    id BIGINT NOT NULL,
    
    -- Columns from JSON data (unmarshaled)
    name TEXT NOT NULL,
    vpn_id_sip TEXT NOT NULL,
    customer_reference TEXT,
    forced_onnet BOOLEAN DEFAULT FALSE,
    private_prefix_length INTEGER,
    offnet_via_bcr_allowed BOOLEAN DEFAULT FALSE,
    it_key TEXT NOT NULL,  -- Primary key
    customer_index_for_billing TEXT,  -- Nullable
    hk_master TEXT,
    hk_redirect_percentage INTEGER,
    partition_right TEXT,
    state TEXT,
    
    -- Metadata columns
    created_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Primary key
    PRIMARY KEY (it_key)
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_vpn_vpn_code 
    ON public.vpn (vpn_code);

CREATE INDEX IF NOT EXISTS idx_vpn_id 
    ON public.vpn (id);

CREATE INDEX IF NOT EXISTS idx_vpn_name 
    ON public.vpn (name);

CREATE INDEX IF NOT EXISTS idx_vpn_state 
    ON public.vpn (state);

CREATE INDEX IF NOT EXISTS idx_vpn_created_date 
    ON public.vpn (created_date DESC);

CREATE INDEX IF NOT EXISTS idx_vpn_last_modified_date 
    ON public.vpn (last_modified_date DESC);

-- ═══════════════════════════════════════════════════════════════════════════
-- 4. SCL TABLE (from NHK_SCL.csv)
-- ═══════════════════════════════════════════════════════════════════════════
CREATE TABLE IF NOT EXISTS public.scl (
    -- CSV columns
    id BIGINT NOT NULL,
    vpn_code BIGINT NOT NULL,
    
    -- JSON data fields (extracted from 'data' column)
    name TEXT,
    global_call_limiter INTEGER,
    outgoing_call_limiter INTEGER,  -- nullable
    incoming_call_limiter INTEGER,  -- nullable
    vpn_id TEXT,
    resource_type TEXT,
    
    -- Metadata columns
    created_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_modified_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Composite primary key
    PRIMARY KEY (id, vpn_code)
);

-- Indexes for scl table
CREATE INDEX IF NOT EXISTS idx_scl_name ON public.scl (name);
CREATE INDEX IF NOT EXISTS idx_scl_vpn_id ON public.scl (vpn_id);
CREATE INDEX IF NOT EXISTS idx_scl_resource_type ON public.scl (resource_type);

-- ═══════════════════════════════════════════════════════════════════════════
-- COMMENTS
-- ═══════════════════════════════════════════════════════════════════════════

COMMENT ON TABLE public.access IS 'Access configuration data from NHK_ACCESS.csv';
COMMENT ON TABLE public.equipment IS 'Equipment configuration data from NHK_EQUIPMENT.csv';
COMMENT ON TABLE public.portability IS 'Portability routing data from NHK_PORTABILITY.csv';
COMMENT ON TABLE public.scl IS 'Service call limiter data from NHK_SCL.csv';
