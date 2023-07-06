// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/smartxworks/terraform-provider-everoute/internal/everoute"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &EverouteProvider{}

// EverouteProvider defines the provider implementation.
type EverouteProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Model describes the provider data model.
type Model struct {
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	CloudtowerServer types.String `tfsdk:"cloudtower_server"`
}

func (p *EverouteProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "everoute"
	resp.Version = p.version
}

func (p *EverouteProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for tower authentication, if not configured, use env CLOUDTOWER_USER.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for tower authentication, if not configured, use env CLOUDTOWER_PASSWORD.",
				Optional:            true,
			},
			"cloudtower_server": schema.StringAttribute{
				MarkdownDescription: "Cloudtower server url, if not configured, use env CLOUDTOWER_SERVER.",
				Optional:            true,
			},
		},
	}
}

func (p *EverouteProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data Model

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	var missingFields []string
	var user string
	var password string
	var server string
	if data.Username.IsNull() || data.Username.IsUnknown() {
		user = os.Getenv("CLOUDTOWER_USER")
		if user == "" {
			missingFields = append(missingFields, "username")
		}
	} else {
		user = data.Username.ValueString()
	}
	if data.Password.IsNull() || data.Username.IsUnknown() {
		password = os.Getenv("CLOUDTOWER_PASSWORD")
		if password == "" {
			missingFields = append(missingFields, "password")
		}
	} else {
		password = data.Password.ValueString()
	}
	if data.CloudtowerServer.IsNull() || data.CloudtowerServer.IsUnknown() {
		server = os.Getenv("CLOUDTOWER_SERVER")
		if server == "" {
			missingFields = append(missingFields, "cloudtower_server")
		}
	} else {
		server = data.CloudtowerServer.ValueString()
	}
	if len(missingFields) > 0 {
		resp.Diagnostics.AddError(
			"missing required configuration fields",
			fmt.Sprintf(
				"missing required configuration fields: %s",
				strings.Join(missingFields, ", "),
			),
		)
		return
	}

	client, err := everoute.NewClient(user, password, server)

	if err != nil {
		resp.Diagnostics.AddError(
			"failed to create everoute client",
			fmt.Sprintf(
				"failed to create everoute client: %s",
				err.Error(),
			),
		)
		return
	}

	// Example client configuration for data sources and resources
	resp.DataSourceData = client
	resp.ResourceData = client
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &EverouteProvider{
			version: version,
		}
	}
}
