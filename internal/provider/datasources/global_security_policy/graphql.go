package global_security_policy

var getWhiteListDocument = `
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
	  global_whitelist {
		egress {
		  ip_block
		  except_ip_block
		  ports {
			port
			protocol
		  }
		  selector {
			id
			key
			value
		  }
		  type
		}
		enable
		ingress {
		  ip_block
		  except_ip_block
		  ports {
			port
			protocol
		  }
		  selector {
			id
			key
			value
		  }
		  type
		}
	  }
	  id
	  name
	}
  }
`
