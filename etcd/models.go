package etcd

import "net"

type Node struct {
	Name   string `json:"name"`
	NodeIp string `json:"nodeIp"`
}
type NetworkCrd struct {
	Name    string   `json:"name"`
	Subnets []Subnet `json:"subnets"`
}

type Subnet struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	CIDR string `json:"cidr"`
}

type AllocatedIp struct {
	Ip string `json:"ip"`
}
type FreeIp struct {
	Ip string `json:"ip"`
}
type PoolData struct {
	Name         string        `json:"name"`
	Id           string        `json:"id"`
	Pool         *net.IPNet    `json:"pool"`
	FreeIps      []FreeIp      `json:"free_ips"`
	AllocatedIps []AllocatedIp `json:"allocated_ips"`
}

type Pod struct {
	Name        string   `json:"name"`
	NameSpace   string   `json:"nameSpace"`
	PodEths     []PodEth `json:"podEths"`
	ContainerId string   `json:"containerId"`
}

type PodEth struct {
	NetworkCrd string    `json:"network_crd"`
	SubnetName string    `json:"subnetName"`
	Mac        string    `json:"mac"`
	FixedIps   []FixedIp `json:"fixed_ips"`
}
type FixedIp struct {
	SubnetId  string `json:"subnet_id"`
	Ipaddress string `json:"ipaddress"`
	GatewayIP string `json:"gateway_ip"`
}
