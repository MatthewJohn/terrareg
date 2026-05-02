package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource/schema"
)

type ResourceModel struct {
	ExampleField types.StringType `tfsdk:"json:"example_field"`
}

type Resource struct {
	tfsdk.Resource
	Schema   ResourceModel
}

// NewExampleResource returns a test resource for documentation extraction
func NewExampleResource() *Resource {
	return &Resource{
		ResourceType: tfsdk.NewResourceType("test_example"),
		Schema:      ResourceModel{
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
