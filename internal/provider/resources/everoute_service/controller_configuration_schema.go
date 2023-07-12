package everoute_service

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ControllerConfigurationModel struct {
	CluterId   types.String              `tfsdk:"cluster_id"`
	SubnetMask types.String              `tfsdk:"subnet_mask"`
	Gateway    types.String              `tfsdk:"gateway"`
	Instances  []ControllerInstanceModel `tfsdk:"instance"`
}

func controllerConfigurationSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "everoute service's controller configuration",
		Required:            true,
		PlanModifiers: []planmodifier.Object{
			objectplanmodifier.RequiresReplace(),
		},
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "everoute service's controller configuration's cluster id, controllers will be deployed to this cluster",
				Required:            true,
			},
			"subnet_mask": schema.StringAttribute{
				MarkdownDescription: "everoute service's controller configuration's subnet mask, applied to all controllers",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^((128|192|224|240|248|252|254)\.0\.0\.0|(255\.(0|128|192|224|240|248|252|254)\.0\.0)|(255\.255\.(0|128|192|224|240|248|252|254)\.0)|(255\.255\.255\.(0|128|192|224|240|248|252|254)))$`),
						"must be a valid subnet mask"),
				},
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "everoute service's controller configuration's gateway, applied to all controllers",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}$`),
						"must be a valid ip address"),
				},
			},
			"instance": controllerInstanceSchema(),
		},
	}
}

type ControllerInstanceModel struct {
	VlanId types.String `tfsdk:"vlan_id"`
	IpAddr types.String `tfsdk:"ip_addr"`
}

func controllerInstanceSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Required:            true,
		MarkdownDescription: "everoute service's controller configuration's instance configuration",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"vlan_id": schema.StringAttribute{
					MarkdownDescription: "everoute service's controller configuration's controller instance's vlan id",
					Required:            true,
				},
				"ip_addr": schema.StringAttribute{
					MarkdownDescription: "everoute service's controller configuration's controller instance's ip address",
					Required:            true,
					Validators: []validator.String{
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}$`),
							"must be a valid ip address"),
					},
				},
			},
		},
		Validators: []validator.List{
			GetControllerInstanceValidator(),
		},
	}
}
