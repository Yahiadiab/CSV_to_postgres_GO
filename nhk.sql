-- SQL Script to create tables for all NHK CSV files
-- Database: hookah
-- Schema: nhk
-- Each table follows the pattern: composite primary key + JSONB data column

-- 1. ACCESS Table

CREATE TABLE IF NOT EXISTS nhk.access (
  id_site            BIGINT  NOT NULL,
  rank               BIGINT  NULL,          
  service_logic_leg  BIGINT  NULL,          
  access_type        BIGINT  NOT NULL,
  t1t7               BIGINT  NOT NULL,
  rn                 BIGINT  NOT NULL,
  data               JSONB   NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (id_site, t1t7, rn)
);

-- 2. ANNOUNCEMENT Table

CREATE TABLE IF NOT EXISTS nhk.announcement (
  id_announcement BIGINT NOT NULL,
  name            TEXT   NOT NULL,
  announcement    JSONB  NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (id_announcement)
);



-- 3. BL Table

CREATE TABLE IF NOT EXISTS nhk.bl (
  id_blacklist BIGINT NOT NULL,
  name         TEXT   NOT NULL,
  blacklist    JSONB  NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (id_blacklist)
);

-- 4. EQUIPMENT Tablex

CREATE TABLE IF NOT EXISTS nhk.equipment (
  vpn_code  BIGINT NOT NULL,
  id_equip  BIGINT NOT NULL,
  data      JSONB  NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (id_equip)
);

-- 5. FULL_BHL Table

CREATE TABLE IF NOT EXISTS nhk.full_bhl (
  vpn_code   BIGINT NOT NULL,
  id         BIGINT NOT NULL,
  data       JSONB  NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (id)
);

-- 6. FULL_CNL Table
CREATE TABLE IF NOT EXISTS nhk.full_cnl (
  vpn_code   BIGINT NOT NULL,
  id         BIGINT NOT NULL,
  data       JSONB  NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (id)
);

-- 7. FULL_SITE Table

CREATE TABLE IF NOT EXISTS nhk.full_site (
  id         BIGINT NOT NULL,
  vpn_code   BIGINT NOT NULL,
  data       JSONB  NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id)

);

-- 8. NR Table

CREATE TABLE IF NOT EXISTS nhk.nr (

  vpn_code   BIGINT NOT NULL,
  id         BIGINT NOT NULL,
  data       JSONB  NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id)
);

-- 9. PORTABILITY Table


CREATE TABLE IF NOT EXISTS nhk.portability (
  id         BIGINT NOT NULL,
  vpn_code   BIGINT NULL,     
  data       JSONB  NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id)
);

-- 10. SAN Table


CREATE TABLE IF NOT EXISTS nhk.san (
  id_san         BIGINT NOT NULL,
  dialled_number TEXT NOT NULL,
  san_type       BIGINT NOT NULL,
  vpn_code       BIGINT NOT NULL,
  data           JSONB  NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (id_san)
);

-- 11. SAN_SMALL Table

CREATE TABLE IF NOT EXISTS nhk.san_small (
  id_san         BIGINT NOT NULL,
  dialled_number TEXT NOT NULL,
  san_type       BIGINT NOT NULL,
  vpn_code       BIGINT NOT NULL,
  data           JSONB  NOT NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (id_san)
);

-- 12. SCL Table

CREATE TABLE IF NOT EXISTS nhk.scl (
  id         BIGINT NOT NULL,
  vpn_code   BIGINT NULL,   -- IMPORTANT: missing in 8 rows
  data       JSONB  NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id)
);



-- 13. SITE Table

CREATE TABLE IF NOT EXISTS nhk.site (
  id         BIGINT NOT NULL,
  site       JSONB  NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id)
);


-- 14. VPN Table (already exists, included for completeness)
CREATE TABLE IF NOT EXISTS nhk.VPN (
   
    vpn_code BIGINT NOT NULL,
    id BIGINT NOT NULL,
    
    data JSONB NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (vpn_code, id)
);

-- 15. WL Table
CREATE TABLE IF NOT EXISTS nhk.wl (
  id_whitelist BIGINT NOT NULL,
  name         TEXT   NOT NULL,
  whitelist    JSONB  NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (id_whitelist)
);





-- Create indexes on JSONB data column for better query performance (optional)

