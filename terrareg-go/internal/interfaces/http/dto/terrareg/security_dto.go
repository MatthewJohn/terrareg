package terrareg

// TerraregSecurityResult represents a tfsec security scan result.
// Contains all fields from tfsec JSON output.
// Python reference: /app/terrareg/models.py get_tfsec_failures()
type TerraregSecurityResult struct {
	RuleID          string                   `json:"rule_id"`
	LongID          string                   `json:"long_id"`
	RuleDescription string                   `json:"rule_description"`
	RuleProvider    string                   `json:"rule_provider"`
	RuleService     string                   `json:"rule_service"`
	Impact          string                   `json:"impact"`
	Resolution      string                   `json:"resolution"`
	Links           []string                 `json:"links"`
	Description     string                   `json:"description"`
	Severity        string                   `json:"severity"`
	Warning         bool                     `json:"warning"`
	Status          int                      `json:"status"`
	Resource        string                   `json:"resource"`
	Location        TerraregSecurityLocation `json:"location"`
}

// TerraregSecurityLocation represents the location of a security issue.
type TerraregSecurityLocation struct {
	Filename  string `json:"filename"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}
