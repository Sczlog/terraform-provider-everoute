package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/smartxworks/terraform-provider-everoute/internal/provider/resources/everoute_service"
	"github.com/smartxworks/terraform-provider-everoute/internal/provider/resources/global_security_policy"
)

func (p *EverouteProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &everoute_service.Resource{} },
		func() resource.Resource { return &global_security_policy.Resource{} },
	}
}
