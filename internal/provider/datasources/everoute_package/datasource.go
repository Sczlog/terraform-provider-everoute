package everoute_package

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	erp "github.com/smartxworks/cloudtower-go-sdk/v2/client/everoute_package"
	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
	"github.com/smartxworks/terraform-provider-everoute/internal/everoute"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DataSource{}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource defines the data source implementation.
type DataSource struct {
	client *everoute.Client
}

// model describes the data source data model.
type model struct {
	Id           types.String                   `tfsdk:"id"`
	Version      types.String                   `tfsdk:"version"`
	Architecture types.String                   `tfsdk:"architecture"`
	Packages     []EveroutePackageResponseModel `tfsdk:"packages"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Get everoute package by version and architecture",
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				MarkdownDescription: "everoute package version",
				Required:            true,
			},
			"architecture": schema.StringAttribute{
				MarkdownDescription: "everoute package architecture, valid values are X86_64 and AARCH64",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("X86_64", "AARCH64"),
				},
			},
			"packages": everoutePackageResponseSchema(),
			"id": schema.StringAttribute{
				MarkdownDescription: "used to mark query time, to keep data source update every time apply",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
		},
	}
}

func (d *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*everoute.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var newState model

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &newState)...)

	if resp.Diagnostics.HasError() {
		return
	}
	var arch models.Architecture
	switch strings.ToUpper(newState.Architecture.ValueString()) {
	case "X86_64":
		arch = models.ArchitectureX8664
	case "AARCH64":
		arch = models.ArchitectureAARCH64
	}
	gerpp := erp.NewGetEveroutePackagesParams()
	gerpp.RequestBody = &models.GetEveroutePackagesRequestBody{
		Where: &models.EveroutePackageWhereInput{
			Arch:    arch.Pointer(),
			Version: newState.Version.ValueStringPointer(),
		},
	}
	result, err := d.client.Api.EveroutePackage.GetEveroutePackages(gerpp)

	if err != nil {
		resp.Diagnostics.AddError("Failed to query everoute package", err.Error())
		return
	}

	if len(result.Payload) == 0 {
		resp.Diagnostics.AddError("No everoute package found", fmt.Sprintf("version: %s, architecture: %s", newState.Version.ValueString(), newState.Architecture.ValueString()))
		return
	}
	resp.Diagnostics.Append(modelToState(ctx, result, &newState)...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func modelToState(ctx context.Context, res *erp.GetEveroutePackagesOK, state *model) diag.Diagnostics {
	var diagsnostics diag.Diagnostics
	var diags diag.Diagnostics

	state.Packages, diags = flattenEveroutePackageResponse(ctx, res.Payload)
	diagsnostics.Append(diags...)

	state.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	return diagsnostics
}
