package cni

import (
	"cni/skel"
	"errors"
	"fmt"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	types "github.com/containernetworking/cni/pkg/types/100"
	"k8s.io/klog"
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

var manager *CNIManager

type CNI interface {
	BootStrap(args *skel.CmdArgs, pluginConfig *PluginConf) (*types.Result, error)
	Unmount(args *skel.CmdArgs, pluginConf *PluginConf) error
	Check(args *skel.CmdArgs, pluginConf *PluginConf) error
	GetMode() string
}

type CNIManager struct {
	cniMap                map[string]CNI
	bootstrapMode         string
	unmountMode           string
	checkMode             string
	bootstrapArgs         *skel.CmdArgs
	unmountArgs           *skel.CmdArgs
	checkArgs             *skel.CmdArgs
	bootstrapPluginConfig *PluginConf
	unmountPluginConfig   *PluginConf
	checkPluginConfig     *PluginConf
	result                *types.Result
}

func (manager *CNIManager) getCNI(mode string) CNI {
	if cni, ok := manager.cniMap[mode]; ok {
		return cni
	}
	return nil
}

func (manager *CNIManager) getBootstrapMode() string {
	return manager.bootstrapMode
}

func (manager *CNIManager) getUnmountMode() string {
	return manager.unmountMode
}

func (manager *CNIManager) getCheckMode() string {
	return manager.checkMode
}

func (manager *CNIManager) getBootStrapArgs() *skel.CmdArgs {
	return manager.bootstrapArgs
}

func (manager *CNIManager) getUnmountArgs() *skel.CmdArgs {
	return manager.unmountArgs
}

func (manager *CNIManager) getCheckArgs() *skel.CmdArgs {
	return manager.checkArgs
}

func (manager *CNIManager) getBootstrapConfigs() *PluginConf {
	return manager.bootstrapPluginConfig
}

func (manager *CNIManager) getUnmountConfigs() *PluginConf {
	return manager.unmountPluginConfig
}

func (manager *CNIManager) getCheckConfigs() *PluginConf {
	return manager.checkPluginConfig
}

func (manager *CNIManager) getResult() *types.Result {
	return manager.result
}

func (manager *CNIManager) Register(cni CNI) error {
	mode := cni.GetMode()
	if mode == "" {
		return errors.New("cni mode cannot be null")
	}
	_cni := manager.getCNI(mode)
	if _cni != nil {
		return errors.New("the cni has exists")
	}
	manager.cniMap[mode] = cni
	return nil
}

func (manager *CNIManager) SetBootstrapConfigs(conf *PluginConf) *CNIManager {
	manager.bootstrapPluginConfig = conf
	return manager
}

func (manager *CNIManager) SetUnmountConfigs(conf *PluginConf) *CNIManager {
	manager.unmountPluginConfig = conf
	return manager
}

func (manager *CNIManager) SetCheckConfigs(conf *PluginConf) *CNIManager {
	manager.checkPluginConfig = conf
	return manager
}

func (manager *CNIManager) SetBootStrapArgs(args *skel.CmdArgs) *CNIManager {
	manager.bootstrapArgs = args
	return manager
}

func (manager *CNIManager) SetUnmountArgs(args *skel.CmdArgs) *CNIManager {
	manager.unmountArgs = args
	return manager
}

func (manager *CNIManager) SetCheckArgs(args *skel.CmdArgs) *CNIManager {
	manager.checkArgs = args
	return manager
}

func (manager *CNIManager) SetBootStrapCNIMode(mode string) *CNIManager {
	manager.bootstrapMode = mode
	return manager
}
func (manager *CNIManager) SetUnmountCNIMode(mode string) *CNIManager {
	manager.unmountMode = mode
	return manager
}
func (manager *CNIManager) SetCheckMode(mode string) *CNIManager {
	manager.checkMode = mode
	return manager
}

func (manager *CNIManager) BootStrapCNI() error {
	mode := manager.getBootstrapMode()
	args := manager.getBootStrapArgs()
	configs := manager.getBootstrapConfigs()
	if mode == "" || args == nil || configs == nil {
		return errors.New("start cni need set mode,args and configs")
	}
	cni := manager.getCNI(mode)
	if cni == nil {
		errMsg := fmt.Sprintf("cannot find %s type cni ", mode)
		return errors.New(errMsg)
	}
	cniRes, err := cni.BootStrap(args, configs)
	if err != nil {
		klog.Errorf("wrong at BootStrapCNI ,err is %s", err)
		return err
	}
	manager.result = cniRes
	return nil
}

func (manager *CNIManager) UnmountCNI() error {
	mode := manager.getBootstrapMode()
	args := manager.getBootStrapArgs()
	configs := manager.getBootstrapConfigs()
	if mode == "" || args == nil || configs == nil {
		return errors.New("unmount cni need set mode,args and configs")
	}
	cni := manager.getCNI(mode)
	if cni == nil {
		errMsg := fmt.Sprintf(" %s cni has not been init,cannot uninstall ", mode)
		return errors.New(errMsg)
	}
	return cni.Unmount(args, configs)
}

func (manager *CNIManager) CheckCNI() error {
	mode := manager.getBootstrapMode()
	args := manager.getBootStrapArgs()
	configs := manager.getBootstrapConfigs()
	if mode == "" || args == nil || configs == nil {
		return errors.New("unmount cni need set mode,args and configs")
	}
	cni := manager.getCNI(mode)
	if cni == nil {
		errMsg := fmt.Sprintf(" %s cni has not been init,cannot check ", mode)
		return errors.New(errMsg)
	}
	return cni.Check(args, configs)
}

func (manager *CNIManager) PrintResult() error {
	result := manager.getResult()
	if result == nil {
		return errors.New("PrintResult cannot get result of cni exec")
	}
	config := manager.getBootstrapConfigs()
	if config == nil {
		return errors.New("PrintResult cannot get result of cni configs")
	}
	version := config.CNIVersion
	if version == "" {
		return errors.New("PrintResult cannot get result of cni version")
	}
	cniTypes.PrintResult(result, version)
	return nil
}

func GetCNIManager() *CNIManager {
	return manager
}

func init() {
	manager = &CNIManager{
		cniMap: map[string]CNI{},
	}
}
