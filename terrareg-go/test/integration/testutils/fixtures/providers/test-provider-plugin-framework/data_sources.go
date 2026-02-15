package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewExampleDataSource defines a test data source
type NewExampleDataSource struct {
	tfsdk.DataSource
}

// NewExampleDataSource returns a test data source
func NewExampleDataSource() *NewExampleDataSource {
	return &NewExampleDataSource{
		DataSource: tfsdk.DataSource{
			Type:     tfsdk.DataSourceType,
			Schema:   tfsdk.Schema{
				Attributes: tfsdk.MapAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
						Optional: false,
					},
					"id": {
						Type:     types.StringType,
						Computed: true,
						Optional: true,
					},
				}),
			},
		},
		},
	}
}
