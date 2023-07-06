package everoute_service

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.List = ControllerInstanceValidator{}

func GetControllerInstanceValidator() validator.List {
	return ControllerInstanceValidator{}
}

type ControllerInstanceValidator struct{}

func (v ControllerInstanceValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v ControllerInstanceValidator) MarkdownDescription(_ context.Context) string {
	return "Validate controller instance configuration, make sure instance's number is 3 or 5, and ip address is unique between instances"
}

func (v ControllerInstanceValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() {
		// temporary ignore unknown value
		return
	}
	var instances = make([]ControllerInstanceModel, 0)
	diags := req.ConfigValue.ElementsAs(ctx, &instances, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// instance's length should be 3 or 5
	if len(instances) != 3 && len(instances) != 5 {
		resp.Diagnostics.AddError("invalid length of controll instance configuration", fmt.Sprintf("controller_instance length must be 3 or 5, but got %d", len(instances)))
	}
	// instance should not share same ip
	var ip_instance_map = make(map[string]int)
	var invalid_ip_set = make(map[string]int)
	for idx, instance := range instances {
		ip := instance.IpAddr.String()
		if ip_instance_map[ip] != 0 {
			invalid_ip_set[ip] = 1
		} else {
			ip_instance_map[ip] = idx + 1
		}
	}
	var invalid_ip_list = make([]string, 0, len(invalid_ip_set))
	for ip := range invalid_ip_set {
		invalid_ip_list = append(invalid_ip_list, ip)
	}
	if len(invalid_ip_set) > 0 {
		resp.Diagnostics.AddError("invalid ip address of controller instance configuration", fmt.Sprintf("controller_instance ip address must be unique, but got following duplicate ip %v", invalid_ip_list))
	}
}

func GetAssociatedClusterValidator() validator.List {
	return AssociatedClusterValidator{}
}

type AssociatedClusterValidator struct{}

func (v AssociatedClusterValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v AssociatedClusterValidator) MarkdownDescription(_ context.Context) string {
	return "Validate associated cluster configuration, make sure cluster is unique and vds is unique"
}

func (v AssociatedClusterValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() {
		// temporary ignore unknown value
		return
	}
	var clusters = make([]AssociatedClusterModel, 0)
	diags := req.ConfigValue.ElementsAs(ctx, &clusters, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// should not associated duplicated cluster
	var cluster_instance_map = make(map[string]int)
	var invalid_cluster_set = make(map[string]int)
	for idx, cluster := range clusters {
		cid := cluster.Id.ValueString()
		if cluster_instance_map[cid] != 0 {
			invalid_cluster_set[cid] = 1
		} else {
			cluster_instance_map[cid] = idx + 1
		}
		vds := cluster.VDSes
		var vds_instance_map = make(map[string]int)
		var invalid_vds_set = make(map[string]int)
		// one cluster should not associated duplicated vds
		for idx, vds := range vds {
			vid := vds.Id.ValueString()
			if vds_instance_map[vid] != 0 {
				invalid_vds_set[vid] = 1
			} else {
				vds_instance_map[vid] = idx + 1
			}
		}
		var invalid_vds_list = make([]string, 0, len(invalid_vds_set))
		for vid := range invalid_vds_set {
			invalid_vds_list = append(invalid_vds_list, vid)
		}
		if len(invalid_vds_set) > 0 {
			resp.Diagnostics.AddError("invalid vds of associated cluster configuration", fmt.Sprintf("associated_cluster %s's vds must be unique, but got following duplicate vds %v", cid, invalid_vds_list))
		}
	}
	var invalid_cluster_list = make([]string, 0, len(invalid_cluster_set))
	for cid := range invalid_cluster_set {
		invalid_cluster_list = append(invalid_cluster_list, cid)
	}
	if len(invalid_cluster_set) > 0 {
		resp.Diagnostics.AddError("invalid cluster of associated cluster configuration", fmt.Sprintf("associated_cluster cluster must be unique, but got following duplicate cluster %v", invalid_cluster_list))
	}
}
