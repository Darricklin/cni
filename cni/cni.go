package cni

import (
	"cni/skel"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	types "github.com/containernetworking/cni/pkg/types/100"
)

type IPAM struct {
	Type       string                     `json:"type"`
	Subnet     string                     `json:"subnet"`
	RangeStart string                     `json:"rangeStart"`
	RangeEnd   string                     `json:"rangeEnd"`
	GateWay    string                     `json:"gateway"`
	Addresses  []struct{ Address string } `json:"addresses"`
	Routes     interface{}                `json:"routes"`
}

type PluginConf struct {
	cniTypes.NetConf
	RuntimeConfig *struct {
		TestConfig map[string]interface{} `json:"testConfig"`
	} `json:"runtimeConfig"`
	IPAM   *IPAM  `json:"ipam"`
	Bridge string `json:"bridge"`
	Subnet string `json:"subnet"`
	Mode   string `json:"mode" default:"host-gw"`
}

type manager *CNIManager

type CNI interface {
	BootStrap(args *skel.CmdArgs, pluginConfig *PluginConf) (*types.Result, error)
}

type CNIManager struct {
	cniMap map[string]CNI
}
