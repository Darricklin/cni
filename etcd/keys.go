package etcd

import "fmt"

func NodesKey() string {
	return fmt.Sprintf("%s%s/", TinyCniPrefix, NodesKeyName)
}

func NodeKey(nodeName string) string {
	return fmt.Sprintf("%s%s", NodesKey(), nodeName)
}

func NetworksKey() string {
	return fmt.Sprintf("%s%s/", TinyCniPrefix, NetworksKeyName)
}

func NetworkKey(networkName string) string {
	return fmt.Sprintf("%s%s", NetworksKey(), networkName)
}

func PodsKey() string {
	return fmt.Sprintf("%s%s/", TinyCniPrefix, PodsKeyName)
}

func PodKey(nameSpace, podName string) string {
	return fmt.Sprintf("%s%s/%s", PodsKey, nameSpace, podName)
}
