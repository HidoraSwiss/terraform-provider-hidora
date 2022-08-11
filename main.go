package main

import (
	"fmt"

	"github.com/HidoraSwiss/terraform-provider-hidora/hidora"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return hidora.Provider()
		},
	})
	fmt.Print("Hello")
}
