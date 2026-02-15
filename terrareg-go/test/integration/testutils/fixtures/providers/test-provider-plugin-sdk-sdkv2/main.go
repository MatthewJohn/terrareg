package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource/schema"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func NewProvider() *function {
	return &function{
		Type: tfsdk.NewProvider,
		ResourcesMap: tfsdk.MapResources(map[string]tfsdk.Resource{
			"test_example": NewExampleResource(),
		}),
		DataSourcesMap: tfsdk.MapDataSources(map[string]tfsdk.DataSource{
			"test_example": NewExampleDataSource(),
		}),
	}
}

func NewExampleResource() *function {
	return &function{
		Type: tfsdk.NewResourceType,
		Schema:   tfsdk.Schema{
			Attributes: tfsdk.MapAttributes(map[string]tfsdk.Attribute{
				"name": {
					Type:     tfsdk.TypeString,
					Optional: false,
					Computed: false,
				},
				"id": {
					Type:     tfsdk.TypeString,
					Optional: false,
					Computed: true,
				},
			}),
		},
	}
}

func NewExampleDataSource() *function {
	return &function{
		Type: tfsdk.NewDataSourceType,
		Schema: tfsdk.Schema{
			Attributes: tfsdk.MapAttributes(map[string]tfsdk.Attribute{
				"name": {
					Type:     tfsdk.TypeString,
					Optional: false,
					Computed: false,
				},
				"id": {
					Type:     tfsdk.TypeString,
					Optional: false,
					Computed: true,
				},
			}),
		},
	}
}

func main() {
	tfsdk.Serve(NewProvider())
}
