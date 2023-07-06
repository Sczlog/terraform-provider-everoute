package everoute

import (
	"context"
	"fmt"

	"github.com/Sczlog/dgql"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/smartxworks/cloudtower-go-sdk/v2/client"
)

type Client struct {
	server  string
	token   string
	DgqlApi *dgql.GraphqlClient
	Api     *apiclient.Cloudtower
}

func NewClient(username string, password string, server string) (*Client, error) {
	client, err := dgql.NewClient(fmt.Sprintf("http://%s/api/", server))
	if err != nil {
		return nil, err
	}
	loginResp, _, err := client.Mutation(context.Background(), "login", map[string]interface{}{
		"data": map[string]interface{}{
			"username": username,
			"password": password,
			"source":   "LOCAL",
		},
	}, nil)
	if err != nil {
		return nil, err
	}
	token := loginResp.Get("login.token").String()
	client.DefaultHeaders["Authorization"] = token

	transport := httptransport.New(server, "/v2/api", []string{"http"})
	transport.DefaultAuthentication = httptransport.APIKeyAuth("Authorization", "header", token)
	apiclient := apiclient.New(transport, strfmt.Default)
	return &Client{
		server:  server,
		token:   token,
		DgqlApi: client,
		Api:     apiclient,
	}, nil
}
