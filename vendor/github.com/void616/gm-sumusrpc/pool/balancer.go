package pool

import (
	"fmt"
)

// DefaultBalancer is default balancer
type DefaultBalancer struct {
	Balancer
}

// Get picks a node from the pool of nodes
func (b *DefaultBalancer) Get(nodes map[string]*NodePool) (*NodePool, error) {
	at := ""
	weight := int64((^uint64(0)) >> 1)
	for i, p := range nodes {
		if p.Available() {
			w := int64(p.ConsumedConnections()) + int64(p.PendingConsumers())
			if weight > w {
				weight = w
				at = i
			}
		}
	}

	if at == "" {
		return nil, fmt.Errorf("failed to find a free node")
	}

	return nodes[at], nil
}
