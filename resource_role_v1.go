package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pactflow/terraform/broker"
	"github.com/pactflow/terraform/client"
)

const (
	administratorRole = "administrator"
	userRole          = "user"
)

var allowedRoles = map[string]string{
	administratorRole: "cf75d7c2-416b-11ea-af5e-53c3b1a4efd8",
}

func validateRoles(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if _, ok := allowedRoles[v]; !ok {
		errs = append(errs, fmt.Errorf("%q must be one of the allowed pre-existing roles %v, got %v", key, allowedRoles, v))
	}

	return
}

// TODO: update to use new API? Or just remove this entirely in favour of the new role assignment resource?
func roleV1() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated. Please update to the newer 'pact_role' resource",
		Create:             roleV1Create,
		Read:               roleV1Read,
		Delete:             roleV1Delete,
		Schema: map[string]*schema.Schema{
			"role": {
				Type:         schema.TypeString,
				Description:  "Role to apply to the user",
				ValidateFunc: validateRoles,
				Required:     true,
				ForceNew:     true,
			},
			"user": {
				Type:        schema.TypeString,
				Description: "UUID of the user of which to apply the role",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the Role",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UUID of API token",
			},
		},
	}
}

func roleV1Create(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	userUUID := d.Get("user").(string)

	// NOTE: we only support the admin role at this time
	log.Println("[DEBUG] creating role for user with UUID:", userUUID)
	_, err := client.AddAdminRoleToUser(broker.User{
		UUID: userUUID,
	})

	if err == nil {
		d.SetId(allowedRoles["administrator"])
		d.Set("name", "Administrator")
	}

	return err
}

func roleV1Read(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func roleV1Delete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	userUUID := d.Get("user").(string)

	log.Println("[DEBUG] deleting role for user with UUID:", userUUID)

	_, err := client.RemoveAdminRoleFromUser(broker.User{
		UUID: userUUID,
	})

	if err != nil {
		d.SetId("")
	}

	return err
}
