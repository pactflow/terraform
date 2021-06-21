package broker

// WebhookEvent represents the types of events that trigger a Webhook
type WebhookEvent struct {
	Name string `json:"name"`
}

// Request is an HTTP request structure
type Request struct {
	Method   string      `json:"method,omitempty"`
	URL      string      `json:"url,omitempty"`
	Username string      `json:"username,omitempty"`
	Password string      `json:"password,omitempty"`
	Headers  Headers     `json:"headers,omitempty"`
	Body     interface{} `json:"body,omitempty"`
}

// Webhook represents a webhook configured in the broker
type Webhook struct {
	ID          string         `json:"id,omitempty"`
	Description string         `json:"description,omitempty"`
	Enabled     bool           `json:"enabled,omitempty"`
	CreatedAt   string         `json:"createdAt,omitempty"`
	Provider    *Pacticipant   `json:"provider,omitempty"`
	Consumer    *Pacticipant   `json:"consumer,omitempty"`
	Events      []WebhookEvent `json:"events,omitempty"`
	Request     Request        `json:"request,omitempty"`
}

// WebhookResponse is the response body for any CRU methods
type WebhookResponse struct {
	Webhook
	HalDoc
}
