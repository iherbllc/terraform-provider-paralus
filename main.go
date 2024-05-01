// Main method used for terraform provider
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/iherbllc/terraform-provider-paralus/internal/provider"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

// main method
func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	err := providerserver.Serve(
		context.Background(),
		provider.New,
		providerserver.ServeOpts{
			Debug:   debug,
			Address: "github.com/iherbllc/paralus",
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
