package global_security_policy

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tidwall/gjson"
)

type NetworkPolicyRuleModel struct {
	IPBlock       types.String `tfsdk:"ip_block"`
	ExceptIPBlock types.List   `tfsdk:"except_ip_block"`
	TCPEnabled    types.Bool   `tfsdk:"tcp_enabled"`
	TCPPorts      types.String `tfsdk:"tcp_ports"`
	UDPEnabled    types.Bool   `tfsdk:"udp_enabled"`
	UDPPorts      types.String `tfsdk:"udp_ports"`
	ICMPEnabled   types.Bool   `tfsdk:"icmp_enabled"`
}

func networkPolicyRuleSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"ip_block": schema.StringAttribute{
				MarkdownDescription: "network policy rule included ip block",
				Required:            true,
			},
			"except_ip_block": schema.ListAttribute{
				MarkdownDescription: "network policy rule excluded ip block",
				ElementType:         types.StringType,
				Default: listdefault.StaticValue(
					types.ListValueMust(types.StringType, []attr.Value{}),
				),
				Optional: true,
				Computed: true,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
				},
			},
			"tcp_enabled": schema.BoolAttribute{
				MarkdownDescription: "if network policy is enabled for tcp protocol",
				Default:             booldefault.StaticBool(true),
				Optional:            true,
				Computed:            true,
			},
			"tcp_ports": schema.StringAttribute{
				MarkdownDescription: "network policy rule's tcp port, seperate by comma",
				Default:             stringdefault.StaticString(""),
				Optional:            true,
				Computed:            true,
			},
			"udp_enabled": schema.BoolAttribute{
				MarkdownDescription: "if network policy is enabled for udp protocol",
				Default:             booldefault.StaticBool(true),
				Optional:            true,
				Computed:            true,
			},
			"udp_ports": schema.StringAttribute{
				MarkdownDescription: "network policy rule's udp port, seperate by comma",
				Default:             stringdefault.StaticString(""),
				Optional:            true,
				Computed:            true,
			},
			"icmp_enabled": schema.BoolAttribute{
				MarkdownDescription: "if network policy is enabled for icmp protocol",
				Default:             booldefault.StaticBool(true),
				Optional:            true,
				Computed:            true,
			},
		},
		Validators: []validator.Object{
			&NetworkRulePolicyValidator{},
		},
	}
}

func readGqlResultToNetworkPolicyRule(input *gjson.Result) []NetworkPolicyRuleModel {
	rules := input.Array()
	var result = make([]NetworkPolicyRuleModel, len(rules))
	for idx, rule := range rules {
		result[idx].IPBlock = types.StringValue(rule.Get("ip_block").String())
		jeipb := rule.Get("except_ip_block")
		var eipb []attr.Value
		if jeipb.Type == gjson.Null {
			eipb = make([]attr.Value, 0)
		} else {
			eipb = make([]attr.Value, len(jeipb.Array()))
			for i, v := range jeipb.Array() {
				eipb[i] = types.StringValue(v.String())
			}
		}
		meipb, _ := types.ListValue(types.StringType, eipb)
		result[idx].ExceptIPBlock = meipb
		jports := rule.Get("ports")
		if jports.Type != gjson.Null {
			jportsArr := jports.Array()
			// if no ports configuration, mean all protocol is enabled
			result[idx].TCPEnabled = types.BoolValue(len(jportsArr) == 0)
			result[idx].TCPPorts = types.StringValue("")
			result[idx].UDPEnabled = types.BoolValue(len(jportsArr) == 0)
			result[idx].UDPPorts = types.StringValue("")
			result[idx].ICMPEnabled = types.BoolValue(len(jportsArr) == 0)

			for _, jport := range jportsArr {
				jpprotocol := jport.Get("protocol").String()
				switch strings.ToUpper(jpprotocol) {
				case "TCP":
					result[idx].TCPEnabled = types.BoolValue(true)
					result[idx].TCPPorts = types.StringValue(jport.Get("port").String())
				case "UDP":
					result[idx].UDPEnabled = types.BoolValue(true)
					result[idx].UDPPorts = types.StringValue(jport.Get("port").String())
				case "ICMP":
					result[idx].ICMPEnabled = types.BoolValue(true)
				}
			}
		}
	}
	return result
}
