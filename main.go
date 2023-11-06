package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"

	"github.com/quarksgroup/wg-mesh/consul"
	"github.com/quarksgroup/wg-mesh/util"
	"github.com/quarksgroup/wg-mesh/wireguard"
)

func main() {

	flag.Parse()
	if err := initConfig(); err != nil {
		log.Fatal("FATAL: ", err.Error())
	}

	// Backend init
	srv, err := ConsulProvide()
	if err != nil {
		log.Fatal("FATAL: ", err)
	}

	newWgInterface, err := initialize(srv)
	if err != nil {
		log.Fatal("FATAL: ", err)
	}

	srv.MonitorKv(config.CS_FullKVPrefix, *newWgInterface)
	if config.GC_Enable {
		srv.MonitorNodes(config.CS_FullKVPrefix, *newWgInterface)
	}

}

func initialize(cos consul.ConsulService) (*wireguard.Interface, error) {

	if err := cos.Lock(config.CS_FullKVPrefix, config.WG_EndpointIP); err != nil {
		return nil, err
	}
	defer cos.Unlock()

	var shouldConfigureWG bool = false

	peers, err := cos.GetPeers(config.CS_FullKVPrefix)
	if err != nil {
		return nil, err
	}

	if config.WG_IP != "" { //WG_IP is set, this is like a pet server
		if indexPeer := util.IsIPUsed(peers, config.WG_IP); indexPeer >= 0 { //WG_IP is already used
			if !peers[indexPeer].IsEndpointIPDefined() { //WG_IP is used by another device
				return nil, fmt.Errorf("the WG_IP=%s is already used by another client", config.WG_IP)
			} else {
				if !peers[indexPeer].HasSameEndpointIP(config.WG_EndpointIP) { //WG_IP is used by an other server
					return nil, fmt.Errorf("the WG_IP=%s is already used by another server %s", peers[indexPeer].GetIP(), peers[indexPeer].GetEndpointIP())
				} else { //WG_IP is used by the current server
					shouldConfigureWG = true
				}
			}
		} else {
			//The WG_IP is not used
			if indexPeer := util.IsEndpointIPExist(peers, config.WG_EndpointIP); indexPeer >= 0 { //WG_EndpointIP already registred
				return nil, fmt.Errorf("the current WG_EndpointIP=%s is already registred with another WG_IP=%s", config.WG_EndpointIP, peers[indexPeer].GetIP())
			} else {
				//WG_EndpointIP is not registred
				shouldConfigureWG = true
			}
		}
	} else {
		//WG_IP is not set, this is like a cattle server
		if indexPeer := util.IsEndpointIPExist(peers, config.WG_EndpointIP); indexPeer >= 0 && peers[indexPeer].IP != nil { //WG_EndpointIP already registred
			config.WG_IP = peers[indexPeer].IP.String()
			shouldConfigureWG = true
		} else { //pick unused IP
			log.Print("INFO: Picking unused IP from the range=", config.WG_Range)
			_, wgRangeIPNet, err := net.ParseCIDR(config.WG_Range)
			if err != nil {
				return nil, err
			}
			wgIPStart := wgRangeIPNet.IP
			util.IncIP(wgIPStart) //Skip IP Network

			//The loop goes over all ips in the network
			for myFutureWG_IP := wgIPStart; wgRangeIPNet.Contains(myFutureWG_IP); util.IncIP(myFutureWG_IP) {
				if indexPeer := util.IsIPUsed(peers, myFutureWG_IP.String()); indexPeer < 0 {
					config.WG_IP = myFutureWG_IP.String()
					break
				}
			}
			if config.WG_IP == "" {
				return nil, fmt.Errorf("all IPs are used")
			} else {
				shouldConfigureWG = true
			}

		}
	}

	wgConfig := wireguard.Configuration{}

	if shouldConfigureWG {
		// configure wireguard
		privKey, pubKey, err := wireguard.InitWgKeys(config.WG_InterfaceConfigFolder)
		if err != nil {
			return nil, err
		}

		var mask string
		if config.WG_Relay {
			mask = strings.Split(config.WG_Range, "/")[1]
		} else {
			mask = "32"
		}

		wgConfig = wireguard.Configuration{
			Interface: wireguard.Interface{
				Name:       config.WG_InterfaceName,
				Address:    fmt.Sprintf("%s/%s", config.WG_IP, mask),
				ListenPort: config.WG_Port,
				PublicKey:  pubKey,
				PrivateKey: privKey,
				PostUp:     config.WG_PostUp,
				PostDown:   config.WG_PostDown,
			},
			Peers: peers,
		}

		//wirete the wireguard config file
		if _, err := wireguard.ConfigureWireguard(wgConfig); err != nil {
			return nil, err
		}

		allowedips := []string{}
		if config.WG_AllowedIPs != "" {
			allowedips = strings.Split(config.WG_AllowedIPs, ",")
		}

		allowedips = append(allowedips, fmt.Sprintf("%s/%s", config.WG_IP, "32"))
		if config.WG_Relay {
			_, rangeIPNet, err := net.ParseCIDR(config.WG_Range)
			if err != nil {
				return nil, err
			}
			allowedips = append(allowedips, rangeIPNet.String())
		}

		sort.Strings(allowedips)

		newPeer := wireguard.Peer{
			PublicKey:    pubKey,
			IP:           net.ParseIP(config.WG_IP),
			EndpointIP:   net.ParseIP(config.WG_EndpointIP),
			EndpointPort: config.WG_Port,
			AllowedIPs:   strings.Join(allowedips, ","),
		}

		//Add my interface as Peer in the Backend
		if err := cos.AddPeer(config.CS_FullKVPrefix, wgConfig.Interface, newPeer); err != nil {
			return nil, err
		}

	}

	return &wgConfig.Interface, nil
}

func ConsulProvide() (consul.ConsulService, error) {

	if config.Driver == "consul" {
		b, err := consul.New(config.Consul_Addr)
		if err != nil {
			return nil, err
		}
		return b, nil
	}

	return nil, fmt.Errorf("unsupported Backend. Available backends: [consul]")

}
