terraform {
  # everoute provider is recommaned to used with cloudtower provider
  required_providers {
    everoute = {
      source = "registry.terraform.io/smartxworks/everoute"
    }
    cloudtower = {
      source = "registry.terraform.io/smartxworks/cloudtower"
    }
  }
}

provider "everoute" {
  cloudtower_server = var.cloudtower["server"]
  username          = var.cloudtower["username"]
  password          = var.cloudtower["password"]
}

provider "cloudtower" {
  cloudtower_server = var.cloudtower["server"]
  username          = var.cloudtower["username"]
  password          = var.cloudtower["password"]
}

data "everoute_package" "package" {
  version      = var.er_package["version"]
  architecture = var.er_package["architecture"]
}

locals {
  associated_clusters_name = [
    for cluster in var.associated_cluster_config :
    cluster["name"]
  ]
}

data "cloudtower_cluster" "cluster" {
  name = var.cluster_config["name"]
}

data "cloudtower_cluster" "associated_clusters" {
  name_in = local.associated_clusters_name
}

locals {
  assocaited_cluster_map = {
    for cluster in data.cloudtower_cluster.associated_clusters.clusters :
    cluster.name => {
      for k, v in cluster : k => v
    }
  }
}


data "cloudtower_vlan" "vlan" {
  name       = var.cluster_config["vlan_name"]
  cluster_id = data.cloudtower_cluster.cluster.clusters[0].id
  type       = "VM"
}

locals {
  controller_instances = [
    for ip in var.cluster_config["ips"] :
    {
      ip_addr = ip
      vlan_id = data.cloudtower_vlan.vlan.vlans[0].id
    }
  ]

  associated_clusters = [
    for cluster_config in var.associated_cluster_config :
    {
      id = local.assocaited_cluster_map[cluster_config["name"]].id,
      vdses = [for vdsid in cluster_config["vdses"] :
        {
          id = vdsid
      }]
    }
  ]
}

resource "everoute_service" "service" {
  name       = "test"
  package_id = data.everoute_package.package.packages[0].id
  controller_configuration = {
    cluster_id  = data.cloudtower_cluster.cluster.clusters[0].id
    subnet_mask = var.cluster_config["subnet_mask"]
    gateway     = var.cluster_config["gateway"]
    instance    = local.controller_instances
  }
  associated_cluster = local.associated_clusters
}