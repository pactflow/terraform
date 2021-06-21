package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pactflow/terraform/broker"
	"github.com/pactflow/terraform/client"
)

const (
	readOnlyTokenType  = "read-only"
	readWriteTokenType = "read-write"
)

var allowedTokenTypes = map[string]string{
	readOnlyTokenType:  "Read only token (developer)",
	readWriteTokenType: "Read/write token (CI)",
}

var tokenType = &schema.Schema{
	Type:     schema.TypeMap,
	Optional: true,
	Elem:     &schema.Resource{},
}

// Used to convert from TF configuration to a broker.APIToken
type apiTokenDefinition struct {
	UUID        string
	Name        string
	Type        string
	Description string
	Value       string
}

func token() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated and will soon be removed",
		Create:             tokenCreate,
		Update:             tokenUpdate,
		Read:               tokenRead,
		Delete:             tokenDelete,
		Importer:           &schema.ResourceImporter{State: schema.ImportStatePassthrough},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A short name for the token. Changing the name will generate a _new_ token each time, removing the existing token",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The type of token to generate (valid values are 'read-only' and 'read-write'",
				ValidateFunc: validateTokenType,
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the token as defined by the broker",
			},
			"value": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The actual API token",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UUID of API token",
			},
		},
	}
}

func validateTokenType(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if _, ok := allowedTokenTypes[v]; !ok {
		errs = append(errs, fmt.Errorf("%q must be one of the allowed types %v, got %v", key, allowedTokenTypes, v))
	}
	return
}

func parseToken(d *schema.ResourceData, meta interface{}) (apiTokenDefinition, error) {
	log.Printf("[DEBUG] create or update token with data %+v \n", d)
	name := d.Get("name").(string)
	tokenType := d.Get("type").(string)
	description := d.Get("description").(string)
	value := d.Get("value").(string)

	token := apiTokenDefinition{
		UUID:        d.Id(),
		Name:        name,
		Type:        tokenType,
		Description: description,
		Value:       value,
	}

	log.Printf("[DEBUG] Have a parsed token %+v \n", token)

	return token, nil
}

// Basically just does a regenerate
func tokenCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	token, _ := parseToken(d, meta)

	// If token UUID is empty, read from remote
	if token.UUID == "" {
		log.Println("[DEBUG] importing resource as no existing UUID was found")
		t, err := client.FindTokenByType(token.Type)
		if err != nil {
			return err
		}
		token.UUID = t.UUID
		// TF definition specific fields
		d.SetId(t.UUID)
		d.Set("type", token.Type)
		d.Set("name", token.Name)

		// Set the (remote state) resource specific fields
		return setTokenState(d, *t)
	}

	return nil
}

// Regenerate
func tokenUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	token, _ := parseToken(d, meta)

	log.Println("[DEBUG] updating (regenearting) token", token)

	updatedToken, err := client.RegenerateToken(broker.APIToken{UUID: token.UUID})

	// At the moment, if you regenerate the access token - you need to use it for new requests!
	if token.Type == readWriteTokenType {
		if client.Config.AccessToken != "" {
			log.Println("[INFO] updating access token as read-write token was re-generated")
			client.Config.AccessToken = updatedToken.Value
		}
	}

	if err != nil {
		return err
	}

	// TF definition specific fields
	d.Set("type", token.Type)
	d.Set("name", token.Name)

	return setTokenState(d, updatedToken.APIToken)
}

func tokenRead(d *schema.ResourceData, meta interface{}) error {
	httpClient := meta.(*client.Client)
	uuid := d.Id()

	token, err := httpClient.ReadToken(uuid)
	if err != nil {
		return err
	}
	return setTokenState(d, *token)
}

// Uncouples from broker
func tokenDelete(d *schema.ResourceData, meta interface{}) error {

	log.Println("[INFO] Deleting API token is currently a no-op, setting id to ''")
	d.SetId("")

	return nil
}

func setTokenState(d *schema.ResourceData, token broker.APIToken) error {
	log.Printf("[DEBUG] setting token state: %+v \n", token)

	d.SetId(token.UUID)
	d.Set("description", token.Description)
	d.Set("uuid", token.UUID)
	d.Set("value", token.Value)
	return nil
}
