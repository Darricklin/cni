package log

import (
	"flag"
	"k8s.io/klog"
)

func InitLogs() error {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	return nil
}
func FlushLogs() {
	klog.Flush()
}
