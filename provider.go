package main

import (
	"crypto/tls"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pactflow/terraform/client"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"pact_role":           role(),
			"pact_role_v1":        roleV1(),
			"pact_team":           team(),
			"pact_user":           user(),
			"pact_application":    application(),
			"pact_pacticipant":    application(),
			"pact_webhook":        webhook(),
			"pact_secret":         secret(),
			"pact_token":          token(),
			"pact_authentication": authentication(),
		},
		ConfigureFunc: configureProvider,
		Schema: map[string]*schema.Schema{
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An API Bearer token to authenticate to a Pactflow account (for Pactflow users only)",
			},
			"basic_auth_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A basic auth username to authenticate to a Pact Broker (not required for Pactflow users)",
			},
			"basic_auth_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A basic auth password to authenticate to a Pact Broker (not required for Pactflow users)",
			},
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A fully qualified hostname (e.g. for a Pactflow account https://mybroker.pact.dius.com.au",
			},
			"tls_insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Disable TLS verification checks for privately hosted brokers",
			},
		},
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	baseURL, err := url.Parse(d.Get("host").(string))
	return client.NewClient(nil, client.Config{
		AccessToken:       d.Get("access_token").(string),
		BasicAuthUsername: d.Get("basic_auth_username").(string),
		BasicAuthPassword: d.Get("basic_auth_password").(string),
		CustomTLSConfig: &tls.Config{
			InsecureSkipVerify: d.Get("tls_insecure").(bool),
		},
		BaseURL: baseURL,
	}), err
}
