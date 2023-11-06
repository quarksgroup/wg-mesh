package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/quarksgroup/wg-mesh/wireguard"
)

type Consul struct {
	client   *api.Client
	lock     *api.Lock
	lockChan <-chan struct{}
}

type ConsulService interface {
	//location can be a key for K/V stores or table for databases,....
	Lock(location string, value string) error
	//unlock the lock
	Unlock()
	//get all peers
	GetPeers(location string) ([]wireguard.Peer, error)
	// add a peer
	AddPeer(location string, wgInterface wireguard.Interface, peer wireguard.Peer) error
	// monitor the K/V store for changes
	MonitorKv(location string, wgInterface wireguard.Interface)
	// monitor the nodes for changes
	MonitorNodes(location string, wgInterface wireguard.Interface)
}
