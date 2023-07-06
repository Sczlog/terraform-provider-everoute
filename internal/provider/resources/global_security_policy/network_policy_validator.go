package global_security_policy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.Object = &NetworkRulePolicyValidator{}

type NetworkRulePolicyValidator struct {
}

func (v *NetworkRulePolicyValidator) Description(context.Context) string {
	return "validate network policy rule"
}

func (v *NetworkRulePolicyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}
func (v *NetworkRulePolicyValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, res *validator.ObjectResponse) {
	attrs := req.ConfigValue.Attributes()
	ate := attrs["tcp_enabled"].(types.Bool)
	aue := attrs["udp_enabled"].(types.Bool)
	aie := attrs["icmp_enabled"].(types.Bool)
	te := true
	ue := true
	ie := true
	if !ate.IsNull() {
		te = ate.ValueBool()
	}
	if !aue.IsNull() {
		ue = aue.ValueBool()
	}
	if !aie.IsNull() {
		ie = aie.ValueBool()
	}
	if !te && !ue && !ie {
		res.Diagnostics.AddError(
			"Failed to validate network policy rule",
			"at least one protocol should be enabled, otherwise remove this rule from ingress or egress",
		)
	}
	if !te {
		tp := attrs["tcp_ports"].(types.String)
		if !tp.IsNull() && tp.ValueString() != "" {
			res.Diagnostics.AddError(
				"Failed to validate network policy rule",
				"if tcp protocol is disabled, tcp ports must be empty",
			)
		}
	}
	if !ue {
		up := attrs["udp_ports"].(types.String)
		if !up.IsNull() && up.ValueString() != "" {
			res.Diagnostics.AddError(
				"Failed to validate network policy rule",
				"if udp protocol is disabled, udp ports must be empty",
			)
		}
	}
}
