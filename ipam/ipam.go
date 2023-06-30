package ipam

import (
	"cni/client"
	"cni/etcd"
	"cni/helper"
	"github.com/containernetworking/plugins/pkg/ip"
	"k8s.io/klog/v2"
	"sync"
)

const prefix = "cni/ipam"

type Get struct {
	etcdClient  *etcd.EtcdClient
	k8sClient   *client.LightK8sClient
	nodeIpCache map[string]string
	cidrCache   map[string]string
}

type Release struct {
	etcdClient *etcd.EtcdClient
	k8sClient  *client.LightK8sClient
}

type Set struct {
	etcdClient *etcd.EtcdClient
	k8sClient  *client.LightK8sClient
}
type operators struct {
	Get     *Get
	Set     *Set
	Release *Release
}

type operator struct {
	*operators
}

type Network struct {
	Name          string
	IP            string
	Hostname      string
	CIDR          string
	IsCurrentHost bool
}

type IpamService struct {
	Subnet             string
	MaskSegment        string
	MaskIP             string
	PodMaskSegment     string
	PodMaskIP          string
	CurrentHostNetwork string
	EtcdClient         *etcd.EtcdClient
	K8sClient          *client.LightK8sClient
	*operator
}

type IPAMOptions struct {
	MaskSegment      string
	PodIpMaskSegment string
	RangeStart       string
	RangeEnd         string
}

var _lock sync.Mutex
var _isLocking bool

func unlock() {
	if _isLocking {
		_lock.Unlock()
		_isLocking = false
	}
}
func lock() {
	if !_isLocking {
		_lock.Lock()
		_isLocking = true
	}
}
func GetLightK8sClient() *client.LightK8sClient {
	paths, err := helper.GetHostAuthenticationInfoPath()
	if err != nil {
		klog.Errorf("failed to GetHostAuthenticationInfoPath,err is %v", err)
	}
	client.Init(paths.CaPath, paths.CertPath, paths.KeyPath)
	k8sClient, err := client.GetLightK8sClient()
	if err != nil {
		return nil
	}
	return k8sClient
}

func CreateNetworkCrd() {}

func AllocationIpFromNetwork(network string) (ipaddr, gw ip.IP, err error) {
	ipaddr = ip.IP{}
	gw = ip.IP{}
	return ipaddr, gw, nil
}
