package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/pactflow/terraform/broker"
	"github.com/pactflow/terraform/client"
)

func team() *schema.Resource {
	return &schema.Resource{
		Create:   teamCreate,
		Update:   teamUpdate,
		Read:     teamRead,
		Delete:   teamDelete,
		Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Team",
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UUID of team",
			},
			"pacticipants": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"users": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func getTeamFromResourceData(d *schema.ResourceData) broker.Team {
	name := d.Get("name").(string)
	uuid := d.Id()
	team := broker.Team{
		UUID: uuid,
		Name: name,
	}

	pacticipants, ok := d.Get("pacticipants").([]interface{})
	log.Println("[DEBUG] resource_team.go pacticipants?", pacticipants, ok)
	if ok && len(pacticipants) > 0 {
		log.Println("[DEBUG] resource_team.go have pacticipants", len(pacticipants), pacticipants)
		items := make([]broker.Pacticipant, len(pacticipants))
		for i, p := range pacticipants {
			items[i] = broker.Pacticipant{
				Name: p.(string),
			}
		}
		team.Embedded.Pacticipants = items
	}

	// IF we do this, the resource creation tries to send the members along - but without all of the data
	// users, ok := d.Get("users").([]interface{})
	// log.Println("[DEBUG] resource_team.go users?", users, ok)
	// if ok && len(users) > 0 {
	// 	log.Println("[DEBUG] resource_team.go have users", len(users), users)
	// 	items := make([]broker.User, len(users))
	// 	for i, u := range users {
	// 		items[i] = broker.User{
	// 			UUID: u.(string),
	// 		}
	// 	}
	// 	team.Embedded.Members = items
	// }

	return team
}

// Removes any users from the team that shouldn't be there, and adds those that should
func assignTeamUsers(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	uuid := d.Id()

	log.Println("[DEBUG] assigning users to team", uuid)

	if d.HasChange("users") {
		old, new := d.GetChange("users")
		log.Println("[DEBUG] teamAssignmentCreate - change. old:", old, "new:", new)

		usersToAdd := interfaceToStringArray(new)
		log.Println("[DEBUG] teamAssignmentCreate - setting users:", usersToAdd)

		req := broker.TeamsAssignmentRequest{
			UUID:  uuid,
			Users: usersToAdd,
		}

		res, err := client.UpdateTeamAssignments(req)

		if err != nil {
			return err
		}

		return setTeamAssignmentState(d, res)
	}

	return nil
}

func teamCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	team := getTeamFromResourceData(d)

	log.Println("[DEBUG] creating team", team)

	created, err := client.CreateTeam(team)

	if err != nil {
		return fmt.Errorf("error creating team: %w", err)
	}

	team.UUID = created.UUID
	d.SetId(created.UUID)
	setTeamState(d, *created)

	err = assignTeamUsers(d, meta)
	if err != nil {
		d.Partial(true)
		log.Printf("\n\n[DEBUG] error assigning team users: %v \n\n", err)
		return fmt.Errorf("error assigning team users: %w", err)
	}

	return err
}

func teamUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	team := getTeamFromResourceData(d)

	log.Println("[DEBUG] updating team", team)

	updated, err := client.UpdateTeam(team)

	if err == nil {
		setTeamState(d, *updated)
	}

	err = assignTeamUsers(d, meta)
	if err != nil {
		d.Partial(true)
		return fmt.Errorf("error assigning team users: %w", err)
	}

	return err
}

func teamRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	uuid := d.Id()

	log.Println("[DEBUG] reading team", uuid)

	team, err := client.ReadTeam(uuid)

	log.Println("[DEBUG] have team for READ", team)

	if err == nil {
		d.SetId(team.UUID)
		setTeamState(d, *team)
	}

	return err
}

func teamDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*client.Client)
	uuid := d.Id()
	team := broker.Team{
		UUID: uuid,
	}

	log.Println("[DEBUG] deleting team", team)

	err := client.DeleteTeam(team)

	if err != nil {
		d.SetId("")
	}

	return err
}

func setTeamState(d *schema.ResourceData, team broker.Team) error {
	log.Printf("[DEBUG] setting team state: %+v \n", team)

	if err := d.Set("name", team.Name); err != nil {
		log.Println("[ERROR] error setting key 'name'", err)
		return err
	}
	if err := d.Set("uuid", team.UUID); err != nil {
		log.Println("[ERROR] error setting key 'uuid'", err)
		return err
	}

	if len(team.Embedded.Pacticipants) > 0 {
		pacticipants := make([]string, len(team.Embedded.Pacticipants))
		for i, p := range team.Embedded.Pacticipants {
			pacticipants[i] = p.Name
		}

		if err := d.Set("pacticipants", pacticipants); err != nil {
			log.Println("[ERROR] error setting key 'pacticipants'", err)
			return err
		}
	}

	if len(team.Embedded.Members) > 0 {
		members := make([]string, len(team.Embedded.Members))
		for i, m := range team.Embedded.Members {
			log.Println("[DEBUG] adding team member with UUID", m.UUID)
			members[i] = m.UUID
		}

		if err := d.Set("users", members); err != nil {
			log.Println("[ERROR] error setting key 'users'", err)
			return err
		}
	}

	return nil
}

func setTeamAssignmentState(d *schema.ResourceData, team *broker.TeamsAssignmentResponse) error {
	log.Printf("[DEBUG] setting team assignment state: %+v \n", team)

	if team != nil {
		if err := d.Set("users", extractUsersFromAPIResponse(team)); err != nil {
			log.Println("[ERROR] error setting key 'users'", err)
			return err
		}
	}

	return nil
}

// Use this to find the delta, and delete them from the team
func extractUsersFromState(d *schema.ResourceData) []string {
	usersRaw := d.Get("users").([]interface{})
	users := make([]string, len(usersRaw))
	for i, u := range usersRaw {
		users[i] = u.(string)
	}

	return users
}

// Use this to find the delta, and delete them from the team
func extractUsersFromAPIResponse(response *broker.TeamsAssignmentResponse) []string {
	users := make([]string, len(response.Embedded.Users))
	for i, u := range response.Embedded.Users {
		users[i] = u.UUID
	}

	return users
}
