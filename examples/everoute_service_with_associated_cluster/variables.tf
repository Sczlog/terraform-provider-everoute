variable "cloudtower" {
  type = object({
    server   = string
    username = string
    password = string
  })
}

variable "er_package" {
  type = object({
    version      = string
    architecture = string
  })
}

variable "cluster_config" {
  type = object({
    name        = string
    vlan_name   = string
    subnet_mask = string
    gateway     = string
    ips         = list(string)
  })
}

variable "associated_cluster_config" {
  type = list(object({
    name  = string
    vdses = list(string)
  }))
}
