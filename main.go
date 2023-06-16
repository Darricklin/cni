package main

import (
	"cni/skel"
	"cni/utils/log"
	"fmt"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"k8s.io/klog"
	"os"
)

func cmdAdd(args *skel.CmdArgs) error {
	klog.Infof("start cmdAdd")

	return nil
}
func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}
func main() {
	if err := log.InitLogs(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer log.FlushLogs()
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("cni"))
}
