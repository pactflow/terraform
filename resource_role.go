package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pactflow/terraform/broker"
	"github.com/pactflow/terraform/client"
)

func role() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough},
		Create:   roleCreate,
		Read:     roleRead,
		Update:   roleUpdate,
		Delete:   roleDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Role",
			},
			"scopes": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "The pre-defined scope to add to the role",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the Role",
			},
		},
	}
}

func validateScopes(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	for _, scope := range broker.AllowedScopes {
		if scope == v {
			return
		}
	}
	errs = append(errs, fmt.Errorf("%q must be one of the allowed scopes %v, got %v", key, broker.AllowedScopes, v))

	return
}

func getRoleFromState(d *schema.ResourceData) broker.Role {
	name := d.Get("name").(string)
	raw, ok := d.Get("scopes").([]interface{})

	permissions := make([]broker.Permission, len(raw))
	if ok && len(raw) > 0 {
		for i, s := range raw {
			permissions[i] = broker.Permission{
				Scope: s.(string),
			}
		}
	}

	return broker.Role{
		UUID:        d.Id(),
		Name:        name,
		Permissions: permissions,
	}
}

func scopesFromUser(u broker.Role) []string {
	scopes := make([]string, len(u.Permissions))

	for i, p := range u.Permissions {
		scopes[i] = p.Scope
	}

	return scopes
}

func setRoleState(d *schema.ResourceData, r *broker.Role) error {
	log.Printf("[DEBUG] setting role state: %v \n", r)

	if err := d.Set("uuid", r.UUID); err != nil {
		return fmt.Errorf("error creating key 'uuid': %w", err)
	}
	if err := d.Set("name", r.Name); err != nil {
		return fmt.Errorf("error creating key 'name': %w", err)
	}
	if err := d.Set("scopes", scopesFromUser(*r)); err != nil {
		return fmt.Errorf("error creating key 'scopes': %w", err)
	}

	return nil
}

func roleCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	role := getRoleFromState(d)

	created, err := client.CreateRole(role)

	if err != nil {
		return fmt.Errorf("error creating role: %w", err)
	}

	d.SetId(created.UUID)

	if err = setRoleState(d, created); err != nil {
		return fmt.Errorf("error setting role state: %w", err)
	}

	return nil
}

func roleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	role, err := client.ReadRole(d.Id())

	if err != nil {
		return fmt.Errorf("error reading role: %w", err)
	}

	if err = setRoleState(d, role); err != nil {
		return fmt.Errorf("error setting role state: %w", err)
	}

	return nil
}

func roleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	role := getRoleFromState(d)
	updated, err := client.UpdateRole(role)

	if err != nil {
		return fmt.Errorf("error updating role: %w", err)
	}

	if err = setRoleState(d, updated); err != nil {
		return fmt.Errorf("error setting role state: %w", err)
	}

	return nil
}

func roleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	uuid := d.Get("uuid").(string)

	log.Println("[DEBUG] deleting role for user with UUID:", uuid)

	err := client.DeleteRole(broker.Role{
		UUID: uuid,
	})

	if err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}

	d.SetId("")

	return nil
}
