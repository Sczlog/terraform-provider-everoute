package everoute_service

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
		name
	  }
	  agent_elf_vdses {
		id
		name
		cluster {
			id
		}
	  }
	  controller_instances {
		ipAddr
		vlan
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
