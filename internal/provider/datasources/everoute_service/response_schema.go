package everoute_service

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tidwall/gjson"
)

type EverouteServiceResponseModel struct {
	Id                     types.String             `tfsdk:"id"`
	Name                   types.String             `tfsdk:"name"`
	Version                types.String             `tfsdk:"version"`
	Controllers            []ControllerModel        `tfsdk:"controllers"`
	GlobalDefaultAction    types.String             `tfsdk:"global_default_action"`
	GlobalWhiteListEnabled types.Bool               `tfsdk:"global_whitelist_enabled"`
	Phase                  types.String             `tfsdk:"phase"`
	AssociatedClusters     []AssociatedClusterModel `tfsdk:"associated_clusters"`
	// Status              types.Object `tfsdk:"status"`
	Installed types.Bool `tfsdk:"installed"`
}

func everouteServiceResponseSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "response queried everoute services",
		Computed:            true,
		Optional:            false,
		Required:            false,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "everoute service's id",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "everoute service's name",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"version": schema.StringAttribute{
					MarkdownDescription: "everoute service's version",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"controllers": controllerSchema(),
				"global_default_action": schema.StringAttribute{
					MarkdownDescription: "everoute service's global default action",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"global_whitelist_enabled": schema.BoolAttribute{
					MarkdownDescription: "if global whitelist enabled",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"phase": schema.StringAttribute{
					MarkdownDescription: "everoute service's phase",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"elf_clusters": associatedClusterSchema(),
				"elf_vdses":    associatedVdsSchema(),
				// "status":       statusSchema(),
				"installed": schema.BoolAttribute{
					MarkdownDescription: "if everoute service is installed",
					Computed:            true,
				},
			},
		},
	}
}

func everouteClusterResponseAttrTypes() map[string]attr.Type {
	return everouteServiceResponseSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func flattenEverouteServiceResponse(ctx context.Context, input *gjson.Result) []EverouteServiceResponseModel {
	jservices := input.Get("everouteClusters").Array()
	result := make([]EverouteServiceResponseModel, 0, len(jservices))
	for _, jservice := range jservices {
		service := EverouteServiceResponseModel{}
		service.Id = types.StringValue(jservice.Get("id").String())
		service.Name = types.StringValue(jservice.Get("name").String())
		service.Version = types.StringValue(jservice.Get("version").String())
		service.Phase = types.StringValue(jservice.Get("phase").String())
		service.Installed = types.BoolValue(jservice.Get("installed").Bool())
		service.GlobalDefaultAction = types.StringValue(jservice.Get("global_default_action").String())

		controllers := flattenController(ctx, jservice.Get("controller_instances"))
		service.Controllers = controllers

		jwhitelist := jservice.Get("global_whitelist.enable")
		if jwhitelist.Exists() {
			service.GlobalWhiteListEnabled = types.BoolValue(jwhitelist.Bool())
		} else {
			service.GlobalWhiteListEnabled = types.BoolValue(false)
		}

		elfCluster := readAssociatedCluster(ctx, jservice.Get("agent_elf_clusters"), jservice.Get("agent_elf_vdses"))
		service.AssociatedClusters = elfCluster

		result = append(result, service)
	}
	return result
}
