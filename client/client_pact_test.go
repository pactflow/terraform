// +build consumer

package client

import (
	"testing"

	"fmt"
	"net/url"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pactflow/terraform/broker"
	"github.com/stretchr/testify/assert"
)

func TestClientPact(t *testing.T) {
	assert.Equal(t, true, true)
}

func TestTerraformClientPact(t *testing.T) {
	log.SetLogLevel("TRACE")

	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "terraform-client",
		Provider: "pactflow-application-saas",
		Host:     "127.0.0.1",
	})
	assert.NoError(t, err)

	t.Run("CreatePacticipant", func(t *testing.T) {

		// Set up our expected interactions.
		mockProvider.
			AddInteraction().
			UponReceiving("a request to create an application").
			WithRequest("POST", matchers.S("/pacticipants")).
			WithHeader("Content-Type", matchers.S("application/json")).
			WithHeader("Authorization", matchers.Like("Bearer 1234")).
			WithBodyMatch(&broker.Pacticipant{}).
			WillRespondWith(200).
			WithHeader("Content-Type", matchers.S("application/hal+json")).
			WithBodyMatch(&broker.Pacticipant{})

			// Execute pact test
		err = mockProvider.ExecuteTest(func(config consumer.MockServerConfig) error {
			client := clientForPact(config)

			p, e := client.CreatePacticipant(broker.Pacticipant{
				Name:          "terraform-consumer",
				RepositoryURL: "https://github.com/pactflow/terraform-provider-pact",
			})
			assert.NoError(t, e)
			assert.Equal(t, "terraform-consumer", p.Name)

			return e
		})
		assert.NoError(t, err)
	})
}

// var clientConfig = Config{
// 	BaseURL:
// }

func clientForPact(config consumer.MockServerConfig) *Client {
	baseURL, err := url.Parse(fmt.Sprintf("http://%s:%d", config.Host, config.Port))
	fmt.Println(baseURL)
	if err != nil {
		panic(fmt.Sprintf("unable to create client for pact test: %s", err))
	}

	return NewClient(nil, Config{
		AccessToken: "1234",
		BaseURL:     baseURL,
	})
}
