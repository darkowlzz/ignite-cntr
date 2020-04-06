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
	"errors"
	"fmt"
	"path"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
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

const (
	defaultUser = "root"
)

var (
	// appEnvVars contains the environment variables to be set in the
	// application container.
	appEnvVars []string
	// netHost is the host networking option for the application container.
	netHost bool
	// envFile is a list of files containing environment variables.
	envFile []string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run <ignite-vm-name> <container-image>",
	Short: "Run a container application inside VM.",
	Long: `Run a container application inside an ignite VM. Configure the
application using flags or application run configuration file.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("require ignite VM name and container image name argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		appContainer := args[1]

		// Get all the env vars.
		envVars, err := combinedEnvVars(appEnvVars, envFile)
		if err != nil {
			fmt.Printf("error while parsing env vars: %v", err)
		}

		if err := runApp(vmName, appContainer, envVars, netHost); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func combinedEnvVars(flagEnvVars, envVarFiles []string) ([]string, error) {
	allEnvVars := flagEnvVars

	for _, envVarFile := range envVarFiles {
		envs, err := godotenv.Read(envVarFile)
		if err != nil {
			return nil, err
		}
		for k, v := range envs {
			allEnvVars = append(allEnvVars, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return allEnvVars, nil
}

func runApp(vmName string, appContainer string, envVars []string, netHost bool) error {
	if syscall.Getuid() != 0 {
		return fmt.Errorf("this command needs to be run as root")
	}

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

	ip, key, err := getIPAndPrivateKey(iclient, vmName)
	if err != nil {
		return err
	}

	appName := "container-app"

	var appSetupCmd strings.Builder

	// Create containerd container command.
	createContainer := fmt.Sprintf(`ctr -n %s container create %s %s`, containerdNamespace, appContainer, appName)
	appSetupCmd.WriteString(createContainer)

	// Set the container environment variables.
	for _, envVar := range envVars {
		env := fmt.Sprintf(" --env %s", envVar)
		appSetupCmd.WriteString(env)
	}

	// Enable host networking for the container if requested.
	if netHost {
		appSetupCmd.WriteString(" --net-host")
	}

	// appSetupCmd := fmt.Sprintf(`ctr -n ignite container create %s %s --env ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379 --env ETCD_ADVERTISE_CLIENT_URLS=http://%s:2379 --net-host`, appContainer, appName, ip)
	fmt.Println("CMD:", appSetupCmd.String())
	_, _, err = ssh.RunSSHCommand(ip, defaultUser, key, appSetupCmd.String())
	if err != nil {
		return err
	}

	// Run containerd container task.
	appRunCmd := fmt.Sprintf(`ctr -n ignite task start -d %s`, appName)
	_, _, err = ssh.RunSSHCommand(ip, defaultUser, key, appRunCmd)

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

	runCmd.Flags().StringArrayVarP(&appEnvVars, "env", "e", appEnvVars, "Set environment variables for the app container (SOME_VAR=someval)")
	runCmd.Flags().BoolVar(&netHost, "net-host", false, "Enable host networking for the container")
	runCmd.Flags().StringArrayVar(&envFile, "env-file", envFile, "Read in a file of environment variables")
}
