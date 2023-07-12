package everoute_service

var duplicatedNameServiceDocument = `
query everouteClusters(
	$after: String
	$before: String
	$first: Int
	$last: Int
	$orderBy: EverouteClusterOrderByInput
	$skip: Int
	$where: EverouteClusterWhereInput
  ) {
	everouteClusters(
	  after: $after
	  before: $before
	  first: $first
	  last: $last
	  orderBy: $orderBy
	  skip: $skip
	  where: $where
	) {
		name
	}
  }
`

var deployEverouteServiceDocument = `
mutation deployEverouteCluster(
	$data: EverouteClusterCreateInput!
	$effect: CreateEverouteClusterEffectInput!
  ) {
	createEverouteCluster(data: $data, effect: $effect) {
	  id
	  name
	}
  }
`

var associatedClusterDocument = `
mutation updateEverouteClusterAssociation(
	$where: EverouteClusterWhereUniqueInput!
	$data: EverouteClusterUpdateInput!
  ) {
	updateEverouteCluster(where: $where, data: $data) {
	  id
	}
  }
`

var getEverouteServiceDocument = `
query everouteClusters(
	$after: String
	$before: String
	$first: Int
	$last: Int
	$orderBy: EverouteClusterOrderByInput
	$skip: Int
	$where: EverouteClusterWhereInput
  ) {
	everouteClusters(
	  after: $after
	  before: $before
	  first: $first
	  last: $last
	  orderBy: $orderBy
	  skip: $skip
	  where: $where
	) {
	  agent_elf_clusters {
		id
	  }
	  agent_elf_vdses {
		id
		name
		cluster {
			id
			name
		}
	  }
	  controller_instances {
		ipAddr
		vlan
	  }
	  controller_template {
		cluster
		gateway
		netmask
	  }
	  global_default_action
	  global_whitelist {
		enable
	  }
	  id
	  installed
	  name
	  phase
	  version
	}
  }
`

var deleteEverouteServiceDocument = `
mutation deleteEverouteCluster($where: EverouteClusterWhereUniqueInput!) {
	deleteEverouteCluster(where: $where) {
	  id
	}
  }
`
