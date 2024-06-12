package everoute_service

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AssociatedClusterModel struct {
	Id    types.String         `tfsdk:"id"`
	Name  types.String         `tfsdk:"name"`
	VDSes []AssociatedVdsModel `tfsdk:"vdses"`
}

func associatedClusterSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "elf cluster's schema",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				// TODO: can use both id or name as identifier
				"id": schema.StringAttribute{
					MarkdownDescription: "elf cluster's id",
					Required:            true,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "elf cluster's name",
					Computed:            true,
				},
				"vdses": associatedVdsSchema(),
			},
		},
		Validators: []validator.List{
			GetAssociatedClusterValidator(),
		},
	}
}

type AssociatedVdsModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func associatedVdsSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "elf vds's schema",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "elf vds's id",
					Required:            true,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "elf vds's name",
					Computed:            true,
				},
			},
		},
	}
}
