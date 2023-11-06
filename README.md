# oWG-Mesh

**Run WireGuard VPN with zero configuration.**

This project provides a convenient way to automatically configure WireGuard. If you're running a Consul cluster and willing to configure WireGuard as VPN solution, you'll find this project very helpful.

`wg-mesh` picks an IP Address from the available pool of addresses, configures the local WireGuard interface as well as the peers and starts it. `Wg-mesh` leverages distributed locking of Consul to ensure that picked IP address is not used by any other WireGuard peer. This method is described in the [leader election](https://www.consul.io/docs/guides/leader-election.html) guide.

`WgMesh` also takes advantage of Consul blocking queries to watch nodes and KV, this allows Wg-Mesh to automatically reconfigure WireGuard Peers when nodes join or leave the Consul cluster.

Wg Mesh uses Consul KV to store WireGuard interface and Peers configurations. This makes WireGuard config distributed and available to all nodes of the cluster.

## Installation

Wg Mesh doesn't install WireGuard. It's expected to be installed and available in the `$PATH`
Wg Mesh is meant to be installed on every node of the cluster where WireGuard is need to be configured. It's better to schedule it as system daemon in all cluster nodes.

1. Clone this repo `git clone https://github.com/quarksgroup/wg-mesh.git`.
2. Cd wg-mesh
3. Make sure that you have [golang](https://go.dev/doc/install) installl and run `make build` command to make binary file.
4. Run it with `./wgmesh`.

## Configuration

Example usage:

* if-name: Network interface whose IP Address will be used for WireGuard endpoints
* wg-range: IP Address range. wg-mesh will pick address within this range
* wg-config-folder: Folder where WireGuard configurations will be stored
* wg-port: WireGuard Port(default 51820)

```
wgmesh -wg-endpoint-ip "192.168.1.10"
```

# Licence

Apache
