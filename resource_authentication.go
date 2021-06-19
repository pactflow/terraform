package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pactflow/terraform/broker"
	"github.com/pactflow/terraform/client"
)

func authentication() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough},
		Create:   authenticationCreate,
		Read:     authenticationRead,
		Update:   authenticationUpdate,
		Delete:   authenticationDelete,
		Schema: map[string]*schema.Schema{
			"github_organizations": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The list of Github organisations allowed access to the account",
			},
			"google_domains": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "The list of Google organisation domains allowed access to the account",
			},
		},
	}
}

func authenticationFromState(d *schema.ResourceData) broker.AuthenticationSettings {
	settings := broker.AuthenticationSettings{
		Providers: broker.AuthenticationProviders{},
	}
	raw, ok := d.Get("github_organizations").([]interface{})

	settings.Providers.Github.Organizations = make([]string, len(raw))
	if ok && len(raw) > 0 {
		for i, s := range raw {
			settings.Providers.Github.Organizations[i] = s.(string)
		}
	}

	raw, ok = d.Get("google_domains").([]interface{})
	settings.Providers.Google.EmailDomains = make([]string, len(raw))
	if ok && len(raw) > 0 {
		for i, s := range raw {
			settings.Providers.Google.EmailDomains[i] = s.(string)
		}
	}

	return settings
}

func authenticationState(d *schema.ResourceData, r *broker.AuthenticationSettingsResponse) error {
	log.Printf("[DEBUG] setting authentication state: %v \n", r)

	if len(r.Providers.Google.EmailDomains) > 0 {
		if err := d.Set("google_domains", r.Providers.Google.EmailDomains); err != nil {
			return fmt.Errorf("error creating key 'google_domains': %w", err)
		}
	}

	if len(r.Providers.Github.Organizations) > 0 {
		if err := d.Set("github_organizations", r.Providers.Github.Organizations); err != nil {
			return fmt.Errorf("error creating key 'github_organizations': %w", err)
		}
	}

	return nil
}

func authenticationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	authentication := authenticationFromState(d)

	created, err := client.SetTenantAuthenticationSettings(authentication)

	if err != nil {
		return fmt.Errorf("error setting authentication: %w", err)
	}

	d.SetId(client.Config.BaseURL.Host)

	if err = authenticationState(d, created); err != nil {
		return fmt.Errorf("error setting authentication state: %w", err)
	}

	return nil
}

func authenticationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	authentication, err := client.ReadTenantAuthenticationSettings()

	if err != nil {
		return fmt.Errorf("error reading authentication settings: %w", err)
	}

	if err = authenticationState(d, authentication); err != nil {
		return fmt.Errorf("error setting authentication settings state: %w", err)
	}

	return nil
}

func authenticationUpdate(d *schema.ResourceData, meta interface{}) error {
	return authenticationCreate(d, meta)
}

func authenticationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)

	log.Println("[DEBUG] deleting (clearing) authentication settings")

	_, err := client.SetTenantAuthenticationSettings(broker.AuthenticationSettings{})

	if err != nil {
		return fmt.Errorf("error deleting authentication: %w", err)
	}

	d.SetId("")

	return nil
}
