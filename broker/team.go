package broker

// Team represents a user login to Pactflow
type Team struct {
	UUID            string            `json:"uuid,omitempty"`
	Name            string            `json:"name,omitempty"`
	NumberOfMembers int               `json:"numberOfMembers,omitempty"`
	Roles           []RoleV1          `json:"roles,omitempty"`
	Embedded        TeamEmbeddedItems `json:"_embedded,omitempty"`
}

type TeamEmbeddedItems struct {
	Pacticipants []Pacticipant `json:"pacticipants,omitempty"`
	Members      []User        `json:"members,omitempty"`
}

type TeamsResponse struct {
	Teams []Team `json:"teams"`
	HalDoc
}

type TeamsAssignmentRequest struct {
	// Team UUID
	UUID string `json:"-"`

	// List of user UUIDs
	Users []string `json:"users,omitempty"`
}

type TeamsAssignmentResponse struct {
	Embedded EmbeddedUsers `json:"_embedded,omitempty"`
}

type EmbeddedUsers = struct {
	Users []User `json:"users,omitempty"`
}

// Create POST /admin/teams
// {"name":"foobar","_embedded":{"pacticipants":[{"name":"Aaaa_Order_API"}]}}

// GET /admin/teams
// {
//   "teams": [
//     {
//       "createdAt": "2020-10-21T02:35:45+00:00",
//       "uuid": "d0511642-ec92-461a-9aa3-fb8bb8fd6ff4",
//       "name": "foobar",
//       "numberOfMembers": 0,
//       "_links": {
//         "self": {
//           "title": "Team",
//           "href": "https://testdemo.pactflow.io/admin/teams/d0511642-ec92-461a-9aa3-fb8bb8fd6ff4"
//         }
//       }
//     },
//     {
//       "createdAt": "2020-10-20T08:20:02+00:00",
//       "uuid": "6d746ad5-919f-49e3-84c0-648cafc5d912",
//       "name": "Pactflow Demos",
//       "numberOfMembers": 1,
//       "_links": {
//         "self": {
//           "title": "Team",
//           "href": "https://testdemo.pactflow.io/admin/teams/6d746ad5-919f-49e3-84c0-648cafc5d912"
//         }
//       }
//     }
//   ],
//   "_links": {
//     "self": {
//       "title": "Teams",
//       "href": "https://testdemo.pactflow.io/admin/teams"
//     }
//   }
// }

// GET /admin/teams/d0511642-ec92-461a-9aa3-fb8bb8fd6ff4
// {
//   "createdAt": "2020-10-20T08:20:02+00:00",
//   "uuid": "6d746ad5-919f-49e3-84c0-648cafc5d912",
//   "name": "Pactflow Demos",
//   "numberOfMembers": 1,
//   "_embedded": {
//     "members": [
//       {
//         "updatedAt": "2020-10-20T06:53:11+00:00",
//         "createdAt": "2020-10-20T06:53:11+00:00",
//         "uuid": "4c260344-b170-41eb-b01e-c0ff10c72f25",
//         "identityProviderId": "d050ce6c-9fa7-4c71-80d3-95af9b3b5cb9",
//         "name": "Matt Fellows",
//         "active": true,
//         "lastLogin": "2020-10-20T09:05:07.623+00:00",
//         "email": "matt.fellows@onegeek.com.au",
//         "type": 0,
//         "typeDescription": "User",
//         "_embedded": {
//           "roles": [
//             {
//               "uuid": "e9282e22-416b-11ea-a16e-57ee1bb61d18",
//               "name": "Test Maintainer",
//               "permissions": [
//                 {
//                   "name": "Read system accounts",
//                   "scope": "system_account:read:*"
//                 },
//                 {
//                   "name": "Read roles",
//                   "scope": "role:read:*"
//                 },
//                 {
//                   "name": "Read users",
//                   "scope": "user:read:*"
//                 },
//                 {
//                   "name": "Read teams",
//                   "scope": "team:read:*"
//                 },
//                 {
//                   "name": "Manage API tokens",
//                   "scope": "token:manage:own"
//                 },
//                 {
//                   "name": "Manage contract data",
//                   "scope": "contract_data:manage:*"
//                 },
//                 {
//                   "name": "Read contract data",
//                   "scope": "contract_data:read:*"
//                 },
//                 {
//                   "name": "Manage webhooks",
//                   "scope": "webhook:manage:*"
//                 },
//                 {
//                   "name": "Manage secrets",
//                   "scope": "secret:manage:*"
//                 },
//                 {
//                   "name": "Bulk delete contract data",
//                   "scope": "contract_data:bulk_delete:*"
//                 }
//               ],
//               "_links": {
//                 "self": {
//                   "title": "User Role",
//                   "name": "Test Maintainer",
//                   "href": "https://testdemo.pactflow.io/admin/users/4c260344-b170-41eb-b01e-c0ff10c72f25/roles/e9282e22-416b-11ea-a16e-57ee1bb61d18"
//                 }
//               }
//             },
//             {
//               "uuid": "cf75d7c2-416b-11ea-af5e-53c3b1a4efd8",
//               "name": "Administrator",
//               "permissions": [
//                 {
//                   "name": "Manage users",
//                   "scope": "user:manage:*"
//                 },
//                 {
//                   "name": "Manage teams",
//                   "scope": "team:manage:*"
//                 },
//                 {
//                   "name": "Invite users",
//                   "scope": "user:invite"
//                 },
//                 {
//                   "name": "Manage system accounts",
//                   "scope": "system_account:manage:*"
//                 },
//                 {
//                   "name": "Manage roles",
//                   "scope": "role:manage:*"
//                 },
//                 {
//                   "name": "Manage API tokens",
//                   "scope": "token:manage:own"
//                 },
//                 {
//                   "name": "Manage contract data",
//                   "scope": "contract_data:manage:*"
//                 },
//                 {
//                   "name": "Bulk delete contract data",
//                   "scope": "contract_data:bulk_delete:*"
//                 },
//                 {
//                   "name": "Manage webhooks",
//                   "scope": "webhook:manage:*"
//                 },
//                 {
//                   "name": "Manage secrets",
//                   "scope": "secret:manage:*"
//                 }
//               ],
//               "_links": {
//                 "self": {
//                   "title": "User Role",
//                   "name": "Administrator",
//                   "href": "https://testdemo.pactflow.io/admin/users/4c260344-b170-41eb-b01e-c0ff10c72f25/roles/cf75d7c2-416b-11ea-af5e-53c3b1a4efd8"
//                 }
//               }
//             }
//           ],
//           "teams": [
//             {
//               "name": "Pactflow Demos",
//               "uuid": "6d746ad5-919f-49e3-84c0-648cafc5d912",
//               "_links": {
//                 "self": {
//                   "title": "Team",
//                   "name": "Pactflow Demos",
//                   "href": "https://testdemo.pactflow.io/teams/6d746ad5-919f-49e3-84c0-648cafc5d912"
//                 },
//                 "pf:membership": {
//                   "title": "Team membership",
//                   "href": "https://testdemo.pactflow.io/teams/6d746ad5-919f-49e3-84c0-648cafc5d912/users/4c260344-b170-41eb-b01e-c0ff10c72f25"
//                 },
//                 "pf:integrations": {
//                   "title": "Team integrations",
//                   "name": "Pactflow Demos",
//                   "href": "https://testdemo.pactflow.io/integrations/team/6d746ad5-919f-49e3-84c0-648cafc5d912"
//                 }
//               }
//             }
//           ]
//         },
//         "_links": {
//           "self": {
//             "title": "User",
//             "href": "https://testdemo.pactflow.io/admin/teams/6d746ad5-919f-49e3-84c0-648cafc5d912/4c260344-b170-41eb-b01e-c0ff10c72f25"
//           }
//         },
//         "roles": [
//           {
//             "updatedAt": "2020-04-02T22:42:28+00:00",
//             "createdAt": "2020-04-02T22:42:28+00:00",
//             "uuid": "cf75d7c2-416b-11ea-af5e-53c3b1a4efd8",
//             "name": "Administrator",
//             "permissions": [
//               {
//                 "name": "Manage users",
//                 "scope": "user:manage:*"
//               },
//               {
//                 "name": "Manage teams",
//                 "scope": "team:manage:*"
//               },
//               {
//                 "name": "Invite users",
//                 "scope": "user:invite"
//               },
//               {
//                 "name": "Manage system accounts",
//                 "scope": "system_account:manage:*"
//               },
//               {
//                 "name": "Manage roles",
//                 "scope": "role:manage:*"
//               },
//               {
//                 "name": "Manage API tokens",
//                 "scope": "token:manage:own"
//               },
//               {
//                 "name": "Manage contract data",
//                 "scope": "contract_data:manage:*"
//               },
//               {
//                 "name": "Bulk delete contract data",
//                 "scope": "contract_data:bulk_delete:*"
//               },
//               {
//                 "name": "Manage webhooks",
//                 "scope": "webhook:manage:*"
//               },
//               {
//                 "name": "Manage secrets",
//                 "scope": "secret:manage:*"
//               }
//             ]
//           },
//           {
//             "updatedAt": "2020-04-02T22:42:28+00:00",
//             "createdAt": "2020-04-02T22:42:28+00:00",
//             "uuid": "e9282e22-416b-11ea-a16e-57ee1bb61d18",
//             "name": "Test Maintainer",
//             "permissions": [
//               {
//                 "name": "Read system accounts",
//                 "scope": "system_account:read:*"
//               },
//               {
//                 "name": "Read roles",
//                 "scope": "role:read:*"
//               },
//               {
//                 "name": "Read users",
//                 "scope": "user:read:*"
//               },
//               {
//                 "name": "Read teams",
//                 "scope": "team:read:*"
//               },
//               {
//                 "name": "Manage API tokens",
//                 "scope": "token:manage:own"
//               },
//               {
//                 "name": "Manage contract data",
//                 "scope": "contract_data:manage:*"
//               },
//               {
//                 "name": "Read contract data",
//                 "scope": "contract_data:read:*"
//               },
//               {
//                 "name": "Manage webhooks",
//                 "scope": "webhook:manage:*"
//               },
//               {
//                 "name": "Manage secrets",
//                 "scope": "secret:manage:*"
//               },
//               {
//                 "name": "Bulk delete contract data",
//                 "scope": "contract_data:bulk_delete:*"
//               }
//             ]
//           }
//         ]
//       }
//     ]
//   },
//   "_links": {
//     "self": {
//       "title": "Team",
//       "href": "https://testdemo.pactflow.io/admin/teams/6d746ad5-919f-49e3-84c0-648cafc5d912"
//     }
//   }
// }
