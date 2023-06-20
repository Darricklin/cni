package helper

import (
	"cni/cni"
	"cni/consts"
	"cni/skel"
	"encoding/json"
	"k8s.io/klog"
)

func GetConfigs(args *skel.CmdArgs) *cni.PluginConf {
	pluginConfig := &cni.PluginConf{}
	if err := json.Unmarshal(args.StdinData, pluginConfig); err != nil {
		klog.Errorf("unmarshal err : %s", err)
		return nil
	}
	return pluginConfig
}

func GetBaseInfo(plugin *cni.PluginConf) (mode string, cniVersion string) {
	mode = plugin.Mode
	if mode == "" {
		mode = consts.MODE_HOST_GW
	}
	cniVersion = plugin.CNIVersion
	if cniVersion == "" {
		cniVersion = "0.3.0"
	}
	return mode, cniVersion
}

func TmpLogArgs(args *skel.CmdArgs) {
	klog.Infof("CmdArgs { ContainerID: %v , NetNs: %v, IfName: %v, Args: %v, Path: %v, StdinData: %v }",
		args.ContainerID, args.Netns, args.IfName, args.Args, args.Path, string(args.StdinData))
}
