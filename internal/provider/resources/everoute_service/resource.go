package everoute_service

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/smartxworks/cloudtower-go-sdk/v2/client/cluster"
	erp "github.com/smartxworks/cloudtower-go-sdk/v2/client/everoute_package"
	"github.com/smartxworks/cloudtower-go-sdk/v2/client/vds"
	"github.com/smartxworks/cloudtower-go-sdk/v2/client/vlan"
	"github.com/smartxworks/cloudtower-go-sdk/v2/models"
	utils "github.com/smartxworks/cloudtower-go-sdk/v2/utils"
	"github.com/smartxworks/terraform-provider-everoute/internal/everoute"
	"github.com/tidwall/gjson"
	g "github.com/zyedidia/generic"
	"github.com/zyedidia/generic/hashset"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *everoute.Client
}

// EverouteServiceResourceModel describes the resource data model.
type EverouteServiceResourceModel struct {
	Id                      types.String                 `tfsdk:"id"`
	Name                    types.String                 `tfsdk:"name"`
	PackageId               types.String                 `tfsdk:"package_id"`
	ControllerConfiguration ControllerConfigurationModel `tfsdk:"controller_configuration"`
	AssociatedCluster       []AssociatedClusterModel     `tfsdk:"associated_cluster"`
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "resource handling an everoute service",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "everoute service's identifier",
				Computed:            true,
				Required:            false,
				Optional:            false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "everoute service's name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"package_id": schema.StringAttribute{
				MarkdownDescription: "everoute service's package id",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"controller_configuration": controllerConfigurationSchema(),
			"associated_cluster":       associatedClusterSchema(),
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*everoute.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *EverouteServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// check duplicate everoute services
	client_resp, _, err := r.client.DgqlApi.Raw(ctx, duplicatedNameServiceDocument, "everouteClusters", map[string]interface{}{
		"where": map[string]interface{}{
			"name": data.Name.ValueString(),
		},
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Create everoute service failed",
			fmt.Sprintf("Unable to check same name everoute service, got error: %s", err),
		)
	}
	if client_resp.Get("everouteClusters.#").Int() > 0 {
		resp.Diagnostics.AddError(
			"Create everoute service failed",
			"Same name everoute service already exists",
		)
	}
	// check package id if existed
	gerpp := erp.NewGetEveroutePackagesParams()
	gerpp.RequestBody = &models.GetEveroutePackagesRequestBody{
		Where: &models.EveroutePackageWhereInput{
			ID: data.PackageId.ValueStringPointer(),
		},
	}
	erps, err := r.client.Api.EveroutePackage.GetEveroutePackages(gerpp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Create everoute service failed",
			fmt.Sprintf("Unable to check package id, got error: %s", err),
		)
	}
	if len(erps.Payload) == 0 {
		resp.Diagnostics.AddError(
			"Create everoute service failed",
			fmt.Sprintf("Package id %s not exist", data.PackageId.ValueString()),
		)
	}
	// check if cluster exist
	gcp := cluster.NewGetClustersParams()
	gcp.RequestBody = &models.GetClustersRequestBody{
		Where: &models.ClusterWhereInput{
			ID: data.ControllerConfiguration.CluterId.ValueStringPointer(),
		},
	}
	cs, err := r.client.Api.Cluster.GetClusters(gcp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Create everoute service failed",
			fmt.Sprintf("Unable to check cluster id, got error: %s", err),
		)
	} else {
		if len(cs.Payload) == 0 {
			resp.Diagnostics.AddError(
				"Create everoute service failed",
				fmt.Sprintf("Cluster id %s not exist", data.ControllerConfiguration.CluterId.ValueString()),
			)
		} else {
			// check if vlan exist and belongs to selected cluster, only do when cluster existed
			cid := cs.Payload[0].ID
			gvp := vlan.NewGetVlansParams()
			gvp.RequestBody = &models.GetVlansRequestBody{
				Where: &models.VlanWhereInput{
					Vds: &models.VdsWhereInput{
						Cluster: &models.ClusterWhereInput{
							ID: cid,
						},
					},
				},
			}
			ovls, err := r.client.Api.Vlan.GetVlans(gvp)
			if err != nil {
				resp.Diagnostics.AddError(
					"Create everoute service failed",
					fmt.Sprintf("Unable to check vlan, got error: %s", err),
				)
			}
			vlan_id_set := make(map[string]bool)
			for _, ovl := range ovls.Payload {
				vlan_id_set[*ovl.ID] = true
			}
			for _, ist := range data.ControllerConfiguration.Instances {
				if _, ok := vlan_id_set[ist.VlanId.ValueString()]; !ok {
					resp.Diagnostics.AddError(
						"Create everoute service failed",
						fmt.Sprintf("Vlan id %s not exist or not belongs to cluster %s", ist.VlanId.ValueString(), *cid),
					)
				}
			}
		}
	}

	// check associated clusters is exist
	aclength := len(data.AssociatedCluster)
	var acIds = make([]string, 0, aclength)
	var acIdMap = make(map[string]*AssociatedClusterModel)
	var outerAssociatedVdsIds = make([]string, 0)
	if aclength > 0 {
		for _, ac := range data.AssociatedCluster {
			acid := ac.Id.ValueString()
			ac := ac
			acIdMap[acid] = &ac
			acIds = append(acIds, acid)
		}
		gcp = cluster.NewGetClustersParams()
		gcp.RequestBody = &models.GetClustersRequestBody{
			Where: &models.ClusterWhereInput{
				IDIn: acIds,
			},
		}
		cs, err = r.client.Api.Cluster.GetClusters(gcp)
		if err != nil {
			resp.Diagnostics.AddError(
				"Create everoute service failed",
				fmt.Sprintf("Unable to check associated cluster id, got error: %s", err),
			)
		} else {
			rawAssociatedClusterIds := make([]string, 0, len(cs.Payload))
			for _, c := range cs.Payload {
				// check if vds is existed under associated cluster
				acm, ok := acIdMap[*c.ID]
				if !ok {
					resp.Diagnostics.AddError(
						"Create everoute service failed",
						fmt.Sprintf("Associated cluster %s not exist in state but readed", *c.ID),
					)
					continue
				}
				rawAssociatedClusterIds = append(rawAssociatedClusterIds, *c.ID)
				acm.Name = types.StringValue(*c.Name)
				acvdsm := acm.VDSes
				if len(acm.VDSes) > 0 {
					acvdsids := make([]string, 0, len(acvdsm))
					acvds_id_map := make(map[string]*AssociatedVdsModel)
					for _, acvds := range acvdsm {
						acvdsid := acvds.Id.ValueString()
						acvds_id_map[acvdsid] = &acvds
						acvdsids = append(acvdsids, acvdsid)
					}
					gvdsp := vds.NewGetVdsesParams()
					gvdsp.RequestBody = &models.GetVdsesRequestBody{
						Where: &models.VdsWhereInput{
							IDIn: acvdsids,
							Cluster: &models.ClusterWhereInput{
								ID: c.ID,
							},
						},
					}
					vdses, err := r.client.Api.Vds.GetVdses(gvdsp)
					if err != nil {
						resp.Diagnostics.AddError(
							"Create everoute service failed",
							fmt.Sprintf("Unable to check vds, got error: %s", err),
						)
					} else {
						for _, vds := range vdses.Payload {
							acvdsm, ok := acvds_id_map[*vds.ID]
							if !ok {
								resp.Diagnostics.AddError(
									"Create everoute service failed",
									fmt.Sprintf("Associated vds %s not exist in state but readed", *vds.ID),
								)
							}
							acvdsm.Name = types.StringValue(*vds.Name)
						}
						if len(vdses.Payload) != len(acvdsids) {
							resp.Diagnostics.AddError(
								"Create everoute service failed",
								fmt.Sprintf("Some of associated vds not exist, expected: %v, response: %v", acvdsids, vdses.Payload),
							)
						} else {
							outerAssociatedVdsIds = append(outerAssociatedVdsIds, acvdsids...)
						}
					}
				}
			}
			// check if missing some associated cluster
			if len(rawAssociatedClusterIds) != len(acIds) {
				resp.Diagnostics.AddError(
					"Create everoute service failed",
					fmt.Sprintf("Some of associated cluster not exist, expected: %v, response: %v", acIds, rawAssociatedClusterIds),
				)
			}
		}
	}

	// check if has prerequiest error before deploy everoute service
	if resp.Diagnostics.HasError() {
		return
	}
	controllerInstance := make([]map[string]interface{}, 0)
	for _, ist := range data.ControllerConfiguration.Instances {
		controllerInstance = append(controllerInstance, map[string]interface{}{
			"vlan":   ist.VlanId.ValueString(),
			"ipAddr": ist.IpAddr.ValueString(),
		})
	}
	client_resp, headers, err := r.client.DgqlApi.Raw(ctx, deployEverouteServiceDocument, "deployEverouteCluster", map[string]interface{}{
		"data": map[string]interface{}{
			"name":    data.Name.ValueString(),
			"version": erps.Payload[0].Version,
			"controller_template": map[string]interface{}{
				"cluster": data.ControllerConfiguration.CluterId.ValueString(),
				"vcpu":    2,
				"memory":  2,
				"size":    30,
				"netmask": data.ControllerConfiguration.SubnetMask.ValueString(),
				"gateway": data.ControllerConfiguration.Gateway.ValueString(),
			},
			"controller_instances": controllerInstance,
			"status":               map[string]interface{}{},
			// configure global whitelist configuration in global whitelist resource
			"global_default_action": "ALLOW",
			"global_whitelist": map[string]interface{}{
				"enable":  false,
				"egress":  []struct{}{},
				"ingress": []struct{}{},
			},
		},
		"effect": map[string]interface{}{
			"package": map[string]interface{}{
				"id": data.PackageId.ValueString(),
			},
		},
	}, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Create everoute service failed",
			fmt.Sprintf("Unable to deploy everoute service, got error: %s", err),
		)
		return
	}

	// wait until task finished
	taskid := headers.Get("X-Task-Id")
	err = utils.WaitTask(ctx, r.client.Api, &taskid, 10*time.Second)
	if err != nil {
		resp.Diagnostics.AddError(
			"Create everoute service failed",
			fmt.Sprintf("Task not complete successfully, got error: %s", err),
		)
		return
	}
	sid := client_resp.Get("createEverouteCluster.id").String()
	// associated everoute service with cluster
	if aclength > 0 {
		acidconnect := make([]map[string]interface{}, 0)
		acvdsset := make([]map[string]interface{}, 0)
		for _, acid := range acIds {
			acidconnect = append(acidconnect, map[string]interface{}{
				"id": acid,
			})
		}
		for _, acvdsid := range outerAssociatedVdsIds {
			acvdsset = append(acvdsset, map[string]interface{}{
				"id": acvdsid,
			})
		}
		_, headers, err = r.client.DgqlApi.Raw(ctx, associatedClusterDocument, "updateEverouteClusterAssociation", map[string]interface{}{
			"where": map[string]interface{}{
				"id": sid,
			},
			"data": map[string]interface{}{
				"agent_elf_clusters": map[string]interface{}{
					"connect": acidconnect,
				},
				"agent_elf_vdses": map[string]interface{}{
					"set": acvdsset,
				},
			},
		}, nil)
		if err != nil {
			resp.Diagnostics.AddError(
				"Create everoute service failed",
				fmt.Sprintf("Unable to associated everoute service with cluster, got error: %s", err),
			)
			// set cluster configuration to empty array
			data.AssociatedCluster = []AssociatedClusterModel{}
			return
		}

		taskid = headers.Get("X-Task-Id")
		err = utils.WaitTask(ctx, r.client.Api, &taskid, 10*time.Second)
		if err != nil {
			resp.Diagnostics.AddError(
				"Create everoute service failed",
				fmt.Sprintf("Task not complete successfully, got error: %s", err),
			)
			return
		}
	}

	data.Id = types.StringValue(sid)

	cluster, diags := getEverouteServiceGqlResult(ctx, r.client, data.Id.ValueString())
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	readGqlResultToState(cluster, data, r.client)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *EverouteServiceResourceModel
	var id string
	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	if diags.HasError() {
		// for import state usage, if can get id from state, use it to query
		idiags := req.State.GetAttribute(ctx, path.Root("id"), &id)
		// create a new datamodel
		data = &EverouteServiceResourceModel{}
		if idiags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		id = data.Id.ValueString()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	cluster, diags := getEverouteServiceGqlResult(ctx, r.client, id)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = readGqlResultToState(cluster, data, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *EverouteServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterParams, vdsesParam := diffAssociatedClusters(&plan.AssociatedCluster, &state.AssociatedCluster)

	id := state.Id.ValueString()

	_, headers, err := r.client.DgqlApi.Raw(ctx, associatedClusterDocument, "updateEverouteClusterAssociation", map[string]interface{}{
		"where": map[string]interface{}{
			"id": id,
		},
		"data": map[string]interface{}{
			"agent_elf_clusters": clusterParams,
			"agent_elf_vdses":    vdsesParam,
		},
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Update everoute service failed",
			fmt.Sprintf("Unable to update everoute service, got error: %s", err),
		)
		return
	}
	taskid := headers.Get("X-Task-Id")
	err = utils.WaitTask(ctx, r.client.Api, &taskid, 10*time.Second)
	if err != nil {
		resp.Diagnostics.AddError(
			"Update everoute service failed",
			fmt.Sprintf("Task not complete successfully, got error: %s", err),
		)
		return
	}

	// re-read the everoute service after update

	cluster, diags := getEverouteServiceGqlResult(ctx, r.client, id)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	diags = readGqlResultToState(cluster, plan, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Id = state.Id

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *EverouteServiceResourceModel

	// delete all related security policy

	// un associate all related cluster

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	id := data.Id.ValueString()
	cluster, diags := getEverouteServiceGqlResult(ctx, r.client, id)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	// update state to time
	// unassociate all clusters
	aciddisconnect := make([]map[string]interface{}, 0)
	for _, aec := range cluster.Get("agent_elf_clusters").Array() {
		aciddisconnect = append(aciddisconnect, map[string]interface{}{
			"id": aec.Get("id").String(),
		})
	}
	_, headers, err := r.client.DgqlApi.Raw(ctx, associatedClusterDocument, "updateEverouteClusterAssociation", map[string]interface{}{
		"where": map[string]interface{}{
			"id": id,
		},
		"data": map[string]interface{}{
			"agent_elf_clusters": map[string]interface{}{
				"disconnect": aciddisconnect,
			},
			"agent_elf_vdses": map[string]interface{}{
				"set": []string{},
			},
		},
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unassociate everoute service", fmt.Sprintf("Unable to unassociate cluster from everoute service, got error: %s", err))
		return
	}
	taskid := headers.Get("X-Task-Id")
	err = utils.WaitTask(ctx, r.client.Api, &taskid, 10*time.Second)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unassociate everoute service", fmt.Sprintf("Unable to unassociate cluster from everoute service, got error: %s", err))
		return
	}
	// delete everoute services
	_, headers, err = r.client.DgqlApi.Raw(ctx, deleteEverouteServiceDocument, "deleteEverouteCluster", map[string]interface{}{
		"where": map[string]interface{}{
			"id": id,
		},
	}, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete everoute service", fmt.Sprintf("Unable to delete everoute service, got error: %s", err))
		return
	}
	taskid = headers.Get("X-Task-Id")
	err = utils.WaitTask(ctx, r.client.Api, &taskid, 10*time.Second)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete everoute service", fmt.Sprintf("Unable to delete everoute service, got error: %s", err))
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func getEverouteServiceGqlResult(ctx context.Context, client *everoute.Client, id string) (*gjson.Result, diag.Diagnostics) {
	var diags diag.Diagnostics
	result, _, err := client.DgqlApi.Raw(ctx, getEverouteServiceDocument, "everouteClusters", map[string]interface{}{
		"where": map[string]interface{}{
			"id": id,
		},
	}, nil)
	if err != nil {
		diags.AddError("Failed to read everoute service", fmt.Sprintf("Unable to read everoute service, got error: %s", err))
		return nil, diags
	}

	cluster := result.Get("everouteClusters.0")

	if cluster.Type == gjson.Null {
		diags.AddError("Failed to read everoute service", fmt.Sprintf("Cannot find everoute service %s", id))
		return nil, diags
	}
	return &cluster, diags
}

func readGqlResultToState(input *gjson.Result, state *EverouteServiceResourceModel, client *everoute.Client) diag.Diagnostics {
	var diagnostic diag.Diagnostics

	state.Id = types.StringValue(input.Get("id").String())
	state.Name = types.StringValue(input.Get("name").String())

	// read cluster configuration
	state.ControllerConfiguration.CluterId = types.StringValue(input.Get("controller_template.cluster").String())
	state.ControllerConfiguration.Gateway = types.StringValue(input.Get("controller_template.gateway").String())
	state.ControllerConfiguration.SubnetMask = types.StringValue(input.Get("controller_template.netmask").String())

	if len(state.ControllerConfiguration.Instances) != 0 {
		// state has configuration
		hm := hashset.New[string](uint64(len(state.ControllerConfiguration.Instances)), g.Equals[string], g.HashString)
		for _, instance := range state.ControllerConfiguration.Instances {
			hm.Put(instance.IpAddr.ValueString())
		}
		for _, instance := range input.Get("controller_instances").Array() {
			ipaddr := instance.Get("ipAddr").String()
			ok := hm.Has(ipaddr)
			if !ok {
				diagnostic.AddError("Failed to read everoute service", "Inconsistent controller instance")
			}
			hm.Remove(ipaddr)
		}
		if hm.Size() != 0 {
			diagnostic.AddError("Failed to read everoute service", "Inconsistent controller instance")
		}
	} else {
		// state doesn't have configuration
		for _, instance := range input.Get("controller_instances").Array() {
			state.ControllerConfiguration.Instances = append(state.ControllerConfiguration.Instances, ControllerInstanceModel{
				IpAddr: types.StringValue(instance.Get("ipAddr").String()),
				VlanId: types.StringValue(instance.Get("vlan").String()),
			})
		}
	}
	// read associated cluster and vdses
	clVdsesMap := make(map[string][]struct {
		Id   string
		Name string
	})
	clIdNameMap := make(map[string]string)
	for _, vds := range input.Get("agent_elf_vdses").Array() {
		vdsId := vds.Get("id").String()
		vdsName := vds.Get("name").String()
		clId := vds.Get("cluster.id").String()
		clVdsesMap[clId] = append(clVdsesMap[clId], struct {
			Id   string
			Name string
		}{Id: vdsId, Name: vdsName})
	}
	for _, ac := range input.Get("agent_elf_clusters").Array() {
		acid := ac.Get("id").String()
		acname := ac.Get("name").String()
		clIdNameMap[acid] = acname
	}
	tac := make([]AssociatedClusterModel, 0)
	for _, ac := range state.AssociatedCluster {
		acid := ac.Id.ValueString()
		if ac, ok := clIdNameMap[acid]; ok {
			acm := AssociatedClusterModel{
				Id:   types.StringValue(acid),
				Name: types.StringValue(ac),
			}
			acm.VDSes = make([]AssociatedVdsModel, 0)
			if vds, ok := clVdsesMap[acid]; ok {
				for _, v := range vds {
					acm.VDSes = append(acm.VDSes, AssociatedVdsModel{
						Id:   types.StringValue(v.Id),
						Name: types.StringValue(v.Name),
					})
				}
			}
			tac = append(tac, acm)
		}
	}
	state.AssociatedCluster = tac
	if state.PackageId.IsUnknown() || state.PackageId.IsNull() {
		// set packageId to empty string when cannot find correct packageId
		// mean package may be deleted, change to other data will cause redeploy
		state.PackageId = types.StringValue("")
		cid := input.Get("controller_template.cluster").String()
		version := input.Get("version").String()
		gcp := cluster.NewGetClustersParams()
		gcp.RequestBody = &models.GetClustersRequestBody{
			Where: &models.ClusterWhereInput{
				ID: &cid,
			},
		}
		cluster_resp, err := client.Api.Cluster.GetClusters(gcp)
		if err != nil {
			diagnostic.AddError("Failed to read everoute service", fmt.Sprintf("Unable to check cluster architecture, got error: %s", err))
		} else if len(cluster_resp.Payload) == 0 {
			diagnostic.AddError("Failed to read everoute service", fmt.Sprintf("Unable to check cluster architecture, got error: %s", "Cannot find cluster"))
		} else {
			gerpp := erp.NewGetEveroutePackagesParams()
			gerpp.RequestBody = &models.GetEveroutePackagesRequestBody{
				Where: &models.EveroutePackageWhereInput{
					Arch:    cluster_resp.Payload[0].Architecture,
					Version: &version,
				},
			}
			erp_resp, err := client.Api.EveroutePackage.GetEveroutePackages(gerpp)
			if err != nil {
				diagnostic.AddError("Failed to read everoute service", fmt.Sprintf("Unable to check everoute package, got error: %s", err))
			}
			if len(erp_resp.Payload) != 0 {
				state.PackageId = types.StringValue(*erp_resp.Payload[0].ID)
			}
		}
	}
	return diagnostic
}

type ConnectIdParams struct {
	id string
}

func diffAssociatedClusters(plan *[]AssociatedClusterModel, state *[]AssociatedClusterModel) (map[string][]ConnectIdParams, map[string][]ConnectIdParams) {
	var disconnectClusters = make([]ConnectIdParams, 0)
	var connectClusters = make([]ConnectIdParams, 0)
	var setVdses = make([]ConnectIdParams, 0)
	oldClusterIdMap := make(map[string]AssociatedClusterModel)
	for _, ac := range *state {
		oldClusterIdMap[ac.Id.ValueString()] = ac
	}
	vdsIdMap := make(map[string]bool)
	for _, ac := range *plan {
		cid := ac.Id.ValueString()
		if len(ac.VDSes) > 0 {
			for _, vds := range ac.VDSes {
				vdsid := vds.Id.ValueString()
				vdsIdMap[vdsid] = true
			}
		}
		if _, ok := oldClusterIdMap[cid]; ok {
			delete(oldClusterIdMap, cid)
		} else {
			connectClusters = append(connectClusters, struct {
				id string
			}{
				id: cid,
			})
		}
		for _, ac := range oldClusterIdMap {
			// remain clusters in oldClusterIdMap are not in plan, delete
			disconnectClusters = append(disconnectClusters, struct {
				id string
			}{
				id: ac.Id.ValueString(),
			})
		}
	}
	for vdsid := range vdsIdMap {
		setVdses = append(setVdses, struct {
			id string
		}{
			id: vdsid,
		})
	}
	return map[string][]ConnectIdParams{
			"connect":    connectClusters,
			"disconnect": disconnectClusters,
		}, map[string][]ConnectIdParams{
			"set": setVdses,
		}
}
