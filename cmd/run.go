/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/weaveworks/gitops-toolkit/pkg/filter"
	api "github.com/weaveworks/ignite/pkg/apis/ignite"
	"github.com/weaveworks/ignite/pkg/client"
	"github.com/weaveworks/ignite/pkg/constants"
	"github.com/weaveworks/ignite/pkg/network"
	"github.com/weaveworks/ignite/pkg/providers"
	providersIgnite "github.com/weaveworks/ignite/pkg/providers/ignite"
	"github.com/weaveworks/ignite/pkg/runtime"

	"github.com/darkowlzz/ignite-cntr/ssh"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("run called")
		if err := runApp(); err != nil {
			fmt.Printf("fail: %v", err)
		}
	},
}

func runApp() error {
	// Set default runtime and network plugin.
	providers.RuntimeName = runtime.RuntimeDocker
	providers.NetworkPluginName = network.PluginDockerBridge

	// Initialize ignite.
	if err := providers.Populate(providersIgnite.Preload); err != nil {
		return fmt.Errorf("failed to initialize ignite preload: %w", err)
	}
	if err := providers.Populate(providersIgnite.Providers); err != nil {
		return fmt.Errorf("failed to initialize ignite providers: %w", err)
	}

	iclient := providers.Client.VMs()

	ip, key, err := getIPAndPrivateKey(iclient, "my-vm")
	if err != nil {
		return err
	}

	cmds := fmt.Sprintf(`ctr -n ignite container create quay.io/coreos/etcd:v3.4.7 etcd --env ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 --env ETCD_ADVERTISE_CLIENT_URLS=http://%s:2379 --net-host
ctr -n ignite task start -d etcd`, ip)
	fmt.Println("CMD:", cmds)
	_, _, err = ssh.RunSSHCommand(ip, "root", key, cmds)
	if err != nil {
		return err
	}

	return nil
}

// getIPAndPrivateKey gets the IP and private key file path of a given machine.
func getIPAndPrivateKey(iclient client.VMClient, name string) (string, string, error) {
	vm, err := getVMByName(iclient, name)
	if err != nil {
		return "", "", err
	}

	if !vm.Running() {
		return "", "", fmt.Errorf("failed to get IP, VM %q is not running", vm.Name)
	}

	ipAddrs := vm.Status.IPAddresses
	if len(ipAddrs) == 0 {
		return "", "", fmt.Errorf("failed to get IP, VM %q has no usable IP addresses", vm.Name)
	}

	privKeyFile := path.Join(vm.ObjectPath(), fmt.Sprintf(constants.VM_SSH_KEY_TEMPLATE, vm.GetUID()))

	return ipAddrs[0].String(), privKeyFile, nil
}

func getVMByName(iclient client.VMClient, name string) (*api.VM, error) {
	return iclient.Find(filter.NewIDNameFilter(name))
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
