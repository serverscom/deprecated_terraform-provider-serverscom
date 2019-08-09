package main

import (
	"github.com/hashicorp/terraform/plugin"
	"servers.com/terraform-provider/provider"
	_ "servers.com/terraform-provider/provider"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.Provider})
}
