package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

type DataSourceModel struct {
	ExampleField types.StringType `tfsdk:"json:"example_field"`
}

type DataSource struct {
	tfsdk.DataSource
	Schema   DataSourceModel
}

// NewExampleDataSource returns a test data source for documentation extraction
func NewExampleDataSource() *DataSource {
	return &DataSource{
		DataSourceType: tfsdk.NewDataSourceType("test_example"),
		Schema:      DataSourceModel{
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
				"description": {
					Type:     tfsdk.TypeString,
					Optional: true,
				},
			}),
		},
	}
}
