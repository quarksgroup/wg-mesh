package ifconfig

import (
	"fmt"
	"net"
)

func GetIpOfIf(interfaceName string) (string, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	if len(ifaces) <= 0 {
		return "", fmt.Errorf("no network interface found")
	}

	for _, i := range ifaces {
		if i.Name == interfaceName {
			addrs, err := i.Addrs()
			if err != nil {
				return "", err
			}
			if len(addrs) > 0 {
				return addrs[0].String(), err
			}
		}
	}
	return "", fmt.Errorf("the network interface %s doesn't exist or doesn't have any IP address", interfaceName)
}

func IsInterfaceStarted(interfaceName string) (bool, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false, err
	}

	if len(ifaces) <= 0 {
		return false, nil
	}

	for _, i := range ifaces {
		if i.Name == interfaceName {
			return true, nil
		}
	}
	return false, nil
}
