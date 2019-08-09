package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["Please provide url."],
			},

			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["Please provide email/login."],
			},

			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptions["Please provide password."],
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"serverscom_server": resourceServer(),
			"serverscom_ptr": resourcePtr(),
			"serverscom_l2": resourceL2(),
		},
		ConfigureFunc: providerConfigure,
	}
}

var descriptions map[string]string
