package main

import (
	"cni/skel"
	"cni/utils/log"
	"fmt"
	"os"
)

func cmdAdd(args *skel.CmdArgs) error {

}
func cmdDel(args *skel.CmdArgs) error {

}

func cmdCheck(args *skel.CmdArgs) error {

}
func main() {
	if err := log.InitLogs(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer log.FlushLogs()
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.buildString("cni"))
}
