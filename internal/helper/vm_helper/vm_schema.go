package vm_helper

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
	"github.com/tidwall/gjson"
)

type VmModel struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
	Ips    types.List   `tfsdk:"ips"`
}

func VmSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "cloudtower's vm schema",
		Computed:            true,
		Attributes:          VmAttributes(),
	}
}

func VmAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "vm's id",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "vm's name",
			Computed:            true,
		},
		"status": schema.StringAttribute{
			MarkdownDescription: "vm's status",
			Computed:            true,
		},
		"ips": schema.ListAttribute{
			MarkdownDescription: "vm's ips",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func VmAttrTypes() map[string]attr.Type {
	return VmSchema().GetType().(types.ObjectType).AttrTypes
}

func ReadSdkVmToModel(input *models.VM, output *VmModel) {
	output.Id = types.StringValue(*input.ID)
	output.Name = types.StringValue(*input.Name)
	output.Status = types.StringValue(string(*input.Status))
	ips := strings.Split(*input.Ips, ",")
	aips := make([]types.String, len(ips))
	for _, ip := range ips {
		aips = append(aips, types.StringValue(ip))
	}
	output.Ips, _ = types.ListValueFrom(context.Background(), types.StringType, aips)
}

func ReadGJsonVmToModel(input *gjson.Result, output *VmModel) {
	output.Id = types.StringValue(input.Get("id").String())
	output.Name = types.StringValue(input.Get("name").String())
	output.Status = types.StringValue(input.Get("status").String())
	ips := strings.Split(input.Get("ips").String(), ",")
	aips := make([]types.String, len(ips))
	for _, ip := range ips {
		aips = append(aips, types.StringValue(ip))
	}
	output.Ips, _ = types.ListValueFrom(context.Background(), types.StringType, aips)
}
