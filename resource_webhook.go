package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/pactflow/terraform/broker"
	"github.com/pactflow/terraform/client"
)

var allowedEvents = []string{
	"contract_content_changed",
	"contract_published",
	"provider_verification_failed",
	"provider_verification_published",
	"provider_verification_succeeded",
}

var pacticipantType = &schema.Schema{
	Type:     schema.TypeMap,
	Optional: true,
	Computed: true,
	ForceNew: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A short description of the webhook",
			},
		},
	},
}

var eventsType = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	Elem: &schema.Schema{
		Type:         schema.TypeString,
		ValidateFunc: validateEvents,
	},
}

var requestType = &schema.Schema{
	Type:     schema.TypeList,
	MaxItems: 1,
	Required: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateURL,
				Description:  "A valid URL to send the webhook request to",
			},
			"method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateMethod,
				Description:  "The HTTP method to use with the request",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An optional (basic auth) username to send with the request",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "An optional (basic auth) password to send with the request",
			},
			"headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Request headers to send with the request",
			},
			"body": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "A request body to send with the request",
				DiffSuppressFunc: ignoreJSONFormatting,
			},
		},
	},
}

func stringContains(s []string, searchterm string) bool {
	sort.Strings(s)
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func validateEvents(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if !stringContains(allowedEvents, v) {
		errs = append(errs, fmt.Errorf("%q must be one of the allowed events %v, got %v", key, allowedEvents, v))
	}
	return
}

func validateURL(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	_, err := url.ParseRequestURI(v)

	if err != nil {
		errs = append(errs, fmt.Errorf("%q must be a valid URL, got: %v", key, err))
	}
	return
}

func validateMethod(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	if matched, _ := regexp.MatchString(`^(GET|PUT|PATCH|POST|DELETE)$`, v); !matched {
		errs = append(errs, fmt.Errorf("%q must one of the following HTTP Verbs 'GET, PUT, PATCH, POST, DELETE', got: %s", key, v))
	}
	return
}

func webhook() *schema.Resource {
	return &schema.Resource{
		Create:   webhookCreate,
		Update:   webhookUpdate,
		Read:     webhookRead,
		Delete:   webhookDelete,
		Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"webhook_provider": pacticipantType,
			"webhook_consumer": pacticipantType,
			"request":          requestType,
			"events":           eventsType,
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
		},
	}
}

func parseWebhook(d *schema.ResourceData, meta interface{}) (broker.Webhook, error) {
	request := new(broker.Request)
	webhook := &broker.Webhook{
		Enabled: true,
		Events:  []broker.WebhookEvent{},
	}

	log.Printf("[DEBUG] create or update webhook with data %+v \n", d)

	webhook.Description = d.Get("description").(string)

	// Provider
	if rawProvider, ok := d.GetOkExists("webhook_provider"); ok {
		provider := new(broker.Pacticipant)
		log.Printf("[DEBUG] raw provider %+v \n", rawProvider)
		err := mapstructure.Decode(rawProvider, provider)
		if err != nil {
			log.Println("[ERROR] error decoding webhook config: webhook_provider", err)
			return *webhook, err
		}

		if provider.Name != "" {
			webhook.Provider = provider
		}
	}

	// Consumer
	if rawConsumer, ok := d.GetOkExists("webhook_consumer"); ok {
		consumer := new(broker.Pacticipant)
		log.Printf("[DEBUG] raw consumer %+v \n", rawConsumer)
		err := mapstructure.Decode(rawConsumer, consumer)
		if err != nil {
			log.Println("[ERROR] error decoding webhook config: webhook_consumer", err)
			return *webhook, err
		}

		if consumer.Name != "" {
			webhook.Consumer = consumer
		}
	}

	// Events
	if eventsRaw, ok := d.GetOkExists("events"); ok {
		events := eventsRaw.([]interface{})
		for _, event := range events {
			if event != nil {
				log.Printf("[DEBUG]event item %+v\n", event.(string))
				webhook.Events = append(webhook.Events, broker.WebhookEvent{
					Name: event.(string),
				})
			}
		}
	}

	// Request
	log.Println("[DEBUG] checking request")
	if rawRequest, ok := d.GetOkExists("request"); ok {
		log.Printf("[DEBUG] have raw request of %+v \n", rawRequest)

		rawRequestList := rawRequest.([]interface{})
		requestMap := rawRequestList[0].(map[string]interface{})
		log.Printf("[DEBUG] have converted request %+v \n", requestMap)

		// Method
		if method, ok := requestMap["method"]; ok {
			request.Method = method.(string)
		}

		// Username
		if username, ok := requestMap["username"]; ok {
			request.Username = username.(string)
		}

		// Username
		if password, ok := requestMap["password"]; ok {
			request.Password = password.(string)
		}

		// URL
		if url, ok := requestMap["url"]; ok {
			request.URL = url.(string)
		}

		// Convert headers JSON string into map type
		if headers, ok := requestMap["headers"]; ok {
			request.Headers = make(map[string]string)
			if headers, ok := headers.(map[string]interface{}); ok {
				for k, v := range headers {
					fmt.Println("[DEBUG] Key", k, "Value", v, "Type", reflect.TypeOf(v))
					request.Headers[k] = v.(string)
				}
			} else {
				err := fmt.Errorf("unable parse request headers into a map[string]interface, got %v", reflect.TypeOf(requestMap["headers"]))
				log.Print("[ERROR] error", err)
				return *webhook, err
			}
		} else {
			log.Printf("[ERROR] 'headers' is a required field")
			return *webhook, fmt.Errorf("headers is a mandatory field")
		}

		// Body
		if body, ok := requestMap["body"]; ok {
			// parse JSON into an intermediate object if possible, as this will avoid double escaping of the
			// JSON (e.g. quotes) when it's sent over the wire
			var i interface{}
			err := json.Unmarshal([]byte(body.(string)), &i)
			if err != nil {
				log.Println("[DEBUG] unable to parse JSON, default to string")
				request.Body = body.(string)
			} else {
				request.Body = i
			}
		}

		log.Printf("[DEBUG] have fully serialised request %+v \n", request)

		webhook.Request = *request
	} else {
		log.Println("[ERROR] request attribute not found")
		return *webhook, fmt.Errorf("request is a mandatory field")
	}

	// Existing webhook for update?
	if d.Id() != "" {
		webhook.ID = d.Id()
	}

	return *webhook, nil
}

func setWebhookState(d *schema.ResourceData, webhook broker.Webhook) error {
	log.Printf("[DEBUG] setting webhook state: %+v \n", webhook)
	if err := d.Set("description", webhook.Description); err != nil {
		log.Println("[ERROR] error setting key 'description'", err)
		return err
	}
	if err := d.Set("enabled", webhook.Enabled); err != nil {
		log.Println("[ERROR] error setting key 'enabled'", err)
		return err
	}

	if webhook.Consumer != nil {
		if err := d.Set("webhook_consumer", map[string]interface{}{
			"name": webhook.Consumer.Name,
		}); err != nil {
			log.Println("[ERROR] error setting key 'webhook_consumer'", err)
			return err
		}
	} else {
		d.Set("webhook_consumer", nil)
	}

	if webhook.Provider != nil {
		if err := d.Set("webhook_provider", map[string]interface{}{
			"name": webhook.Provider.Name,
		}); err != nil {
			log.Println("[ERROR] error setting key 'webhook_provider", err)
			return err
		}
	} else {
		d.Set("webhook_provider", nil)
	}

	if err := d.Set("events", flattenEvents(webhook)); err != nil {
		log.Println("[ERROR] error setting key 'events'", err)
		return err
	}

	if err := d.Set("request", flattenRequest(d, webhook.Request)); err != nil {
		log.Println("[ERROR] error setting key 'request'", err)
		return err
	}
	return nil
}

func flattenEvents(w broker.Webhook) []string {
	events := make([]string, len(w.Events), len(w.Events))
	for i, event := range w.Events {
		events[i] = event.Name
	}
	return events
}

func flattenRequest(d *schema.ResourceData, r broker.Request) []interface{} {
	// NOTE: the top level structure to set is a map
	m := make(map[string]interface{})
	m["url"] = r.URL
	m["method"] = r.Method
	m["username"] = r.Username

	if r.Password != "" && !strings.HasPrefix(r.Password, "*****") {
		// First time, set the value
		log.Println("[DEBUG] setting webhook password")
		m["password"] = r.Password
	} else {
		// Broker obscures the value to "******", set the value to what it was previously
		// to prevent it always thinking the value is ""
		if original, ok := d.GetOkExists("request.0.password"); ok {
			m["password"] = original.(string)
		} else {
			log.Println("[DEBUG] could not find original value for 'password'")
		}
	}
	m["headers"] = mapStringStringToMapStringInterface(r.Headers) // TODO

	// We want to store the body as a string in the state file
	// Try to parse body into JSON, fallback to a string if not
	if bodyAsStr, ok := r.Body.(string); ok {
		log.Println("[DEBUG] parsed webhook body as string", bodyAsStr)
		m["body"] = bodyAsStr
	} else if bytes, err := json.Marshal(r.Body); err == nil {
		log.Println("[DEBUG] parsed webhook body as JSON", string(bytes))
		m["body"] = string(bytes)
	} else {
		log.Println("[DEBUG] unable to parse the body as a JSON string or a plain string!")
	}

	return []interface{}{m}
}

// Lowercases all keys
func mapStringStringToMapStringInterface(in map[string]string) map[string]interface{} {
	var out = make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func webhookCreate(d *schema.ResourceData, meta interface{}) error {
	httpClient := meta.(*client.Client)
	webhook, err := parseWebhook(d, meta)
	if err != nil {
		return err
	}

	res, err := httpClient.CreateWebhook(webhook)
	log.Printf("[DEBUG] response from creating webhook %+v\n", res)

	if err == nil {
		items := strings.Split(res.Links["self"].Href, "/")
		id := items[len(items)-1]
		d.SetId(id)

		return setWebhookState(d, webhook)
	}

	log.Println("[ERROR] webhook creation failed", err)
	d.SetId("")
	return err
}

func webhookUpdate(d *schema.ResourceData, meta interface{}) error {
	httpClient := meta.(*client.Client)
	webhook, err := parseWebhook(d, meta)
	if err != nil {
		return err
	}

	res, err := httpClient.UpdateWebhook(webhook)
	log.Printf("[DEBUG] response from updating webhook %+v\n", res)

	if err != nil {
		log.Println("[ERROR] webhook creation failed", err)
		d.SetId("")
	}
	d.Set("webhook", res)

	return nil
}

func webhookRead(d *schema.ResourceData, meta interface{}) error {
	httpClient := meta.(*client.Client)
	res, err := httpClient.ReadWebhook(d.Id())
	log.Printf("[DEBUG] response from reading webhook %+v\n", res)

	if err != nil {
		log.Println("[ERROR] webhook read failed", err)
		d.SetId("")
		return nil
	}
	return setWebhookState(d, *res)
}

func webhookDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] deleting webhook with data %+v\n", d)
	httpClient := meta.(*client.Client)
	webhook, err := parseWebhook(d, meta)
	if err != nil {
		return err
	}

	log.Println("[DEBUG] deleting webhook", webhook)

	err = httpClient.DeleteWebhook(webhook)
	if err == nil {
		d.SetId("")
	}

	return err
}

func tryParseJSONObject(s string) interface{} {
	log.Println("[DEBUG] checking if", s, "is a JSON string")
	var i interface{}
	err := json.Unmarshal([]byte(s), &i)

	if err != nil {
		log.Println("[DEBUG] input body is not JSON")
		return nil
	}

	return i
}

func ignoreJSONFormatting(k, old, new string, d *schema.ResourceData) bool {
	// old = strings.TrimSpace(tryParseJSONString(old))
	// new = strings.TrimSpace(tryParseJSONString(new))
	log.Println("[DEBUG] checking if we should ignore white space and JSON formatting", old, new)

	if tryParseJSONObject(old) != nil && reflect.DeepEqual(tryParseJSONObject(old), tryParseJSONObject(new)) {
		log.Println("[DEBUG] JSON bodies are identical")
		return true
	}

	if strings.TrimSpace(old) == strings.TrimSpace(new) {
		return true
	}

	return false
}
