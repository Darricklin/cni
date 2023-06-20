package hostgw

import (
	"cni/cni"
	"cni/consts"
	"cni/skel"
)

const MODE = consts.MODE_HOST_GW

type HostGatewayCNI struct {
}

func (hostgw *HostGatewayCNI) BootStrap(args *skel.CmdArgs, conf cni.PluginConf) (*types.Result, error) {
	ipam.Init(conf.Subnet, nil)

}
