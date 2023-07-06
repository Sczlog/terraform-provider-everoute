package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/smartxworks/terraform-provider-everoute/internal/provider/datasources/everoute_package"
	"github.com/smartxworks/terraform-provider-everoute/internal/provider/datasources/everoute_service"
	"github.com/smartxworks/terraform-provider-everoute/internal/provider/datasources/global_security_policy"
)

func (p *EverouteProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &everoute_package.DataSource{} },
		func() datasource.DataSource { return &everoute_service.DataSource{} },
		func() datasource.DataSource { return &global_security_policy.DataSource{} },
	}
}
