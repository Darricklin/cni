package etcd

type Node struct {
	Name   string `json:"name"`
	NodeIp string `json:"nodeIp"`
}
type NetworkCrd struct {
	Name    string   `json:"name"`
	Subnets []Subnet `json:"subnets"`
}

type Subnet struct {
	Name         string `json:"name"`
	ID           string `json:"id"`
	CIDR         string `json:"cidr"`
	AllocatedIps []AllocatedIp
}

type AllocatedIp struct {
	Ip   string `json:"ip"`
	Port string `json:"port"`
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
