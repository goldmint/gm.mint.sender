package pool

import (
	"fmt"
)

// DefaultBalancer is default balancer
type DefaultBalancer struct {
	Balancer
}

// Get picks a node from the pool of nodes
func (b *DefaultBalancer) Get(nodes []NodeMeta) (NodeMeta, error) {
	var at = -1
	var weight = ^uint32(0)
	for i, p := range nodes {
		if p.Available {
			w := uint32(p.BusyConnections) + uint32(p.PendingConsumers)
			if weight > w {
				weight = w
				at = i
			}
		}
	}
	if at < 0 {
		return NodeMeta{}, fmt.Errorf("failed to find a free node")
	}
	return nodes[at], nil
}
