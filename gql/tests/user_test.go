package tests

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sunfmin/auth1/ent"
	"github.com/sunfmin/auth1/gql/api"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func VerificationCodeTest() string {
	code := "111111"
	return code
}
func SendMailTest(stuEmail string, subject string, body string) (err error) {
	if stuEmail == "test_error" {
		err := fmt.Errorf("Verification code sending failed")
		return err
	}
	fmt.Printf("send success")
	return nil
}
func SendMsgTest(tel string, code string) (err error) {
	if tel == "test_error" {
		err := fmt.Errorf("Verification code sending failed")
		return err
	}
	fmt.Print("send success")
	return nil
}
func CreateAccessTokenTest(name string) (string, error) {
	token := TestAccessToken
	return token, nil
}
func CreateAccessToken(name string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Username": name,
		"exp":      time.Now().Add(time.Second * 3600).Unix(),
	})

	return token.SignedString([]byte("welcomelogin"))
}

func TimeSubTestPass(input string) (err error) {
	return
}
func TimeSubTestFail(input string) (err error) {
	err = fmt.Errorf("Captcha timeout")
	return err
}

var TestAccessToken, _ = CreateAccessToken("test")
var code_hash, _ = bcrypt.GenerateFromPassword([]byte("111111"), bcrypt.DefaultCost)
var password_hash, _ = bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

var NormalCfg = &api.BootConfig{
	AllowSignInWithVerifiedEmailAddress: true,
	AllowSignInWithVerifiedPhoneNumber:  false,
	AllowSignInWithPreferredUsername:    false,
	TimeSubFunc:                         TimeSubTestPass,
	SendMailFunc:                        SendMailTest,
	SendMsgFunc:                         SendMsgTest,
	CreateAccessTokenFunc:               CreateAccessTokenTest,
	CreateVerificationCodeFunc:          VerificationCodeTest,
}

var TimeoutCfg = &api.BootConfig{
	AllowSignInWithVerifiedEmailAddress: true,
	AllowSignInWithVerifiedPhoneNumber:  false,
	AllowSignInWithPreferredUsername:    false,
	TimeSubFunc:                         TimeSubTestFail,
	SendMailFunc:                        SendMailTest,
	SendMsgFunc:                         SendMsgTest,
	CreateAccessTokenFunc:               CreateAccessTokenTest,
	CreateVerificationCodeFunc:          VerificationCodeTest,
}

var userMutationCases = []GraphqlCase{
	{
		name:       "SignUp normal",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
			mutation ($input: SignUpInput!) {
				SignUp(input: $input) {
					CodeDeliveryDetails{
						AttributeName,
						DeliveryMedium,
						Destination
					  },
				   UserConfirmed,
	                  UserSub
				}
			}
			`,
		vars: []Var{
			{
				name: "input",
				val: api.SignUpInput{
					Username: "test",
					UserAttributes: []*api.AttributeType{
						{
							Name:  "email",
							Value: "test@test.com",
						},
						{
							Name:  "phone_number",
							Value: "test",
						},
					},
					Password: "test",
				},
			},
		},
		expected: &api.Data{
			SignUp: &api.User{
				CodeDeliveryDetails: &api.CodeDeliveryDetails{
					AttributeName:  "",
					DeliveryMedium: "",
					Destination:    "",
				},
			},
		},
	}, {
		name:       "SignUp verification code sending failed",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
			mutation ($input: SignUpInput!) {
				SignUp(input: $input) {
					CodeDeliveryDetails{
						AttributeName,
						DeliveryMedium,
						Destination
					  },
				   UserConfirmed,
	                  UserSub
				}
			}
			`,
		vars: []Var{
			{
				name: "input",
				val: api.SignUpInput{
					Username: "test_error",
					UserAttributes: []*api.AttributeType{
						{
							Name:  "email",
							Value: "test_error",
						},
						{
							Name:  "phone_number",
							Value: "test_error",
						},
					},
					Password: "test_error",
				},
			},
		},
		expectedError: "graphql: Verification code sending failed",
	}, {
		name: "ConfirmSignUp normal",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetConfirmationCodeHash(string(code_hash)).
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
			mutation ConfirmSignUp($input:ConfirmSignUpInput!){
			  ConfirmSignUp(input:$input){
						ConfirmStatus,
				}
			}
			`,
		vars: []Var{
			{
				name: "input",
				val: api.ConfirmSignUpInput{
					Username:         "test",
					ConfirmationCode: "111111",
				},
			},
		},
		expected: &api.Data{
			ConfirmSignUp: &api.ConfirmOutput{
				ConfirmStatus: true,
			},
		},
	}, {
		name:       "ConfirmSignUp account does not exist",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
			mutation ConfirmSignUp($input:ConfirmSignUpInput!){
			  ConfirmSignUp(input:$input){
						ConfirmStatus,
				}
			}
			`,
		vars: []Var{
			{
				name: "input",
				val: api.ConfirmSignUpInput{
					Username:         "test",
					ConfirmationCode: "111111",
				},
			},
		},
		expectedError: "graphql: Account does not exist",
	}, {
		name: "ConfirmSignUp verification code error",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test_error").
				SetPhoneNumber("test_error").
				SetEmail("test_error").
				SetID(uuid.New()).
				SetConfirmationCodeHash(string(code_hash)).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
			mutation ConfirmSignUp($input:ConfirmSignUpInput!){
			  ConfirmSignUp(input:$input){
						ConfirmStatus,
				}
			}
			`,
		vars: []Var{
			{
				name: "input",
				val: api.ConfirmSignUpInput{
					Username:         "test_error",
					ConfirmationCode: "000000",
				},
			},
		},
		expectedError: "graphql: Wrong verification code",
	}, {
		name: "ConfirmSignUp verification code timeout",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test_error").
				SetPhoneNumber("test_error").
				SetEmail("test_error").
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: TimeoutCfg,
		query: `
			mutation ConfirmSignUp($input:ConfirmSignUpInput!){
			  ConfirmSignUp(input:$input){
						ConfirmStatus,
				}
			}
			`,
		vars: []Var{
			{
				name: "input",
				val: api.ConfirmSignUpInput{
					Username:         "test_error",
					ConfirmationCode: "111111",
				},
			},
		},
		expectedError: "graphql: Captcha timeout",
	}, {
		name: "ChangePassword normal",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetPasswordHash(string(password_hash)).
				SetTokenState(1).
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ChangePassword($input:ChangePasswordInput!){
		  ChangePassword(input:$input){
			ConfirmStatus,
		  }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ChangePasswordInput{
					AccessToken:      TestAccessToken,
					PreviousPassword: "password",
					ProposedPassword: "new_password",
				},
			},
		},
		expected: &api.Data{
			ChangePassword: &api.ConfirmOutput{
				ConfirmStatus: true,
			},
		},
	}, {
		name: "ChangePassword token is invalid",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetPasswordHash(string(password_hash)).
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ChangePassword($input:ChangePasswordInput!){
		  ChangePassword(input:$input){
			ConfirmStatus,
		  }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ChangePasswordInput{
					AccessToken:      TestAccessToken,
					PreviousPassword: "password",
					ProposedPassword: "new_password",
				},
			},
		},
		expectedError: "graphql: Token is invalid",
	}, {
		name:       "ChangePassword accesstoken is nil",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
		mutation ChangePassword($input:ChangePasswordInput!){
		  ChangePassword(input:$input){
			ConfirmStatus,
		  }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ChangePasswordInput{
					AccessToken:      "",
					PreviousPassword: "password",
					ProposedPassword: "new_password",
				},
			},
		},
		expectedError: "graphql: AccessToken is nil",
	}, {
		name: "ChangePassword wrong previouspassword",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetTokenState(1).
				SetPasswordHash(string(password_hash)).
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ChangePassword($input:ChangePasswordInput!){
		  ChangePassword(input:$input){
			ConfirmStatus,
		  }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ChangePasswordInput{
					AccessToken:      TestAccessToken,
					PreviousPassword: "wrong_password",
					ProposedPassword: "new_password",
				},
			},
		},
		expectedError: "graphql: Wrong PreviousPassword",
	}, {
		name: "ChangePassword password no change",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetTokenState(1).
				SetPasswordHash(string(password_hash)).
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ChangePassword($input:ChangePasswordInput!){
		  ChangePassword(input:$input){
			ConfirmStatus,
		  }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ChangePasswordInput{
					AccessToken:      TestAccessToken,
					PreviousPassword: "password",
					ProposedPassword: "password",
				},
			},
		},
		expectedError: "graphql: The new password cannot be the same as the old password",
	}, {
		name: "ForgotPassword normal",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ($input:ForgotPasswordInput!){
		  ForgotPassword(input:$input){
			AttributeName,
			DeliveryMedium,
			Destination
		  }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ForgotPasswordInput{
					Username: "test",
				},
			},
		},
		expected: &api.Data{
			ForgotPassword: &api.CodeDeliveryDetails{},
		},
	}, {
		name:       "ForgotPassword account does not exist",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
		mutation ($input:ForgotPasswordInput!){
		  ForgotPassword(input:$input){
			AttributeName,
			DeliveryMedium,
			Destination
		  }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ForgotPasswordInput{
					Username: "test_err",
				},
			},
		},
		expectedError: "graphql: Account does not exist",
	}, {
		name: "ForgotPassword verification code sending failed",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test_error").
				SetPhoneNumber("test_error").
				SetEmail("test_error").
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ($input:ForgotPasswordInput!){
		  ForgotPassword(input:$input){
			AttributeName,
			DeliveryMedium,
			Destination
		  }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ForgotPasswordInput{
					Username: "test_error",
				},
			},
		},
		expectedError: "graphql: Verification code sending failed",
	}, {
		name: "ConfirmForgotPassword normal",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetConfirmationCodeHash(string(code_hash)).
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ($input:ConfirmForgotPasswordInput!){
		  ConfirmForgotPassword(input:$input){
				ConfirmStatus,
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ConfirmForgotPasswordInput{
					Username:         "test",
					ConfirmationCode: "111111",
					Password:         "test",
				},
			},
		},
		expected: &api.Data{
			ConfirmForgotPassword: &api.ConfirmOutput{
				ConfirmStatus: true,
			},
		},
	}, {
		name:       "ConfirmForgotPassword account does not exist",
		bootConfig: NormalCfg,
		query: `
		mutation ($input:ConfirmForgotPasswordInput!){
		  ConfirmForgotPassword(input:$input){
				ConfirmStatus,
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ConfirmForgotPasswordInput{
					Username:         "test",
					ConfirmationCode: "111111",
					Password:         "test",
				},
			},
		},
		expectedError: "graphql: Account does not exist",
	}, {
		name: "ConfirmForgotPassword verification code error",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test_error").
				SetPhoneNumber("test_error").
				SetEmail("test_error").
				SetConfirmationCodeHash(string(code_hash)).
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ($input:ConfirmForgotPasswordInput!){
		  ConfirmForgotPassword(input:$input){
				ConfirmStatus,
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ConfirmForgotPasswordInput{
					Username:         "test_error",
					ConfirmationCode: "000000",
					Password:         "test_error",
				},
			},
		},
		expectedError: "graphql: Wrong verification code",
	}, {
		name: "ConfirmForgotPassword verification code timeout",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test_error").
				SetPhoneNumber("test_error").
				SetEmail("test_error").
				SetConfirmationCodeHash(string(code_hash)).
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: TimeoutCfg,
		query: `
		mutation ($input:ConfirmForgotPasswordInput!){
		  ConfirmForgotPassword(input:$input){
				ConfirmStatus,
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ConfirmForgotPasswordInput{
					Username:         "test_error",
					ConfirmationCode: "111111",
					Password:         "test_error",
				},
			},
		},
		expectedError: "graphql: Captcha timeout",
	}, {
		name: "ResendConfirmationCode normal",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ($input:ResendConfirmationCodeInput!){
		  ResendConfirmationCode(input:$input){
				ConfirmStatus,
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ResendConfirmationCodeInput{
					Username: "test",
				},
			},
		},
		expected: &api.Data{
			ResendConfirmationCode: &api.ConfirmOutput{
				ConfirmStatus: true,
			},
		},
	}, {
		name: "ResendConfirmationCode verification code sending failed",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test_error").
				SetPhoneNumber("test_error").
				SetEmail("test_error").
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ($input:ResendConfirmationCodeInput!){
		  ResendConfirmationCode(input:$input){
				ConfirmStatus,
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.ResendConfirmationCodeInput{
					Username: "test_error",
				},
			},
		},
		expectedError: "graphql: Verification code sending failed",
	}, {
		name: "InitiateAuth normal",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetID(uuid.New()).
				SetPasswordHash(string(password_hash)).
				SetActiveState(1).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation InitiateAuth($input:InitiateAuthInput!){
		  InitiateAuth(input:$input){
			AccessToken,
			ExpiresIn,
			IdToken,
			RefreshToken,
			TokenType
          }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.InitiateAuthInput{
					AuthFlow: "USER_PASSWORD_AUTH",
					AuthParameters: &api.AuthParameters{
						Username: "test",
						Password: "password",
					},
				},
			},
		},
		expected: &api.Data{
			InitiateAuth: &api.AuthenticationResult{
				ExpiresIn: 3600,
			},
		},
	}, {
		name: "InitiateAuth user is not activated",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test_error").
				SetPhoneNumber("test_error").
				SetEmail("test_error").
				SetID(uuid.New()).
				SetActiveState(0).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation InitiateAuth($input:InitiateAuthInput!){
		  InitiateAuth(input:$input){
			AccessToken,
			ExpiresIn,
			IdToken,
			RefreshToken,
			TokenType
          }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.InitiateAuthInput{
					AuthFlow: "USER_PASSWORD_AUTH",
					AuthParameters: &api.AuthParameters{
						Username: "test_error",
						Password: "password",
					},
				},
			},
		},
		expectedError: "graphql: The user is not activated",
	}, {
		name:       "InitiateAuth AuthFlow is nil",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
		mutation InitiateAuth($input:InitiateAuthInput!){
		  InitiateAuth(input:$input){
			AccessToken,
			ExpiresIn,
			IdToken,
			RefreshToken,
			TokenType
          }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.InitiateAuthInput{
					AuthFlow: "",
					AuthParameters: &api.AuthParameters{
						Username: "test_error",
						Password: "password",
					},
				},
			},
		},
		expectedError: "graphql: AuthFlow is nil",
	}, {
		name:       "InitiateAuth Unknown AuthFlow",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
		mutation InitiateAuth($input:InitiateAuthInput!){
		  InitiateAuth(input:$input){
			AccessToken,
			ExpiresIn,
			IdToken,
			RefreshToken,
			TokenType
          }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.InitiateAuthInput{
					AuthFlow: "Unknown_AuthFlow",
					AuthParameters: &api.AuthParameters{
						Username: "test_error",
						Password: "password",
					},
				},
			},
		},
		expectedError: "graphql: Unknown AuthFlow",
	}, {
		name:       "InitiateAuth account does not exist",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
		mutation InitiateAuth($input:InitiateAuthInput!){
		  InitiateAuth(input:$input){
			AccessToken,
			ExpiresIn,
			IdToken,
			RefreshToken,
			TokenType
          }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.InitiateAuthInput{
					AuthFlow: "USER_PASSWORD_AUTH",
					AuthParameters: &api.AuthParameters{
						Username: "test_error",
						Password: "password",
					},
				},
			},
		},
		expectedError: "graphql: Account does not exist",
	}, {
		name: "InitiateAuth Wrong Password",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test_error").
				SetPhoneNumber("test_error").
				SetEmail("test_error").
				SetID(uuid.New()).
				SetPasswordHash(string(password_hash)).
				SetActiveState(1).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation InitiateAuth($input:InitiateAuthInput!){
		  InitiateAuth(input:$input){
			AccessToken,
			ExpiresIn,
			IdToken,
			RefreshToken,
			TokenType
          }
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.InitiateAuthInput{
					AuthFlow: "USER_PASSWORD_AUTH",
					AuthParameters: &api.AuthParameters{
						Username: "test_error",
						Password: "wrong_password",
					},
				},
			},
		},
		expectedError: "graphql: Wrong password",
	}, {
		name: "GlobalSignOut normal",
		fixture: func(ctx context.Context, client *ent.Client) {
			client.User.Create().
				SetUsername("test").
				SetPhoneNumber("test").
				SetEmail("test").
				SetID(uuid.New()).
				SaveX(ctx)
		},
		bootConfig: NormalCfg,
		query: `
		mutation ($input:GlobalSignOutInput!){
		  GlobalSignOut(input:$input){
				ConfirmStatus,
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.GlobalSignOutInput{
					AccessToken: TestAccessToken,
				},
			},
		},
		expected: &api.Data{
			GlobalSignOut: &api.ConfirmOutput{
				ConfirmStatus: true,
			},
		},
	}, {
		name:       "GlobalSignOut accessToken is nil",
		fixture:    nil,
		bootConfig: NormalCfg,
		query: `
		mutation ($input:GlobalSignOutInput!){
		  GlobalSignOut(input:$input){
				ConfirmStatus,
			}
		}
		`,
		vars: []Var{
			{
				name: "input",
				val: api.GlobalSignOutInput{
					AccessToken: "",
				},
			},
		},
		expectedError: "graphql: AccessToken is nil",
	},
}
