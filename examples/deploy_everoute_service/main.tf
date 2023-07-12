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

data "cloudtower_cluster" "cluster" {
  name = var.cluster_config["name"]
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
  associated_cluster = []
}