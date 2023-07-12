package global_security_policy

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smartxworks/cloudtower-go-sdk/v2/utils"
	"github.com/smartxworks/terraform-provider-everoute/internal/everoute"
	"github.com/tidwall/gjson"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
var _ resource.ResourceWithValidateConfig = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *everoute.Client
}

// GlobalSecurityPolicyResourceModel describes the resource data model.
type GlobalSecurityPolicyResourceModel struct {
	Id            types.String             `tfsdk:"id"`
	ServiceId     types.String             `tfsdk:"service_id"`
	Enable        types.Bool               `tfsdk:"enable"`
	DefaultAction types.String             `tfsdk:"default_action"`
	Ingress       []NetworkPolicyRuleModel `tfsdk:"ingress"`
	Egress        []NetworkPolicyRuleModel `tfsdk:"egress"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_security_policy"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "everoute service global security policy configuration",

		Attributes: map[string]schema.Attribute{
			"service_id": schema.StringAttribute{
				MarkdownDescription: "service id global security policy listbelongs to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enable": schema.BoolAttribute{
				MarkdownDescription: "if global security policy is enabled",
				Required:            true,
			},
			"default_action": schema.StringAttribute{
				MarkdownDescription: "global security policy's default action, valid value: ALLOW, DROP",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("ALLOW", "DROP"),
				},
			},
			"ingress": schema.ListNestedAttribute{
				MarkdownDescription: "global security policy's ingress configuration",
				Required:            true,
				NestedObject:        networkPolicyRuleSchema(),
			},
			"egress": schema.ListNestedAttribute{
				MarkdownDescription: "global security policy's egress configuration",
				Required:            true,
				NestedObject:        networkPolicyRuleSchema(),
			},
			"id": schema.StringAttribute{
				Computed:            true,
				Optional:            false,
				Required:            false,
				MarkdownDescription: "identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*everoute.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *everoute.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *GlobalSecurityPolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// precheck everoute service existed
	serviceId := data.ServiceId.ValueString()
	result, _, err := r.client.DgqlApi.Raw(ctx, getGlobalWhitelistDocument, "getEverouteClusters", map[string]interface{}{
		"where": map[string]interface{}{
			"id": serviceId,
		},
	}, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create global security policy",
			fmt.Sprintf("Failed to get everoute service: %s", err),
		)
		return
	}

	jService := result.Get("everouteClusters.0")
	if jService.Type == gjson.Null {
		resp.Diagnostics.AddError(
			"Failed to create global security policy",
			fmt.Sprintf("Everoute service %s not found", serviceId),
		)
		return
	}

	// create global whitelist
	updateInput, diags := buildUpdateInput(data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, headers, err := r.client.DgqlApi.Raw(ctx, updateGlobalWhiteListDocument, "updateEverouteClusterGlobalAction", map[string]interface{}{
		"where": map[string]interface{}{
			"id": data.ServiceId.ValueString(),
		},
		"data": updateInput,
	}, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create global security policy",
			fmt.Sprintf("Failed to create everoute service global security policy: %s", err),
		)
		return
	}

	taskId := headers.Get("X-Task-Id")
	err = utils.WaitTask(ctx, r.client.Api, &taskId, 5*time.Second)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create global security policy",
			fmt.Sprintf("Failed to create everoute service global security policy: %s", err),
		)
		return
	}
	result, _, err = r.client.DgqlApi.Raw(ctx, getGlobalWhitelistDocument, "", map[string]interface{}{}, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get everoute service global security policy",
			fmt.Sprintf("Failed to get everoute service global security policy: %s", err),
		)
		return
	}

	jService = result.Get("everouteClusters.0")
	readGqlResultToState(&jService, data)
	if resp.Diagnostics.HasError() {
		return
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *GlobalSecurityPolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	result, _, err := r.client.DgqlApi.Raw(ctx, getGlobalWhitelistDocument, "", map[string]interface{}{}, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get everoute service global security policy",
			fmt.Sprintf("Failed to get everoute service global security policy: %s", err),
		)
		return
	}

	jService := result.Get("everouteClusters.0")
	readGqlResultToState(&jService, data)
	if resp.Diagnostics.HasError() {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *GlobalSecurityPolicyResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateInput, diags := buildUpdateInput(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update global whitelist
	_, headers, err := r.client.DgqlApi.Raw(ctx, updateGlobalWhiteListDocument, "updateEverouteClusterGlobalAction", map[string]interface{}{
		"where": map[string]interface{}{
			"id": plan.ServiceId.ValueString(),
		},
		"data": updateInput,
	}, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update global security policy",
			fmt.Sprintf("Failed to update everoute service global security policy: %s", err),
		)
		return
	}

	taskId := headers.Get("X-Task-Id")
	err = utils.WaitTask(ctx, r.client.Api, &taskId, 5*time.Second)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update global security policy",
			fmt.Sprintf("Failed to update everoute service global security policy: %s", err),
		)
		return
	}

	result, _, err := r.client.DgqlApi.Raw(ctx, getGlobalWhitelistDocument, "", map[string]interface{}{}, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get everoute service global security policy",
			fmt.Sprintf("Failed to get everoute service global security policy: %s", err),
		)
		return
	}

	jService := result.Get("everouteClusters.0")
	readGqlResultToState(&jService, plan)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *GlobalSecurityPolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	_, headers, err := r.client.DgqlApi.Raw(ctx, updateGlobalWhiteListDocument, "updateEverouteClusterGlobalAction", map[string]interface{}{
		"where": map[string]interface{}{
			"id": data.ServiceId.ValueString(),
		},
		"data": map[string]interface{}{
			"global_default_action": "ALLOW",
			"global_whitelist": map[string]interface{}{
				"enable":  false,
				"egress":  []interface{}{},
				"ingress": []interface{}{},
			},
		},
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete everoute service global security policy",
			fmt.Sprintf("Failed to delete everoute service global security policy: %s", err),
		)
		return
	}
	taskId := headers.Get("X-Task-Id")
	err = utils.WaitTask(ctx, r.client.Api, &taskId, 5*time.Second)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete everoute service global security policy",
			fmt.Sprintf("Failed to delete everoute service global security policy: %s", err),
		)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data *GlobalSecurityPolicyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	enabled := data.Enable.ValueBool()
	if !enabled {
		if len(data.Egress) != 0 {
			resp.Diagnostics.AddError(
				"Invalid egress",
				"egress is not allowed when global security policy is disabled",
			)
		}
		if len(data.Ingress) != 0 {
			resp.Diagnostics.AddError(
				"Invalid ingress",
				"ingress is not allowed when global security policy is disabled",
			)
		}
	} else if len(data.Egress)+len(data.Ingress) == 0 {
		resp.Diagnostics.AddError(
			"Invalid egress and ingress",
			"egress or ingress cannot be both empty when global security policy is enabled",
		)
	}
}

func readGqlResultToState(input *gjson.Result, state *GlobalSecurityPolicyResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	state.Id = types.StringValue(input.Get("id").String())
	state.ServiceId = state.Id
	state.Enable = types.BoolValue(input.Get("global_whitelist.enable").Bool())
	state.DefaultAction = types.StringValue(input.Get("global_default_action").String())
	jingress := input.Get("global_whitelist.ingress")
	jegress := input.Get("global_whitelist.egress")
	state.Ingress = readGqlResultToNetworkPolicyRule(&jingress)
	state.Egress = readGqlResultToNetworkPolicyRule(&jegress)
	return diags
}

func buildUpdateInput(state *GlobalSecurityPolicyResourceModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	ingress := make([]map[string]interface{}, 0)
	egress := make([]map[string]interface{}, 0)
	enabled := state.Enable.ValueBool()
	if enabled {
		for _, rule := range state.Ingress {
			ports := make([]map[string]interface{}, 0)
			if rule.TCPEnabled.ValueBool() {
				tcpport := make(map[string]interface{})
				tcpport["protocol"] = "TCP"
				tcpport["port"] = rule.TCPPorts.ValueString()
				ports = append(ports, tcpport)
			}
			if rule.UDPEnabled.ValueBool() {
				udpport := make(map[string]interface{})
				udpport["protocol"] = "UDP"
				udpport["port"] = rule.UDPPorts.ValueString()
				ports = append(ports, udpport)
			}
			if rule.ICMPEnabled.ValueBool() {
				icmpport := make(map[string]interface{})
				icmpport["protocol"] = "ICMP"
				ports = append(ports, icmpport)
			}
			eips := make([]string, 0)
			diags = rule.ExceptIPBlock.ElementsAs(context.Background(), &eips, false)
			ingress = append(ingress, map[string]interface{}{
				"type":            "IP_BLOCK",
				"ip_block":        rule.IPBlock.ValueString(),
				"ports":           ports,
				"except_ip_block": eips,
			})
		}
		for _, rule := range state.Egress {
			ports := make([]map[string]interface{}, 0)
			if rule.TCPEnabled.ValueBool() {
				tcpport := make(map[string]interface{})
				tcpport["protocol"] = "TCP"
				tcpport["port"] = rule.TCPPorts.ValueString()
				ports = append(ports, tcpport)
			}
			if rule.UDPEnabled.ValueBool() {
				udpport := make(map[string]interface{})
				udpport["protocol"] = "UDP"
				udpport["port"] = rule.UDPPorts.ValueString()
				ports = append(ports, udpport)
			}
			if rule.ICMPEnabled.ValueBool() {
				icmpport := make(map[string]interface{})
				icmpport["protocol"] = "ICMP"
				ports = append(ports, icmpport)
			}
			eips := make([]string, 0)
			diags = rule.ExceptIPBlock.ElementsAs(context.Background(), &eips, false)
			egress = append(egress, map[string]interface{}{
				"type":            "IP_BLOCK",
				"ip_block":        rule.IPBlock.ValueString(),
				"ports":           ports,
				"except_ip_block": eips,
			})
		}
	}
	return map[string]interface{}{
		"global_default_action": state.DefaultAction.ValueString(),
		"global_whitelist": map[string]interface{}{
			"enable":  enabled,
			"egress":  egress,
			"ingress": ingress,
		},
	}, diags
}
