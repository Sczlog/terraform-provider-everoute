package global_security_policy

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smartxworks/terraform-provider-everoute/internal/helper/label_helper"
	"github.com/tidwall/gjson"
)

type GlobalSecurityPolicyResponseModel struct {
	Enable  types.Bool               `tfsdk:"enable"`
	Ingress []NetworkPolicyRuleModel `tfsdk:"ingress"`
	Egress  []NetworkPolicyRuleModel `tfsdk:"egress"`
}

func globalSecurityPolicyResponseSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "everoute cluster's global white list configuration",
		Computed:            true,
		Attributes: map[string]schema.Attribute{
			"enable": schema.BoolAttribute{
				MarkdownDescription: "if global white list is enabled for current everoute cluster",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
			"ingress": schema.ListNestedAttribute{
				MarkdownDescription: "everoute cluster's global white list ingress rule",
				Computed:            true,
				Optional:            false,
				Required:            false,
				NestedObject:        networkPolicyRuleSchema(),
			},
			"egress": schema.ListNestedAttribute{
				MarkdownDescription: "everoute cluster's global white list egress rule",
				Computed:            true,
				Optional:            false,
				Required:            false,
				NestedObject:        networkPolicyRuleSchema(),
			},
		},
	}
}

func serviceGlobalWhiteListAttrTypes() map[string]attr.Type {
	return globalSecurityPolicyResponseSchema().GetType().(types.ObjectType).AttrTypes
}

func readGlobalSecurityPolicy(ctx context.Context, input *gjson.Result) GlobalSecurityPolicyResponseModel {
	state := GlobalSecurityPolicyResponseModel{}
	gje := input.Get("enable")
	if gje.Exists() {
		state.Enable = types.BoolValue(gje.Bool())
	} else {
		state.Enable = types.BoolValue(false)
	}
	gji := input.Get("ingress")
	if gji.Exists() {
		ingress := flattenNetworkPolicyRule(ctx, &gji)
		state.Ingress = ingress
	}
	gjeg := input.Get("egress")
	if gjeg.Exists() {
		egress := flattenNetworkPolicyRule(ctx, &gjeg)
		state.Egress = egress
	}
	return state
}

type NetworkPolicyRuleModel struct {
	IpBlock       types.String                 `tfsdk:"ip_block"`
	ExceptIpBlock []types.String               `tfsdk:"except_ip_block"`
	Ports         []NetworkPolicyRulePortModel `tfsdk:"ports"`
	Selectors     []label_helper.LabelModel    `tfsdk:"selectors"`
	Type          types.String                 `tfsdk:"type"`
}

func networkPolicyRuleSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"ip_block": schema.StringAttribute{
				MarkdownDescription: "network policy rule included ip block",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
			"except_ip_block": schema.ListAttribute{
				MarkdownDescription: "network policy rule excluded ip block",
				Computed:            true,
				Optional:            false,
				Required:            false,
				ElementType:         types.StringType,
			},
			"ports": schema.ListNestedAttribute{
				MarkdownDescription: "network policy rule ports configuration",
				Computed:            true,
				Optional:            false,
				Required:            false,
				NestedObject:        networkPolicyRulePortSchema(),
			},
			"selectors": schema.ListNestedAttribute{
				MarkdownDescription: "network policy rule selector labels",
				Computed:            true,
				Optional:            false,
				Required:            false,
				NestedObject: schema.NestedAttributeObject{
					Attributes: label_helper.LabelAttrs(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "network policy rule type",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
		},
	}
}

func networkPolicyRuleAttrType() map[string]attr.Type {
	return networkPolicyRuleSchema().Type().(types.ObjectType).AttrTypes
}

func flattenNetworkPolicyRule(ctx context.Context, input *gjson.Result) []NetworkPolicyRuleModel {
	rules := input.Array()
	var result = make([]NetworkPolicyRuleModel, 0, len(rules))
	for _, rule := range rules {
		state := NetworkPolicyRuleModel{}
		gjip := rule.Get("ip_block")
		if gjip.Type == gjson.String {
			state.IpBlock = types.StringValue(gjip.String())
		}
		gjeipb := rule.Get("except_ip_block")
		gjeipbl := rule.Get("except_ip_block.#").Int()
		if gjeipb.Exists() && gjeipbl > 0 {
			for i := 0; i < int(gjeipbl); i++ {
				eipb := gjeipb.Get(fmt.Sprintf("%d", i))
				state.ExceptIpBlock = append(state.ExceptIpBlock, types.StringValue(eipb.String()))
			}
		}
		gjps := rule.Get("ports")
		if gjps.Exists() {
			ports := flattenNetworkPolicyRulePorts(ctx, &gjps)
			state.Ports = ports
		}
		gjt := rule.Get("type")
		if gjt.Exists() {
			state.Type = types.StringValue(gjt.String())
		}
		gjss := rule.Get("selectors")
		if gjss.Exists() {
			for _, gjs := range gjss.Array() {
				selector := label_helper.LabelModel{}
				label_helper.ReadGJsonLabelToModel(&gjs, &selector)
				state.Selectors = append(state.Selectors, selector)
			}
		}
		result = append(result, state)
	}
	return result
}

type NetworkPolicyRulePortModel struct {
	Protocol types.String   `tfsdk:"protocol"`
	Port     []types.String `tfsdk:"port"`
}

func networkPolicyRulePortSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"protocol": schema.StringAttribute{
				MarkdownDescription: "network policy rule's protocol, valid value: TCP, UDP, ICMP",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
			"port": schema.ListAttribute{
				MarkdownDescription: "network policy rule's port",
				Computed:            true,
				Optional:            false,
				Required:            false,
				ElementType:         types.StringType,
			},
		},
	}
}

func NetworkPolicyRulePortAttrType() map[string]attr.Type {
	return networkPolicyRulePortSchema().Type().(types.ObjectType).AttrTypes
}

func flattenNetworkPolicyRulePorts(ctx context.Context, input *gjson.Result) []NetworkPolicyRulePortModel {
	ports := input.Array()
	result := make([]NetworkPolicyRulePortModel, 0, len(ports))
	for _, p := range ports {
		port := NetworkPolicyRulePortModel{}
		gjps := p.Get("port")
		if gjps.Exists() && gjps.Type == gjson.String {
			ports := strings.Split(gjps.String(), ",")
			for _, p := range ports {
				port.Port = append(port.Port, types.StringValue(p))
			}
		}
		gjpt := p.Get("protocol")
		if gjpt.Type == gjson.String {
			port.Protocol = types.StringValue(gjpt.String())
		}
		result = append(result, port)
	}
	return result
}
