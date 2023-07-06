package global_security_policy

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smartxworks/terraform-provider-everoute/internal/everoute"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &DataSource{}
var _ datasource.DataSourceWithConfigValidators = &DataSource{}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource defines the data source implementation.
type DataSource struct {
	client *everoute.Client
}

// GlobalSecurityPolicyDataSourceModel describes the data source data model.
type GlobalSecurityPolicyDataSourceModel struct {
	Id                   types.String                       `tfsdk:"id"`
	ServiceName          types.String                       `tfsdk:"service_name"`
	ServiceId            types.String                       `tfsdk:"service_id"`
	GlobalSecurityPolicy *GlobalSecurityPolicyResponseModel `tfsdk:"global_security_policy"`
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_security_policy"
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "everoute service's global security policy configuration",

		Attributes: map[string]schema.Attribute{
			"service_name": schema.StringAttribute{
				MarkdownDescription: "query global security policy by service name, must provided if service_id not provided",
				Optional:            true,
			},
			"service_id": schema.StringAttribute{
				MarkdownDescription: "query global security policy by service id, must provided if service_name not provided",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Example identifier",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
			"global_security_policy": globalSecurityPolicyResponseSchema(),
		},
	}
}

func (d DataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot("service_name"),
			path.MatchRoot("service_id"),
		),
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
			fmt.Sprintf("Expected *everoute.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GlobalSecurityPolicyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	var whereInput = make(map[string]interface{})
	if !state.ServiceId.IsNull() {
		whereInput["id"] = state.ServiceId.ValueString()
	} else {
		whereInput["name"] = state.ServiceName.ValueString()
	}

	gqlResp, _, err := d.client.DgqlApi.Raw(ctx, getWhiteListDocument, "everouteClusters", map[string]interface{}{
		"where": whereInput,
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError("read service global whitelist failed", err.Error())
		return
	}
	jgsp := gqlResp.Get("everouteClusters.0.global_whitelist")
	gsp := readGlobalSecurityPolicy(ctx, &jgsp)
	state.GlobalSecurityPolicy = &gsp

	if state.ServiceName.IsNull() {
		state.ServiceName = types.StringValue(gqlResp.Get("everouteClusters.0.name").String())
	}
	if state.ServiceId.IsNull() {
		state.ServiceId = types.StringValue(gqlResp.Get("everouteClusters.0.id").String())
	}

	state.Id = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
