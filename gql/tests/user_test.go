package tests

import (
	"github.com/sunfmin/auth1/gql/api"
)

var userMutationCases = []GraphqlCase{
	{
		name:    "SignUp normal",
		fixture: nil,
		query: `
		mutation ($input: SignUpInput!) {
			signUp(input: $input) {
				id
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.SignUpInput{
					Password: "hello",
				},
			},
		},
		expected: &api.Data{
			SignUp: &api.User{},
		},
	},
}
