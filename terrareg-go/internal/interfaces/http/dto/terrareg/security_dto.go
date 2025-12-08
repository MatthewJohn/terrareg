package terrareg

// TerraregSecurityResult represents a tfsec security scan result.
type TerraregSecurityResult struct {
	RuleID      string                      `json:"rule_id"`
	Severity    string                      `json:"severity"`
	Title       string                      `json:"title"`
	Description string                      `json:"description"`
	Location    TerraregSecurityLocation    `json:"location"`
}

// TerraregSecurityLocation represents the location of a security issue.
type TerraregSecurityLocation struct {
	Filename  string `json:"filename"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}