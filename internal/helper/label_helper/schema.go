package label_helper

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
	"github.com/tidwall/gjson"
)

type LabelModel struct {
	Id    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func LabelSchema() dschema.Attribute {
	return dschema.SingleNestedAttribute{
		MarkdownDescription: "cloudtower's label schema",
		Computed:            true,
		Attributes:          LabelAttrs(),
	}
}

func LabelAttrs() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"id": dschema.StringAttribute{
			MarkdownDescription: "label's id",
			Computed:            true,
		},
		"key": dschema.StringAttribute{
			MarkdownDescription: "label's key",
			Computed:            true,
		},
		"value": dschema.StringAttribute{
			MarkdownDescription: "label's value",
			Computed:            true,
		},
	}
}

func LabelResourceAttrs() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id": rschema.StringAttribute{
			MarkdownDescription: "label's id",
			Optional:            true,
			Computed:            true,
		},
		"key": rschema.StringAttribute{
			MarkdownDescription: "label's key",
			Optional:            true,
			Computed:            true,
		},
		"value": rschema.StringAttribute{
			MarkdownDescription: "label's value",
			Optional:            true,
			Computed:            true,
		},
	}
}

func LabelAttrTypes() map[string]attr.Type {
	return LabelSchema().GetType().(types.ObjectType).AttrTypes
}

func ReadSdkLabelToModel(input *models.Label, output *LabelModel) {
	output.Id = types.StringValue(*input.ID)
	output.Key = types.StringValue(*input.Key)
	output.Value = types.StringValue(*input.Value)
}

func ReadGJsonLabelToModel(input *gjson.Result, output *LabelModel) {
	output.Id = types.StringValue(input.Get("id").String())
	output.Key = types.StringValue(input.Get("key").String())
	output.Value = types.StringValue(input.Get("value").String())
}
