package internal

type FileConfig struct {
	FileName            string   
	TableName           string   
	PKColumnsRaw        []string 
	PKColumnsStructured []string 
	SupportsStructured  bool     
}

// GetPKColumns returns the appropriate primary key columns based on mode
func (fc *FileConfig) GetPKColumns(mode string) []string {
	if mode == "raw" {
		return fc.PKColumnsRaw
	}
	return fc.PKColumnsStructured
}

// GetFileConfigs returns the configuration for all CSV files
func GetFileConfigs() []FileConfig {
	return []FileConfig{
		{
			FileName:            "NHK_ACCESS.csv",
			TableName:           "access",
			PKColumnsRaw:        []string{"id_site", "t1t7", "rn"},
			PKColumnsStructured: []string{"id_site", "rank"},
			SupportsStructured:  true,
		},
		{
			FileName:            "NHK_EQUIPMENT.csv",
			TableName:           "equipment",
			PKColumnsRaw:        []string{"id_equip"},
			PKColumnsStructured: []string{"vpn_code", "id_equip"},
			SupportsStructured:  true,
		},
		{
			FileName:            "NHK_PORTABILITY.csv",
			TableName:           "portability",
			PKColumnsRaw:        []string{"id"},
			PKColumnsStructured: []string{"vpn_code", "id"},
			SupportsStructured:  true,
		},
		{
			FileName:            "NHK_SCL.csv",
			TableName:           "scl",
			PKColumnsRaw:        []string{"id"},
			PKColumnsStructured: []string{"id", "vpn_code"},
			SupportsStructured:  true,
		},
		{
			FileName:            "NHK_VPN.csv",
			TableName:           "vpn",
			PKColumnsRaw:        []string{"vpn_code", "id"},
			PKColumnsStructured: []string{"vpn_code", "id"},
			SupportsStructured:  true,
		},


		{FileName: "NHK_ANNOUNCEMENT.csv", TableName: "announcement", PKColumnsRaw: []string{"id_announcement"}, SupportsStructured: false},
		{FileName: "NHK_BL.csv", TableName: "bl", PKColumnsRaw: []string{"id_blacklist"}, SupportsStructured: false},
		{FileName: "NHK_FULL_BHL.csv", TableName: "full_bhl", PKColumnsRaw: []string{"id"}, SupportsStructured: false},
		{FileName: "NHK_FULL_CNL.csv", TableName: "full_cnl", PKColumnsRaw: []string{"id"}, SupportsStructured: false},
		{FileName: "NHK_FULL_SITE.csv", TableName: "full_site", PKColumnsRaw: []string{"id"}, SupportsStructured: false},
		{FileName: "NHK_NR.csv", TableName: "nr", PKColumnsRaw: []string{"id"}, SupportsStructured: false},
		{FileName: "NHK_SAN.csv", TableName: "san", PKColumnsRaw: []string{"id_san"}, SupportsStructured: false},
		{FileName: "NHK_SAN_SMALL.csv", TableName: "san_small", PKColumnsRaw: []string{"id_san"}, SupportsStructured: false},
		{FileName: "NHK_SITE.csv", TableName: "site", PKColumnsRaw: []string{"id"}, SupportsStructured: false},
		{FileName: "NHK_WL.csv", TableName: "wl", PKColumnsRaw: []string{"id_whitelist"}, SupportsStructured: false},
	}
}

// DATABASE CONFIGURATION


type DatabaseConfig struct {
	Name        string
	URL         string
	Description string
}


func GetDatabaseConfigs() map[string]DatabaseConfig {
	return map[string]DatabaseConfig{
		"hookah": {
			Name:        "hookah",
			URL:         "postgres://VPN_USER:Orange123@localhost:5432/hookah?sslmode=disable",
			Description: "Hookah database (nhk schema) - for raw mode",
		},
		"newkah": {
			Name:        "newkah",
			URL:         "postgres://VPN_USER:Orange123@localhost:5432/newkah?sslmode=disable",
			Description: "Newkah database (public schema) - for structured mode",
		},
	}
}
