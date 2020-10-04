package main

import (
	"log"
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/sunfmin/auth1/gql"
	"github.com/sunfmin/auth1/gql/api"
	"github.com/sunfmin/auth1/gql/boot"
)

const defaultPort = "8080"


func SendMailTest(stuEmail string, subject string, body string) (err error) {
	fmt.Print("send success")
	return nil
}
func SendMsgTest(tel string, code string) (err error) {
	fmt.Print("send success")
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", handler.NewDefaultServer(
		gql.NewExecutableSchema(
			gql.Config{
				Resolvers: gql.NewResolver(
					boot.MustGetEntClient(),
					&api.BootConfig{
						AllowSignInWithVerifiedEmailAddress: true,
						AllowSignInWithVerifiedPhoneNumber: false,
						AllowSignInWithPreferredUsername: false,
						SendMailFunc:SendMailTest,
						SendMsgFunc: SendMsgTest,
					},
				),
			},
		)))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
