package provider

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type Config  struct {
	Url			  string
	Email     	  string
	Pwd     	  string
}


func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		Url: d.Get("url").(string),
		Email: d.Get("email").(string),
		Pwd: d.Get("password").(string),
	}
	return config.Client()
}

type Client struct {
	Url			  string
	Email     	  string
	Token         string
	Pwd     	  string
}

func (c *Config) Client() (interface{}, error) {
	t, err := GetToken(c.Url, c.Email, c.Pwd)
	if err != nil {
		return nil, err
	}
	client := &Client{Url:c.Url, Email:c.Email, Token:t, Pwd:c.Pwd}
	return client, nil
}
