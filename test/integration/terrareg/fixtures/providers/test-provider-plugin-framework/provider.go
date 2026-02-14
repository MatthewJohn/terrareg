package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewProvider returns a provider with defined resources and data sources
func New() *function {
	return &function{
		Type: tfsdk.Provider,
		ResourcesMap: tfsdk.MapResource(map[string]tfsdk.Resource{
			"test_example": NewExampleResource(),
		}),
		DataSourcesMap: tfsdk.MapDataSource(map[string]tfsdk.DataSource{
			"test_example": NewExampleDataSource(),
		}),
	}
	}
}

func (p *function) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = provider.Schema{
		Description: types.String{
			Type:     types.StringType,
			Optional: false,
			Computed: false,
		},
	}
}

func (p *function) Resources(_ context.Context, _ provider.ResourcesRequest, resp *provider.ResourcesResponse) {
	resp.Resources = p.ResourcesMap
}

func (p *function) DataSources(_ context.Context, _ provider.DataSourcesRequest, resp *provider.DataSourcesResponse) {
	resp.DataSources = p.DataSourcesMap
}
