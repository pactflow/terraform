// +build consumer

package client

import (
	"testing"

	"fmt"
	"net/url"

	"github.com/pactflow/terraform/broker"
	"github.com/stretchr/testify/assert"

	. "github.com/pact-foundation/pact-go/v2/sugar"
)

func TestClientPact(t *testing.T) {
	assert.Equal(t, true, true)
}

func TestTerraformClientPact(t *testing.T) {
	SetLogLevel("ERROR")

	mockProvider, err := NewV2Pact(MockHTTPProviderConfig{
		Consumer: "terraform-client",
		Provider: "pactflow-application-saas",
		Host:     "127.0.0.1",
	})
	assert.NoError(t, err)

	pacticipant := broker.Pacticipant{
		Name:          "terraform-client",
		RepositoryURL: "https://github.com/pactflow/new-terraform-provider-pact",
	}

	t.Run("Pacticipant", func(t *testing.T) {
		t.Run("CreatePacticipant", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				UponReceiving("a request to create a pacticipant").
				WithRequest("POST", S("/pacticipants")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithBodyMatch(&broker.Pacticipant{}).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithBodyMatch(&broker.Pacticipant{})

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				p, e := client.CreatePacticipant(broker.Pacticipant{
					Name:          "terraform-client",
					RepositoryURL: "https://github.com/pactflow/terraform-provider-pact",
				})
				assert.NoError(t, e)
				assert.Equal(t, "terraform-client", p.Name)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("ReadPacticipant", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a pacticipant with name terraform-client exists").
				UponReceiving("a request to get a pacticipant").
				WithRequest("GET", S("/pacticipants/terraform-client")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithBodyMatch(&broker.Pacticipant{})

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				p, e := client.ReadPacticipant("terraform-client")
				assert.NoError(t, e)
				assert.Equal(t, "terraform-client", p.Name)
				assert.NotEmpty(t, p.RepositoryURL)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("UpdatePacticipant", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a pacticipant with name terraform-client exists").
				UponReceiving("a request to update a pacticipant").
				WithRequest("PATCH", S("/pacticipants/terraform-client")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(pacticipant)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithBodyMatch(&broker.Pacticipant{})

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				p, e := client.UpdatePacticipant(pacticipant)
				assert.NoError(t, e)
				assert.Equal(t, "terraform-client", p.Name)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("DeletePacticipant", func(t *testing.T) {
			newPacticipant := broker.Pacticipant{
				Name:          "terraform-client",
				RepositoryURL: "https://github.com/pactflow/new-terraform-provider-pact",
			}

			mockProvider.
				AddInteraction().
				Given("a pacticipant with name terraform-client exists").
				UponReceiving("a request to delete a pacticipant").
				WithRequest("DELETE", S("/pacticipants/terraform-client")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WillRespondWith(200)

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				return client.DeletePacticipant(newPacticipant)
			})
			assert.NoError(t, err)
		})
	})

	t.Run("Team", func(t *testing.T) {
		team := broker.Team{
			Name: "terraform-team",
			Embedded: broker.TeamEmbeddedItems{
				Pacticipants: []broker.Pacticipant{
					{
						Name: "terraform-client",
					},
				},
			},
		}

		update := broker.Team{
			Name: "terraform-team",
			UUID: "1234",
			Embedded: broker.TeamEmbeddedItems{
				Pacticipants: []broker.Pacticipant{
					pacticipant,
				},
				Members: []broker.User{
					{
						UUID:   "4c260344-b170-41eb-b01e-c0ff10c72f25",
						Active: true,
					},
				},
			},
		}

		t.Run("ReadTeam", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a pacticipant with name terraform-client exists").
				UponReceiving("a request to get a team").
				WithRequest("GET", S("/admin/teams/terraform-team")). // NOTE: other resources use the UUID
				WithHeader("Authorization", Like("Bearer 1234")).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(update))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, e := client.ReadTeam("terraform-team")

				assert.NoError(t, e)
				assert.NotNil(t, res)
				assert.Equal(t, "terraform-team", res.Name)
				assert.Equal(t, "1234", res.UUID)
				assert.Len(t, res.Embedded.Members, 1)
				assert.Len(t, res.Embedded.Pacticipants, 1)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("CreateTeam", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a pacticipant with name terraform-client exists").
				UponReceiving("a request to create a team").
				WithRequest("POST", S("/admin/teams")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(team)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(broker.TeamsResponse{
					Teams: []broker.Team{
						update,
					},
				}))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, e := client.CreateTeam(team)

				assert.NoError(t, e)
				assert.NotNil(t, res)
				assert.Equal(t, "terraform-team", res.Name)
				assert.Equal(t, "1234", res.UUID)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("UpdateTeam", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a team with uuid 1234 exists").
				UponReceiving("a request to update a team").
				WithRequest("PUT", S("/admin/teams/1234")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(update)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(update))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				p, e := client.UpdateTeam(update)

				assert.NoError(t, e)
				assert.Equal(t, "terraform-team", p.Name)
				assert.Len(t, p.Embedded.Pacticipants, 1)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("DeleteTeam", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a team with name terraform-team exists").
				UponReceiving("a request to delete a pacticipant").
				WithRequest("DELETE", S("/admin/teams/1234")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WillRespondWith(200)

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				return client.DeleteTeam(update)
			})
			assert.NoError(t, err)
		})

		t.Run("UpdateTeamAssignments", func(t *testing.T) {
			req := broker.TeamsAssignmentRequest{
				UUID: update.UUID,
				Users: []string{
					"05064a18-229d-4dfd-b37c-f00ec9673a49",
				},
			}

			mockProvider.
				AddInteraction().
				Given("a team with name terraform-team and user with uuid 05064a18-229d-4dfd-b37c-f00ec9673a49 exists").
				UponReceiving("a request to update team assignments").
				WithRequest("PUT", S("/admin/teams/1234/users")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(req)).
				WillRespondWith(200).
				WithJSONBody(broker.TeamsAssignmentResponse{
					Embedded: broker.EmbeddedUsers{
						Users: []broker.User{
							{
								UUID:   "4c260344-b170-41eb-b01e-c0ff10c72f25",
								Active: true,
							},
						},
					},
				})

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, err := client.UpdateTeamAssignments(req)

				assert.Len(t, res.Embedded.Users, 1)

				return err
			})
			assert.NoError(t, err)
		})
	})

	t.Run("Secret", func(t *testing.T) {
		secret := broker.Secret{
			Name:        "terraform-secret",
			Description: "terraform secret",
			Value:       "supersecret",
		}

		created := broker.Secret{
			UUID:        "b6af03cd-018c-4f1b-9546-c778d214f305",
			Name:        secret.Name,
			Description: secret.Description,
		}

		update := broker.Secret{
			UUID:        created.UUID,
			Name:        secret.Name,
			Description: "updated description",
			Value:       "supersecret",
		}

		t.Run("CreateSecret", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				UponReceiving("a request to create a secret").
				WithRequest("POST", S("/secrets")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(secret)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(created))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				p, e := client.CreateSecret(secret)
				assert.NoError(t, e)
				assert.Equal(t, "terraform-secret", p.Name)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("UpdateSecret", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a secret with uuid b6af03cd-018c-4f1b-9546-c778d214f305 exists").
				UponReceiving("a request to update a secret").
				WithRequest("PUT", S("/secrets/b6af03cd-018c-4f1b-9546-c778d214f305")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(update)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(update))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				p, e := client.UpdateSecret(update)
				assert.NoError(t, e)
				assert.Equal(t, "terraform-secret", p.Name)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("DeleteSecret", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a secret with uuid b6af03cd-018c-4f1b-9546-c778d214f305 exists").
				UponReceiving("a request to delete a secret").
				WithRequest("DELETE", S("/secrets/b6af03cd-018c-4f1b-9546-c778d214f305")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WillRespondWith(200)

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				return client.DeleteSecret(created)
			})
			assert.NoError(t, err)
		})
	})

	t.Run("Role", func(t *testing.T) {
		role := broker.Role{
			Name: "terraform-role",
			Permissions: []broker.Permission{
				{
					Name:        "role name",
					Scope:       "user:manage:*",
					Label:       "role label",
					Description: "role description",
				},
			},
		}

		created := broker.Role{
			UUID:        "e1407277-2a25-4559-8fed-4214dd12a1e8",
			Name:        role.Name,
			Permissions: role.Permissions,
		}

		update := broker.Role{
			UUID:        created.UUID,
			Name:        role.Name,
			Permissions: created.Permissions,
		}

		t.Run("CreateRole", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				UponReceiving("a request to create a role").
				WithRequest("POST", S("/admin/roles")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(role)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(created))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, e := client.CreateRole(role)
				assert.NoError(t, e)
				assert.Equal(t, "terraform-role", res.Name)
				assert.Len(t, res.Permissions, 1)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("ReadRole", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				UponReceiving("a request to get a role").
				WithRequest("GET", S("/admin/roles/e1407277-2a25-4559-8fed-4214dd12a1e8")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(created))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, e := client.ReadRole(created.UUID)
				assert.NoError(t, e)
				assert.Equal(t, "terraform-role", res.Name)
				assert.Len(t, res.Permissions, 1)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("UpdateRole", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a role with uuid e1407277-2a25-4559-8fed-4214dd12a1e8 exists").
				UponReceiving("a request to update a role").
				WithRequest("PUT", S("/admin/roles/e1407277-2a25-4559-8fed-4214dd12a1e8")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(update)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(update))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, e := client.UpdateRole(update)
				assert.NoError(t, e)
				assert.Equal(t, "terraform-role", res.Name)
				assert.Len(t, res.Permissions, 1)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("DeleteRole", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a role with uuid e1407277-2a25-4559-8fed-4214dd12a1e8 exists").
				UponReceiving("a request to delete a role").
				WithRequest("DELETE", S("/admin/roles/e1407277-2a25-4559-8fed-4214dd12a1e8")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WillRespondWith(200)

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				return client.DeleteRole(created)
			})
			assert.NoError(t, err)
		})
	})

	t.Run("User", func(t *testing.T) {
		user := broker.User{
			Name:   "terraform user",
			Email:  "terraform.user@some.domain",
			Active: true,
			Type:   broker.RegularUser,
		}

		created := broker.User{
			UUID:   "819f6dbf-dd7a-47ff-b369-e3ed1d2578a0",
			Name:   user.Name,
			Email:  user.Email,
			Active: user.Active,
			Type:   user.Type,
			Embedded: struct {
				Roles []broker.Role `json:"roles,omitempty"`
				Teams []broker.Team `json:"teams,omitempty"`
			}{
				Roles: []broker.Role{
					{
						Name: "terraform-role",
						UUID: "84f66fab-1c42-4351-96bf-88d3a09d7cd2",
						Permissions: []broker.Permission{
							{
								Name:        "role name",
								Scope:       "user:manage:*",
								Label:       "role label",
								Description: "role description",
							},
						},
					},
				},
			},
		}

		update := created

		setUserRoles := broker.SetUserRolesRequest{
			Roles: []string{
				"84f66fab-1c42-4351-96bf-88d3a09d7cd2",
			},
		}

		t.Run("CreateUser", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				UponReceiving("a request to create a user").
				WithRequest("POST", S("/admin/users/invite-user")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(user)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(created))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, e := client.CreateUser(user)
				assert.NoError(t, e)
				assert.Equal(t, "terraform user", res.Name)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("ReadUser", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				UponReceiving("a request to get a user").
				WithRequest("GET", S("/admin/users/819f6dbf-dd7a-47ff-b369-e3ed1d2578a0")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(created))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, e := client.ReadUser(created.UUID)
				assert.NoError(t, e)
				assert.Equal(t, "terraform user", res.Name)
				assert.Len(t, res.Embedded.Roles, 1)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("UpdateUser", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a user with uuid 819f6dbf-dd7a-47ff-b369-e3ed1d2578a0 exists").
				UponReceiving("a request to update a user").
				WithRequest("PUT", S("/admin/users/819f6dbf-dd7a-47ff-b369-e3ed1d2578a0")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(update)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(update))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				res, e := client.UpdateUser(update)
				assert.NoError(t, e)
				assert.Equal(t, "terraform user", res.Name)
				assert.Len(t, res.Embedded.Roles, 1)

				return e
			})
			assert.NoError(t, err)
		})

		t.Run("DeleteUser", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a user with uuid 819f6dbf-dd7a-47ff-b369-e3ed1d2578a0 exists").
				UponReceiving("a request to delete a user").
				WithRequest("PUT", S("/admin/users/819f6dbf-dd7a-47ff-b369-e3ed1d2578a0")).
				WithHeader("Content-Type", S("application/json")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(update)).
				WillRespondWith(200).
				WithHeader("Content-Type", S("application/hal+json")).
				WithJSONBody(Like(update))

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				return client.DeleteUser(created)
			})
			assert.NoError(t, err)
		})

		t.Run("SetUserRoles", func(t *testing.T) {
			mockProvider.
				AddInteraction().
				Given("a user with uuid 819f6dbf-dd7a-47ff-b369-e3ed1d2578a0 exists").
				UponReceiving("a request to set user roles").
				WithRequest("PUT", S("/admin/users/819f6dbf-dd7a-47ff-b369-e3ed1d2578a0/roles")).
				WithHeader("Authorization", Like("Bearer 1234")).
				WithJSONBody(Like(setUserRoles)).
				WillRespondWith(200)

			err = mockProvider.ExecuteTest(t, func(config MockServerConfig) error {
				client := clientForPact(config)

				return client.SetUserRoles(created.UUID, setUserRoles)
			})
			assert.NoError(t, err)
		})
	})

}

func clientForPact(config MockServerConfig) *Client {
	baseURL, err := url.Parse(fmt.Sprintf("http://%s:%d", config.Host, config.Port))
	if err != nil {
		panic(fmt.Sprintf("unable to create client for pact test: %s", err))
	}

	return NewClient(nil, Config{
		AccessToken: "1234",
		BaseURL:     baseURL,
	})
}
