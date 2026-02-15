package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewExampleResource defines a test resource
type NewExampleResource struct {
	tfsdk.Resource
}

// NewExampleResource returns a test resource
func NewExampleResource() *NewExampleResource {
	return &NewExampleResource{
		Resource: tfsdk.Resource{
			Type:     tfsdk.ManagedResourceType,
			Schema:   tfsdk.Schema{
				Attributes: tfsdk.MapAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Required: true,
						Optional: false,
					},
					"description": {
						Type:     types.StringType,
						Required: false,
						Optional: true,
					},
				}),
			},
		},
		},
	}
}
