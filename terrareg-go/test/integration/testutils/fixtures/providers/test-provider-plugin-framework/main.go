package main

import (
	"github.com/hashicorp/terraform-plugin-framework/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/example/test-provider-plugin-framework/provider"
)

func main() {
	terraform-plugin-framework.Serve(
		"main.New",
	)
}
