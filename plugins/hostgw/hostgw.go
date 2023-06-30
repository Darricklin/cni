package hostgw

import (
	"cni/cni"
	"cni/consts"
	"cni/etcd"
	"cni/ipam"
	"cni/skel"
	"cni/utils/k8s"
	"cni/utils/utils"
	"crypto/rand"
	"fmt"
	"github.com/containernetworking/cni/pkg/types"
	types100 "github.com/containernetworking/cni/pkg/types/100"
	"io"
	"k8s.io/klog"
	"os"

	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
)

const MODE = consts.MODE_HOST_GW
const (
	NETWORK     = "tinycni.io/network"
	HostVethMac = "ee:ee:ee:ee:ee:ee"
)

type HostGatewayCNI struct {
	k8sClient  *k8s.Client
	etcdClient *etcd.EtcdClient
}

func (hostgw *HostGatewayCNI) GetNetconf(ns, name string) (string, error) {
	lables, annos, err := hostgw.k8sClient.GetPodAnnoAndLabels(ns, name)
	if err != nil {
		return "", err
	}
	if networkName, ok := annos[NETWORK]; ok {
		if networkName != "" {
			return networkName, nil
		} else {
			return "", fmt.Errorf("wrong network")
		}
	} else if networkName, ok := lables[NETWORK]; ok {
		if networkName != "" {
			return networkName, nil
		} else {
			return "", fmt.Errorf("wrong network")
		}
	}
	return "", fmt.Errorf("no network find")
}
func (hostgw *HostGatewayCNI) MakeArgsMap(args string) (map[string]string, error) {
	argsMap := make(map[string]string)
	pairs := strings.Split(args, ";")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("ARGS :invilid pair %q", pair)
		}
		key := kv[0]
		value := kv[1]
		argsMap[key] = value
	}
	return argsMap, nil
}
func GeneratePortRandomMacAddress() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	macAddr := fmt.Sprintf("de:%02x:%02x:%02x:%02x:%02x", buf[1], buf[2], buf[3], buf[4], buf[5])
	return macAddr

}

func writeProcSys(path, value string) error {
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	n, err := f.Write([]byte(value))
	if err == nil && n < len(value) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

func configureSysctls(hostVethName string, hasIPv4, hasIPv6 bool) error {
	var err error

	if hasIPv4 {
		// Enable routing to localhost.  This is required to allow for NAT to the local
		// host.
		err := writeProcSys(fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/route_localnet", hostVethName), "1")
		if err != nil {
			return fmt.Errorf("failed to set net.ipv4.conf.%s.route_localnet=1: %s", hostVethName, err)
		}

		// Normally, the kernel has a delay before responding to proxy ARP but we know
		// that's not needed in a Calico network so we disable it.
		if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv4/neigh/%s/proxy_delay", hostVethName), "0"); err != nil {
			klog.Warningf("failed to set net.ipv4.neigh.%s.proxy_delay=0: %s", hostVethName, err)
		}

		// Enable proxy ARP, this makes the host respond to all ARP requests with its own
		// MAC. We install explicit routes into the containers network
		// namespace and we use a link-local address for the gateway.  Turing on proxy ARP
		// means that we don't need to assign the link local address explicitly to each
		// host side of the veth, which is one fewer thing to maintain and one fewer
		// thing we may clash over.
		if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/proxy_arp", hostVethName), "1"); err != nil {
			return fmt.Errorf("failed to set net.ipv4.conf.%s.proxy_arp=1: %s", hostVethName, err)
		}

		// Enable IP forwarding of packets coming _from_ this interface.  For packets to
		// be forwarded in both directions we need this flag to be set on the fabric-facing
		// interface too (or for the global default to be set).
		if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", hostVethName), "1"); err != nil {
			return fmt.Errorf("failed to set net.ipv4.conf.%s.forwarding=1: %s", hostVethName, err)
		}
	}

	if hasIPv6 {
		// Make sure ipv6 is enabled on the hostVeth interface in the host network namespace.
		// Interfaces won't get a link local address without this sysctl set to 0.
		if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/disable_ipv6", hostVethName), "0"); err != nil {
			return fmt.Errorf("failed to set net.ipv6.conf.%s.disable_ipv6=0: %s", hostVethName, err)
		}

		// Enable proxy NDP, similarly to proxy ARP, described above in IPv4 section.
		if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/proxy_ndp", hostVethName), "1"); err != nil {
			return fmt.Errorf("failed to set net.ipv6.conf.%s.proxy_ndp=1: %s", hostVethName, err)
		}

		// Enable IP forwarding of packets coming _from_ this interface.  For packets to
		// be forwarded in both directions we need this flag to be set on the fabric-facing
		// interface too (or for the global default to be set).
		if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/forwarding", hostVethName), "1"); err != nil {
			return fmt.Errorf("failed to set net.ipv6.conf.%s.forwarding=1: %s", hostVethName, err)
		}
	}

	if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/accept_ra", hostVethName), "0"); err != nil {
		klog.Warningf("failed to set net.ipv6.conf.%s.accept_ra=0: %s", hostVethName, err)
	}

	return nil
}

func SetupVethPair(ifName, podMac, hostVethName string, podIp, podGw ip.IP, mtu int, netNs ns.NetNS) (*types100.Interface, *types100.Interface, error) {
	hostinterface := &types100.Interface{}
	continterface := &types100.Interface{}
	// 创建vethpair，配置容器ip，默认路由，mtu
	err := netNs.Do(func(hostNs ns.NetNS) error {
		_, containerVeth, err := ip.SetupVethWithName(ifName, hostVethName, mtu, podMac, hostNs)
		if err != nil {
			return err
		}
		continterface.Name = containerVeth.Name
		continterface.Mac = containerVeth.HardwareAddr.String()
		continterface.Sandbox = netNs.Path()
		contlink, err := netlink.LinkByName(containerVeth.Name)
		if err != nil {
			return err
		}

		err = netlink.AddrAdd(contlink, &netlink.Addr{IPNet: &podIp.IPNet})
		if err != nil {
			return err
		}
		defaultNet := net.IPNet{}
		if podIp.IP.To4() != nil {
			defaultNet.IP = net.IPv4zero
		} else {
			defaultNet.IP = net.IPv6zero
		}
		if podGw.IP == nil {
			podGw.IP = net.IPv4(169, 254, 1, 1)
		}
		defaultRoute := &types.Route{Dst: defaultNet, GW: podGw.IP}
		err = ip.AddRoute(&defaultRoute.Dst, defaultRoute.GW, contlink)
		if err != nil {
			return err
		}
		if err := netlink.LinkSetMTU(contlink, mtu); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return hostinterface, continterface, err
	}

	// 配置默认的mac，mtu，路由
	hostlink, err := netlink.LinkByName(hostVethName)
	if err != nil {
		return hostinterface, continterface, err
	}
	hardwareaddr, err := net.ParseMAC(HostVethMac)
	if err != nil {
		return hostinterface, continterface, err
	}
	if err := netlink.LinkSetHardwareAddr(hostlink, hardwareaddr); err != nil {
		return hostinterface, continterface, err
	}
	if err := netlink.LinkSetMTU(hostlink, mtu); err != nil {
		return hostinterface, continterface, err
	}
	hostinterface.Name = hostVethName
	hostinterface.Mac = HostVethMac
	podIPNet := net.IPNet{}
	hasIpv4 := false
	hasIpv6 := false
	if podIp.IP.To4() != nil {
		podIPNet.IP = podIp.IP.To4()
		podIPNet.Mask = net.CIDRMask(32, 32)
		hasIpv4 = true
	} else {
		podIPNet.IP = podIp.IP.To16()
		podIPNet.Mask = net.CIDRMask(128, 128)
		hasIpv6 = true
	}
	defaultRoute := &types.Route{Dst: podIPNet, GW: podGw.IP}
	err = ip.AddRoute(&defaultRoute.Dst, defaultRoute.GW, hostlink)
	if err != nil {
		return hostinterface, continterface, err
	}
	err = configureSysctls(hostVethName, hasIpv4, hasIpv6)
	if err != nil {
		return hostinterface, continterface, err
	}
	return hostinterface, continterface, err
}

func (hostgw *HostGatewayCNI) BootStrap(args *skel.CmdArgs, conf cni.PluginConf) (*types100.Result, error) {
	//ipam.Init(conf.Subnet, nil)
	argsMap, err := hostgw.MakeArgsMap(args.Args)
	if err != nil {
		return nil, err
	}
	podNamespace := argsMap["K8S_POD_NAMESPACE"]
	podName := argsMap["K8S_POD_NAME"]
	network, err := hostgw.GetNetconf(podNamespace, podName)
	if err != nil {
		return nil, err
	}
	result := &types100.Result{
		CNIVersion: conf.CNIVersion,
	}
	podIp, gwIp, err := ipam.AllocationIpFromNetwork(network)
	if err != nil {
		klog.Error(err)
	}
	ifmac := GeneratePortRandomMacAddress()
	podNs, err := ns.GetNS(podNamespace)
	if err != nil {
		klog.Error(err)
	}
	hostVethName := "tiny" + args.ContainerID[:utils.Min(11, len(args.ContainerID))]
	hostinterface, continterface, err := SetupVethPair(args.IfName, ifmac, hostVethName, podIp, gwIp, 1500, podNs)
	if err != nil {
		klog.Error(err)
	}
	result.Interfaces = []*types100.Interface{hostinterface, continterface}
	podIpconfig := &types100.IPConfig{
		Interface: types100.Int(1),
		Address:   podIp.IPNet,
		Gateway:   gwIp.IP,
	}
	result.IPs = []*types100.IPConfig{podIpconfig}
	return result, nil
}
