package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pactflow/terraform/broker"
	"github.com/pactflow/terraform/client"
)

var secretType = &schema.Schema{
	Type:     schema.TypeMap,
	Optional: true,
	Elem:     &schema.Resource{},
}

func secret() *schema.Resource {
	return &schema.Resource{
		Create:   secretCreate,
		Update:   secretUpdate,
		Read:     secretRead,
		Delete:   secretDelete,
		Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "A short name of the secret (alphanumeric characters only)",
				ValidateFunc: validateName,
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A longer description for the secret",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The actual secret",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UUID of secret",
			},
		},
	}
}

func validateName(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if matched, _ := regexp.MatchString(`[^a-zA-z0-9].*`, v); matched {
		errs = append(errs, fmt.Errorf("%q must be a string containing alphanumeric letters, got: %s", key, v))
	}
	return
}

func parseSecret(d *schema.ResourceData, meta interface{}) (broker.Secret, error) {
	log.Printf("[DEBUG] create or update secret with data %+v \n", d)
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	value := d.Get("value").(string)

	secret := broker.Secret{
		Name:        name,
		Description: description,
		Value:       value,
	}

	// Existing secret?
	if d.Id() != "" {
		secret.UUID = d.Id()
	}

	return secret, nil
}

func secretCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	secret, _ := parseSecret(d, meta)
	log.Println("[DEBUG] creating secret", secret)

	res, err := client.CreateSecret(secret)

	if err == nil {
		items := strings.Split(res.Links["self"].Href, "/")
		id := items[len(items)-1]
		d.SetId(id)

		return setSecretState(d, secret)
	}

	return err
}

func secretUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	secret, _ := parseSecret(d, meta)

	log.Println("[DEBUG] updatding secret", secret)

	_, err := client.UpdateSecret(secret)

	if err == nil {
		return setSecretState(d, secret)
	}

	return err
}

func secretRead(d *schema.ResourceData, meta interface{}) error {
	httpClient := meta.(*client.Client)

	secret, err := httpClient.ReadSecret(d.Id())
	if err != nil {
		return err
	}

	return setSecretState(d, secret.Secret)
}

func secretDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	secret, _ := parseSecret(d, meta)

	log.Println("[DEBUG] deleting secret", secret)

	err := client.DeleteSecret(secret)

	if err == nil {
		d.SetId("")
	}

	return err
}

func setSecretState(d *schema.ResourceData, secret broker.Secret) error {
	log.Printf("[DEBUG] setting secret state: %+v \n", secret)

	d.Set("name", secret.Name)
	d.Set("uuid", secret.UUID)
	d.Set("description", secret.Description)

	if secret.Value != "" {
		// First time, set the value
		d.Set("value", secret.Value)
	} else {
		// Broker does not return the value, to prevent it always thinking the value is ""
		// and requires an update, set to original
		if original, ok := d.GetOkExists("value"); ok {
			d.Set("value", original.(string))
		} else {
			log.Println("[DEBUG] could not find original value for 'value'")
		}
	}

	return nil
}
