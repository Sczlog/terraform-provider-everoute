package everoute_service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tidwall/gjson"

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

// EverouteServiceModel describes the data source data EverouteServiceModel.
type EverouteServiceModel struct {
	Id       types.String                   `tfsdk:"id"`
	Name     types.String                   `tfsdk:"name"`
	Services []EverouteServiceResponseModel `tfsdk:"services"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Get everoute cluster by name",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "query everoute cluster by name",
				Required:            true,
			},
			"services": everouteServiceResponseSchema(),
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
	var newState EverouteServiceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &newState)...)

	if resp.Diagnostics.HasError() {
		return
	}

	result, _, err := d.client.DgqlApi.Raw(ctx, getEverouteServiceDocument, "everouteClusters", map[string]interface{}{
		"where": map[string]interface{}{
			"name": newState.Name.ValueString(),
		},
	}, nil)

	if err != nil {
		resp.Diagnostics.AddError("Failed to query everoute services", err.Error())
		return
	}

	if len(result.Get("everouteClusters").Array()) == 0 {
		resp.Diagnostics.AddError("No everoute service found", fmt.Sprintf("name %s", newState.Name.ValueString()))
		return
	}
	resp.Diagnostics.Append(modelToState(ctx, result, &newState)...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func modelToState(ctx context.Context, input *gjson.Result, state *EverouteServiceModel) diag.Diagnostics {
	var diagsnostics diag.Diagnostics

	clusters := flattenEverouteServiceResponse(ctx, input)
	state.Services = clusters

	state.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	return diagsnostics
}
