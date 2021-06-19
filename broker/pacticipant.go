package broker

type Pacticipant struct {
	Name          string `json:"name,omitempty" pact:"example=terraform-consumer"`
	RepositoryURL string `json:"repositoryUrl,omitempty" pact:"example=https://github.com/pactflow/terraform-provider-pact"`
}
