package everoute_service

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tidwall/gjson"
)

type ControllerModel struct {
	VlanId types.String `tfsdk:"vlan_id"`
	IpAddr types.String `tfsdk:"ip_addr"`
}

func controllerSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "everoute cluster's controller info",
		Computed:            true,
		Optional:            false,
		Required:            false,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"vlan_id": schema.StringAttribute{
					MarkdownDescription: "everoute cluster's controller vlan id",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"ip_addr": schema.StringAttribute{
					MarkdownDescription: "everoute cluster's controller ip address",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
			},
		},
	}
}

func controllerAttrTypes() map[string]attr.Type {
	return controllerSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func flattenController(ctx context.Context, input gjson.Result) []ControllerModel {
	jcontrollers := input.Array()
	result := make([]ControllerModel, 0, len(jcontrollers))

	for _, jcontroller := range jcontrollers {
		result = append(result, ControllerModel{
			VlanId: types.StringValue(jcontroller.Get("vlan").String()),
			IpAddr: types.StringValue(jcontroller.Get("ipAddr").String()),
		})
	}
	return result
}
