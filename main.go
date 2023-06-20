package main

import (
	"cni/cni"
	"cni/helper"
	"cni/skel"
	"cni/utils/log"
	"errors"
	"fmt"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"k8s.io/klog"
	"os"
)

func cmdAdd(args *skel.CmdArgs) error {
	klog.Infof("start cmdAdd")
	helper.TmpLogArgs(args)
	pluginConfig := helper.GetConfigs(args)
	if pluginConfig == nil {
		errMsg := fmt.Sprintf("Cmd ADD：failed to get plugin config , conifg is %s ", string(args.StdinData))
		klog.Errorf("%s", errMsg)
		return errors.New(errMsg)
	}
	mode, cniVersion := helper.GetBaseInfo(pluginConfig)
	if pluginConfig.CNIVersion == "" {
		pluginConfig.CNIVersion = cniVersion
	}
	cniManager := cni.GetCNIManager().SetBootstrapConfigs(pluginConfig).SetBootStrapArgs(args).SetBootStrapCNIMode(mode)
	if cniManager == nil {
		klog.Errorf("Cmd ADD： cni init failed ")
		return errors.New("Cmd ADD: cni init failed ")
	}
	err := cniManager.BootStrapCNI()
	if err != nil {
		klog.Errorf("setup cni failed: %s", err)
		return err
	}
	err = cniManager.PrintResult()
	if err != nil {
		klog.Errorf("failed print result :%v", err)
		return err
	}
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	klog.Infof("start CmdDel")
	helper.TmpLogArgs(args)
	pluginConfig := helper.GetConfigs(args)
	if pluginConfig == nil {
		errMsg := fmt.Sprintf("del : failed to get plugin config ,config: %s", string(args.StdinData))
		klog.Errorf("%s", errMsg)
		return errors.New(errMsg)
	}
	mode, _ := helper.GetBaseInfo(pluginConfig)
	cniManager := cni.GetCNIManager().SetUnmountConfigs(pluginConfig).SetUnmountArgs(args).SetUnmountCNIMode(mode)
	return cniManager.UnmountCNI()
}

func cmdCheck(args *skel.CmdArgs) error {
	klog.Infof("start CmdCheck")
	helper.TmpLogArgs(args)
	pluginConfig := helper.GetConfigs(args)
	if pluginConfig == nil {
		errMsg := fmt.Sprintf("check : failed to get plugin config ,config: %s", string(args.StdinData))
		klog.Errorf("%s", errMsg)
		return errors.New(errMsg)
	}
	mode, _ := helper.GetBaseInfo(pluginConfig)
	cniManager := cni.GetCNIManager().SetCheckConfigs(pluginConfig).SetCheckArgs(args).SetCheckMode(mode)
	return cniManager.CheckCNI()
}
func main() {
	if err := log.InitLogs(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer log.FlushLogs()
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("cni"))
}
