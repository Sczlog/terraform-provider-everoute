部署一个 everoute 服务，不关联集群，部署于指定集群上

一个简单的 variable file 配置样例如下：
```hcl
cloudtower = {
  server   = "192.168.30.163"
  username = "username"
  password = "password"
}

er_package = {
  version      = "1.1.2"
  architecture = "X86_64"
}

cluster_config = {
  name        = "cluster-name"
  vlan_name   = "default"
  subnet_mask = "255.255.240.0"
  gateway     = "192.168.16.1"
}
```