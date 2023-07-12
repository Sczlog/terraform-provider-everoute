package global_security_policy

var getGlobalWhitelistDocument = `
query getEverouteClusters($where: EverouteClusterWhereInput) {
	everouteClusters(where: $where, first: 1) {
	  id
	  global_default_action
	  global_whitelist {
		enable
		egress {
		  ip_block
		  except_ip_block
		  ports {
			port
			protocol
		  }
		}
		ingress {
		  ip_block
		  except_ip_block
		  ports {
			port
			protocol
		  }
		}
	  }
	}
  }
  
`

var updateGlobalWhiteListDocument = `
mutation updateEverouteClusterGlobalAction(
	$where: EverouteClusterWhereUniqueInput!
	$data: EverouteClusterUpdateInput!
  ) {
	updateEverouteCluster(where: $where, data: $data) {
	  id
	  global_default_action
	  __typename
	}
  }
`
