package everoute_package

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
)

type EveroutePackageResponseModel struct {
	Version      types.String `tfsdk:"version"`
	Architecture types.String `tfsdk:"architecture"`
	Name         types.String `tfsdk:"name"`
	Id           types.String `tfsdk:"id"`
}

func everoutePackageResponseSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "response queried everoute packages",
		Computed:            true,
		Optional:            false,
		Required:            false,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"version": schema.StringAttribute{
					MarkdownDescription: "everoute package version",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"architecture": schema.StringAttribute{
					MarkdownDescription: "everoute package architecture",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "everoute package name",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"id": schema.StringAttribute{
					MarkdownDescription: "everoute package id",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
			},
		},
	}
}

func everoutePackageResponseAttrTypes() map[string]attr.Type {
	return everoutePackageResponseSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func flattenEveroutePackageResponse(ctx context.Context, packages []*models.EveroutePackage) ([]EveroutePackageResponseModel, diag.Diagnostics) {
	result := make([]EveroutePackageResponseModel, 0, len(packages))

	for _, pkg := range packages {
		var arch string
		switch *pkg.Arch {
		case models.ArchitectureAARCH64:
			arch = "AARCH64"
		case models.ArchitectureX8664:
			arch = "X86_64"
		}
		result = append(result, EveroutePackageResponseModel{
			Version:      types.StringValue(*pkg.Version),
			Architecture: types.StringValue(arch),
			Name:         types.StringValue(*pkg.Name),
			Id:           types.StringValue(*pkg.ID),
		})
	}
	return result, nil
}
