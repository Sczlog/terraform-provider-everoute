package everoute_service

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/tidwall/gjson"
)

type AssociatedClusterModel struct {
	Id    types.String         `tfsdk:"id"`
	Name  types.String         `tfsdk:"name"`
	VDSes []AssociatedVdsModel `tfsdk:"vdses"`
}

func associatedClusterSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "associated cluster's schema",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "associated cluster's id",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "associated cluster's name",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"vdses": associatedVdsSchema(),
			},
		},
	}
}

func associatedClusterAttrTypes() map[string]attr.Type {
	return associatedClusterSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func readAssociatedCluster(ctx context.Context, clusteri gjson.Result, vdsi gjson.Result) []AssociatedClusterModel {
	jclusters := clusteri.Array()
	result := make([]AssociatedClusterModel, 0, len(jclusters))
	vdsmap := readAssociatedVdses(ctx, vdsi)
	for _, jcluster := range jclusters {
		vds, ok := vdsmap[jcluster.Get("id").String()]
		cluster := AssociatedClusterModel{
			Id:   types.StringValue(jcluster.Get("id").String()),
			Name: types.StringValue(jcluster.Get("name").String()),
		}
		if ok {
			cluster.VDSes = vds
		}
		result = append(result, cluster)
	}
	return result
}

type AssociatedVdsModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func associatedVdsSchema() schema.Attribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "associated vds's schema",
		Computed:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "associated vds's id",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
				"name": schema.StringAttribute{
					MarkdownDescription: "associated vds's name",
					Computed:            true,
					Optional:            false,
					Required:            false,
				},
			},
		},
	}
}

func associatedVdsAttrTypes() map[string]attr.Type {
	return associatedVdsSchema().GetType().(types.ListType).ElemType.(types.ObjectType).AttrTypes
}

func readAssociatedVdses(ctx context.Context, vdsi gjson.Result) map[string][]AssociatedVdsModel {
	jvdses := vdsi.Array()
	result := make(map[string][]AssociatedVdsModel)

	for _, jvds := range jvdses {
		cid := jvds.Get("cluster.id").String()
		if _, ok := result[cid]; !ok {
			result[cid] = make([]AssociatedVdsModel, 0)
		}
		vds := AssociatedVdsModel{
			Id:   types.StringValue(jvds.Get("id").String()),
			Name: types.StringValue(jvds.Get("name").String()),
		}
		result[cid] = append(result[cid], vds)
	}
	return result
}
