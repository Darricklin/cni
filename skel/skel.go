package skel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/utils"
	"github.com/containernetworking/cni/pkg/version"
	"io"
	"io/ioutil"
	"k8s.io/klog"
	"strings"
)

type CmdArgs struct {
	ContainerID string
	Netns       string
	IfName      string
	Args        string
	Path        string
	StdinData   []byte
}

type dispatcher struct {
	Getenv             func(string) string
	Stdin              io.Reader
	Stdout             io.Writer
	Stderr             io.Writer
	ConfVersionDecoder version.ConfigDecoder
	VersionReconciler  version.Reconciler
}

type reqFromCmdEntry map[string]bool

func (t *dispatcher) getCmdArgsFromEnv() (string, *CmdArgs, *types.Error) {
	var cmd, contID, netns, ifName, args, path string
	vars := []struct {
		name       string
		val        *string
		reqFromCmd reqFromCmdEntry
	}{
		{"CNI_COMMAND",
			&cmd, reqFromCmdEntry{"ADD": true, "CHECK": true, "DEL": true}},
		{"CNI_CONTAINERID",
			&contID, reqFromCmdEntry{"ADD": true, "CHECK": true, "DEL": true}},
		{"CNI_NETNS",
			&netns, reqFromCmdEntry{"ADD": true, "CHECK": true, "DEL": false}},
		{"CNI_IFNAME",
			&ifName, reqFromCmdEntry{"ADD": true, "CHECK": true, "DEL": true}},
		{"CNI_ARGS",
			&args, reqFromCmdEntry{"ADD": false, "CHECK": false, "DEL": false}},
		{"CNI_PATH",
			&path, reqFromCmdEntry{"ADD": true, "CHECK": true, "DEL": true}},
	}
	argsMissing := make([]string, 0)
	for _, v := range vars {
		*v.val = t.Getenv(v.name)
		if *v.val == "" {
			if v.reqFromCmd[cmd] || v.name == "CNI_COMMAND" {
				argsMissing = append(argsMissing, v.name)
			}
		}
	}
	if len(argsMissing) > 0 {
		joined := strings.Join(argsMissing, ",")
		return "", nil, types.NewError(types.ErrInvalidEnvironmentVariables, fmt.Sprintf("required env variables [%s] missing", joined), "")
	}
	if cmd == "VERSION" {
		t.Stdin = bytes.NewReader(nil)
	}
	stdinData, err := ioutil.ReadAll(t.Stdin)
	if err != nil {
		return "", nil, types.NewError(types.ErrIOFailure, fmt.Sprintf("error reading from stdin : %v", err), "")
	}
	cmdArgs := &CmdArgs{
		ContainerID: contID,
		Netns:       netns,
		IfName:      ifName,
		Args:        args,
		Path:        path,
		StdinData:   stdinData,
	}
	return cmd, cmdArgs, nil
}

func (t *dispatcher) checkVersionAndCall(args *CmdArgs, pluginVersionInfo version.PluginInfo, toCall func(cmdArgs *CmdArgs) error) *types.Error {
	configVersion, err := t.ConfVersionDecoder.Decode(args.StdinData)
	if err != nil {
		return types.NewError(types.ErrDecodingFailure, err.Error(), "")
	}
	verErr := t.VersionReconciler.Check(configVersion, pluginVersionInfo)
	if verErr != nil {
		return types.NewError(types.ErrIncompatibleCNIVersion, "incompatible CNI version", verErr.Details())
	}
	if err = toCall(args); err != nil {
		if e, ok := err.(*types.Error); ok {
			return e
		}
		return types.NewError(types.ErrInternal, err.Error(), "")
	}
	return nil
}

func validateConfig(jsonBytes []byte) *types.Error {
	var conf struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(jsonBytes, &conf); err != nil {
		return types.NewError(types.ErrDecodingFailure, fmt.Sprintf("err unmarshall network config"), "")
	}
	if conf.Name == "" {
		return types.NewError(types.ErrInvalidNetworkConfig, "missing network name", "")
	}
	if err := utils.ValidateNetworkName(conf.Name); err != nil {
		return err
	}
	return nil
}

func (t *dispatcher) pluginMain(cmdAdd, cmdCheck, cmdDel func(_ *CmdArgs) error, versionInfo version.PluginInfo, about string) *types.Error {
	cmd, cmdArgs, err := t.getCmdArgsFromEnv()
	if err != nil {
		if err.Code == types.ErrInvalidEnvironmentVariables && t.Getenv("CNI_COMMAND") == "" && about != "" {
			_, _ = fmt.Fprintf(t.Stderr, about)
			return nil
		}
		return err
	}
	klog.Infof("enter pluginMain ,cmd is %+v", cmd)
	if cmd != "VERSION" {
		if err = validateConfig(cmdArgs.StdinData); err != nil {
			return err
		}
		if err = utils.ValidateContainerID(cmdArgs.ContainerID); err != nil {
			return err
		}
		if err = utils.ValidateInterfaceName(cmdArgs.IfName); err != nil {
			return err
		}
	}
}
