package entity

import "net"

type PrimaryInterfaceInfo struct {
	Name string
	IP   net.IP
}
