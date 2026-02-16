package internal

// DATA STRUCTURES FOR STRUCTURED MODE
type VPNData struct {
	ID                      int64   `json:"id"`
	VpnCode                 string  `json:"vpn_code"`
	Name                    string  `json:"name"`
	VpnIDSip                string  `json:"vpn_id_sip"`
	CustomerReference       string  `json:"customer_reference"`
	ForcedOnnet             bool    `json:"forced_onnet"`
	PrivatePrefixLength     int     `json:"private_prefix_length"`
	OffnetViaBcrAllowed     bool    `json:"offnet_via_bcr_allowed"`
	ITKey                   string  `json:"it_key"`
	CustomerIndexForBilling *string `json:"customer_index_for_billing"` 
	HKMaster                string  `json:"hk_master"`
	HKRedirectPercentage    int     `json:"hk_redirect_percentage"`
	PartitionRight          string  `json:"partition_right"`
	State                   string  `json:"state"`
}

// AccessData represents the JSON structure from NHK_ACCESS.csv
type AccessData struct {
	ID                            int64   `json:"id"`
	EquipmentID                   int64   `json:"equipment_id"`
	SiteID                        int64   `json:"site_id"`
	Rank                          int     `json:"rank"`
	Load                          int     `json:"load"`
	CdrInfo                       string  `json:"cdr_info"`
	CalledPrefix                  string  `json:"called_prefix"`
	CalledNumberInPublicFormat    bool    `json:"called_number_in_public_format"`
	FlexibleCalled                *string `json:"flexible_called"`  
	FlexibleCalling               *string `json:"flexible_calling"` 
	CallingNumberInPublicFormat   bool    `json:"calling_number_in_public_format"`
	FullPrefix                    string  `json:"full_prefix"`
	CallingFormat                 string  `json:"calling_format"`
	CalledCliPolicy               string  `json:"called_cli_policy"`
}

// EquipmentData represents the JSON structure from NHK_EQUIPMENT.csv
type EquipmentData struct {
	ID                      int64   `json:"id"`
	Name                    string  `json:"name"`
	T1T7                    string  `json:"t1t7"`
	VpnID                   string  `json:"vpn_id"`
	IDDefaultSite           int64   `json:"id_default_site"`
	CspCustomerIdentifier   *string `json:"csp_customer_identifier"` 
	GlobalCallLimiter       *int    `json:"global_call_limiter"`     
	SclID                   *int    `json:"scl_id"`                  
	NumberPreAnalysisName   string  `json:"number_pre_analysis_name"`
	TrustedCliPolicy        bool    `json:"trusted_cli_policy"`
	CallingCliMatchAllSites bool    `json:"calling_cli_match_all_sites"`
	EquipmentType           string  `json:"equipment_type"`
	CallingSiteCliAlgo      string  `json:"calling_site_cli_algo"`
}

// PortabilityData represents the JSON structure from NHK_PORTABILITY.csv
type PortabilityData struct {
	ID                          int64   `json:"id"`
	InputPrefix                 string  `json:"input_prefix"`
	RoutingPrefix               string  `json:"routing_prefix"`
	OutputPrefix                string  `json:"output_prefix"`
	PortabilityType             string  `json:"portability_type"`
	ReqURIAdditionalParameters  string  `json:"req_uri_additional_parameters"`
	Comment                     *string `json:"comment"`     
	BcrProduct                  string  `json:"bcr_product"`
	BcrRegion                   *string `json:"bcr_region"`  
}

// SCLData represents the JSON structure from NHK_SCL.csv
type SCLData struct {
	ID                   int64   `json:"id"`
	Name                 string  `json:"name"`
	GlobalCallLimiter    int     `json:"global_call_limiter"`
	OutgoingCallLimiter  *int    `json:"outgoing_call_limiter"` 
	IncomingCallLimiter  *int    `json:"incoming_call_limiter"` 
	VpnID                string  `json:"vpn_id"`
	ResourceType         string  `json:"resource_type"`
}

// GenericRow represents a parsed CSV row with generic data
type GenericRow struct {
	CSVColumns map[string]string 
	JSONData   interface{}      
	RawDataStr string            
}
