package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/iherbllc/terraform-provider-paralus/internal/provider"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debug,
		ProviderAddr: "github.com/iherbllc/terraform-provider-paralus",
		ProviderFunc: func() *schema.Provider {
			return provider.Provider()
		},
	}

	plugin.Serve(opts)
}
