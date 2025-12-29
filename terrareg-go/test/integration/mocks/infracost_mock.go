package mocks

import (
	"encoding/json"
	"fmt"
)

// InfracostMock provides a mock implementation of Infracost API output
// This is used for testing module cost estimation without requiring the actual Infracost API
type InfracostMock struct {
	ExpectedEstimate *CostEstimate
	ShouldError      bool
	ErrorMessage     string
}

// CostEstimate represents the cost estimate output from Infracost
type CostEstimate struct {
	Currency          string              `json:"currency"`
	TotalMonthlyCost  float64             `json:"totalMonthlyCost"`
	TotalHourlyCost   float64             `json:"totalHourlyCost"`
	Projects          []CostProject       `json:"projects"`
	Summary           CostSummary         `json:"summary"`
}

// CostProject represents a single project in the cost estimate
type CostProject struct {
	Name              string         `json:"name"`
	TotalMonthlyCost  float64        `json:"totalMonthlyCost"`
	TotalHourlyCost   float64        `json:"totalHourlyCost"`
	Breakdown         []CostResource `json:"breakdown"`
}

// CostResource represents a resource with its cost breakdown
type CostResource struct {
	Name              string               `json:"name"`
	Metadata          CostMetadata         `json:"metadata"`
	MonthlyCost       float64              `json:"monthlyCost"`
	HourlyCost        float64              `json:"hourlyCost"`
	CostComponents    []CostComponent      `json:"costComponents"`
}

// CostMetadata provides additional information about a cost item
type CostMetadata struct {
	Type           string `json:"type,omitempty"`
	Region         string `json:"region,omitempty"`
	UsageType      string `json:"usageType,omitempty"`
	OperatingSystem string `json:"operatingSystem,omitempty"`
}

// CostComponent represents a detailed cost component
type CostComponent struct {
	Name              string  `json:"name"`
	Unit              string  `json:"unit"`
	MonthlyCost       float64 `json:"monthlyCost"`
	HourlyCost        float64 `json:"hourlyCost"`
	Quantity          float64 `json:"quantity"`
	Price             float64 `json:"price"`
}

// CostSummary provides a summary of costs
type CostSummary struct {
	Total    CostSummaryDetail `json:"total"`
	ResourceCounts map[string]int `json:"resourceCounts,omitempty"`
}

// CostSummaryDetail provides detailed summary information
type CostSummaryDetail struct {
	TotalMonthlyCost float64 `json:"totalMonthlyCost"`
	TotalHourlyCost  float64 `json:"totalHourlyCost"`
}

// Estimate returns the infracost output as JSON or an error
func (m *InfracostMock) Estimate() ([]byte, error) {
	if m.ShouldError {
		if m.ErrorMessage != "" {
			return nil, fmt.Errorf("infracost mock error: %s", m.ErrorMessage)
		}
		return nil, fmt.Errorf("infracost mock error")
	}

	// If no expected estimate is set, return a default empty estimate
	if m.ExpectedEstimate == nil {
		m.ExpectedEstimate = &CostEstimate{
			Currency:         "USD",
			TotalMonthlyCost: 0.0,
			TotalHourlyCost:  0.0,
			Projects:         []CostProject{},
			Summary: CostSummary{
				Total: CostSummaryDetail{
					TotalMonthlyCost: 0.0,
					TotalHourlyCost:  0.0,
				},
				ResourceCounts: make(map[string]int),
			},
		}
	}

	jsonOutput, err := json.MarshalIndent(m.ExpectedEstimate, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal infracost output: %w", err)
	}

	return jsonOutput, nil
}

// SetEstimate sets the expected cost estimate for the mock
func (m *InfracostMock) SetEstimate(estimate *CostEstimate) {
	m.ExpectedEstimate = estimate
}

// SetError configures the mock to return an error
func (m *InfracostMock) SetError(message string) {
	m.ShouldError = true
	m.ErrorMessage = message
}

// ClearError clears the error state
func (m *InfracostMock) ClearError() {
	m.ShouldError = false
	m.ErrorMessage = ""
}

// NewInfracostMock creates a new mock with default test data
func NewInfracostMock() *InfracostMock {
	return &InfracostMock{
		ExpectedEstimate: CreateSimpleEC2Estimate(),
		ShouldError:      false,
	}
}

// NewEmptyInfracostMock creates a mock with zero costs
func NewEmptyInfracostMock() *InfracostMock {
	return &InfracostMock{
		ExpectedEstimate: &CostEstimate{
			Currency:         "USD",
			TotalMonthlyCost: 0.0,
			TotalHourlyCost:  0.0,
			Projects:         []CostProject{},
			Summary: CostSummary{
				Total: CostSummaryDetail{
					TotalMonthlyCost: 0.0,
					TotalHourlyCost:  0.0,
				},
				ResourceCounts: make(map[string]int),
			},
		},
		ShouldError: false,
	}
}

// NewErrorInfracostMock creates a mock that always returns an error
func NewErrorInfracostMock(message string) *InfracostMock {
	return &InfracostMock{
		ShouldError:  true,
		ErrorMessage: message,
	}
}

// Helper functions to create common cost estimates

// CreateSimpleEC2Estimate creates a cost estimate for a simple EC2 instance
func CreateSimpleEC2Estimate() *CostEstimate {
	hourlyCost := 0.0116
	monthlyCost := hourlyCost * 730 // 730 hours per month

	return &CostEstimate{
		Currency:         "USD",
		TotalHourlyCost:  hourlyCost,
		TotalMonthlyCost: monthlyCost,
		Projects: []CostProject{
			{
				Name:              "terraform-project",
				TotalHourlyCost:   hourlyCost,
				TotalMonthlyCost:  monthlyCost,
				Breakdown: []CostResource{
					{
						Name:        "aws_instance.web",
						Metadata: CostMetadata{
							Type:  "aws_instance",
							Region: "us-east-1",
						},
						HourlyCost:  hourlyCost,
						MonthlyCost: monthlyCost,
						CostComponents: []CostComponent{
							{
								Name:        "Instance usage (t2.micro)",
								Unit:        "hours",
								HourlyCost:  hourlyCost,
								MonthlyCost: monthlyCost,
								Quantity:    730.0,
								Price:       0.0116,
							},
						},
					},
				},
			},
		},
		Summary: CostSummary{
			Total: CostSummaryDetail{
				TotalHourlyCost:  hourlyCost,
				TotalMonthlyCost: monthlyCost,
			},
			ResourceCounts: map[string]int{
				"aws_instance": 1,
			},
		},
	}
}

// CreateMultiResourceEstimate creates a cost estimate for multiple resources
func CreateMultiResourceEstimate() *CostEstimate {
	instanceHourlyCost := 0.0116
	instanceMonthlyCost := instanceHourlyCost * 730

	eipHourlyCost := 0.005
	eipMonthlyCost := eipHourlyCost * 730

	totalHourlyCost := instanceHourlyCost + eipHourlyCost
	totalMonthlyCost := instanceMonthlyCost + eipMonthlyCost

	return &CostEstimate{
		Currency:         "USD",
		TotalHourlyCost:  totalHourlyCost,
		TotalMonthlyCost: totalMonthlyCost,
		Projects: []CostProject{
			{
				Name:              "terraform-project",
				TotalHourlyCost:   totalHourlyCost,
				TotalMonthlyCost:  totalMonthlyCost,
				Breakdown: []CostResource{
					{
						Name:        "aws_instance.web",
						Metadata: CostMetadata{
							Type:  "aws_instance",
							Region: "us-east-1",
						},
						HourlyCost:  instanceHourlyCost,
						MonthlyCost: instanceMonthlyCost,
						CostComponents: []CostComponent{
							{
								Name:        "Instance usage (t2.micro)",
								Unit:        "hours",
								HourlyCost:  instanceHourlyCost,
								MonthlyCost: instanceMonthlyCost,
								Quantity:    730.0,
								Price:       0.0116,
							},
						},
					},
					{
						Name:        "aws_eip.web",
						Metadata: CostMetadata{
							Type:  "aws_eip",
							Region: "us-east-1",
						},
						HourlyCost:  eipHourlyCost,
						MonthlyCost: eipMonthlyCost,
						CostComponents: []CostComponent{
							{
								Name:        "EIP IPv4 address",
								Unit:        "hours",
								HourlyCost:  eipHourlyCost,
								MonthlyCost: eipMonthlyCost,
								Quantity:    730.0,
								Price:       0.005,
							},
						},
					},
				},
			},
		},
		Summary: CostSummary{
			Total: CostSummaryDetail{
				TotalHourlyCost:  totalHourlyCost,
				TotalMonthlyCost: totalMonthlyCost,
			},
			ResourceCounts: map[string]int{
				"aws_instance": 1,
				"aws_eip":      1,
			},
		},
	}
}

// CreateS3BucketEstimate creates a cost estimate for S3 bucket storage
func CreateS3BucketEstimate() *CostEstimate {
	// 1 GB storage at standard tier
	gbStorageCost := 0.023 // per GB per month

	// Request costs (1000 PUT requests, 10000 GET requests)
	putRequestCost := 0.000005 // per 1000 requests
	getRequestCost := 0.0000004 // per 1000 requests

	monthlyStorageCost := gbStorageCost
	monthlyRequestCost := (putRequestCost * 1) + (getRequestCost * 10)
	totalMonthlyCost := monthlyStorageCost + monthlyRequestCost

	return &CostEstimate{
		Currency:         "USD",
		TotalHourlyCost:  0.0,
		TotalMonthlyCost: totalMonthlyCost,
		Projects: []CostProject{
			{
				Name:              "terraform-project",
				TotalHourlyCost:   0.0,
				TotalMonthlyCost:  totalMonthlyCost,
				Breakdown: []CostResource{
					{
						Name:        "aws_s3_bucket.data",
						Metadata: CostMetadata{
							Type:      "aws_s3_bucket",
							Region:    "us-east-1",
							UsageType: "AWS-S3-Standard",
						},
						HourlyCost:  0.0,
						MonthlyCost: totalMonthlyCost,
						CostComponents: []CostComponent{
							{
								Name:        "Standard storage",
								Unit:        "GB-Mo",
								HourlyCost:  0.0,
								MonthlyCost: monthlyStorageCost,
								Quantity:    1.0,
								Price:       0.023,
							},
							{
								Name:        "PUT requests",
								Unit:        "1000 requests",
								HourlyCost:  0.0,
								MonthlyCost: putRequestCost,
								Quantity:    1.0,
								Price:       0.000005,
							},
							{
								Name:        "GET requests",
								Unit:        "1000 requests",
								HourlyCost:  0.0,
								MonthlyCost: getRequestCost * 10,
								Quantity:    10.0,
								Price:       0.0000004,
							},
						},
					},
				},
			},
		},
		Summary: CostSummary{
			Total: CostSummaryDetail{
				TotalHourlyCost:  0.0,
				TotalMonthlyCost: totalMonthlyCost,
			},
			ResourceCounts: map[string]int{
				"aws_s3_bucket": 1,
			},
		},
	}
}

// CreateRDSInstanceEstimate creates a cost estimate for RDS instance
func CreateRDSInstanceEstimate() *CostEstimate {
	// db.t3.micro instance (multi-az)
	hourlyCost := 0.032
	monthlyCost := hourlyCost * 730

	// Storage (20 GB general purpose SSD)
	storageCostPerGB := 0.115 // per GB per month
	storageMonthlyCost := 20 * storageCostPerGB

	totalHourlyCost := hourlyCost
	totalMonthlyCost := monthlyCost + storageMonthlyCost

	return &CostEstimate{
		Currency:         "USD",
		TotalHourlyCost:  totalHourlyCost,
		TotalMonthlyCost: totalMonthlyCost,
		Projects: []CostProject{
			{
				Name:              "terraform-project",
				TotalHourlyCost:   totalHourlyCost,
				TotalMonthlyCost:  totalMonthlyCost,
				Breakdown: []CostResource{
					{
						Name:        "aws_db_instance.main",
						Metadata: CostMetadata{
							Type:           "aws_db_instance",
							Region:         "us-east-1",
							UsageType:      "Aurora",
							OperatingSystem: "Linux",
						},
						HourlyCost:  totalHourlyCost,
						MonthlyCost: totalMonthlyCost,
						CostComponents: []CostComponent{
							{
								Name:        "DB instance (db.t3.micro, multi-az)",
								Unit:        "hours",
								HourlyCost:  hourlyCost,
								MonthlyCost: monthlyCost,
								Quantity:    730.0,
								Price:       0.032,
							},
							{
								Name:        "Database storage (gp3)",
								Unit:        "GB-Mo",
								HourlyCost:  0.0,
								MonthlyCost: storageMonthlyCost,
								Quantity:    20.0,
								Price:       0.115,
							},
						},
					},
				},
			},
		},
		Summary: CostSummary{
			Total: CostSummaryDetail{
				TotalHourlyCost:  totalHourlyCost,
				TotalMonthlyCost: totalMonthlyCost,
			},
			ResourceCounts: map[string]int{
				"aws_db_instance": 1,
			},
		},
	}
}

// CreateHighCostEstimate creates a cost estimate for expensive resources
func CreateHighCostEstimate() *CostEstimate {
	// m5.2xlarge instance (expensive)
	hourlyCost := 0.384
	monthlyCost := hourlyCost * 730

	return &CostEstimate{
		Currency:         "USD",
		TotalHourlyCost:  hourlyCost,
		TotalMonthlyCost: monthlyCost,
		Projects: []CostProject{
			{
				Name:              "terraform-project",
				TotalHourlyCost:   hourlyCost,
				TotalMonthlyCost:  monthlyCost,
				Breakdown: []CostResource{
					{
						Name:        "aws_instance.app_server",
						Metadata: CostMetadata{
							Type:  "aws_instance",
							Region: "us-west-2",
						},
						HourlyCost:  hourlyCost,
						MonthlyCost: monthlyCost,
						CostComponents: []CostComponent{
							{
								Name:        "Instance usage (m5.2xlarge)",
								Unit:        "hours",
								HourlyCost:  hourlyCost,
								MonthlyCost: monthlyCost,
								Quantity:    730.0,
								Price:       0.384,
							},
						},
					},
				},
			},
		},
		Summary: CostSummary{
			Total: CostSummaryDetail{
				TotalHourlyCost:  hourlyCost,
				TotalMonthlyCost: monthlyCost,
			},
			ResourceCounts: map[string]int{
				"aws_instance": 1,
			},
		},
	}
}
