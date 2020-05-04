package cmd

import (
	"errors"
	"fmt"
	"math/rand"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/weaveworks/gitops-toolkit/pkg/filter"
	igniteRun "github.com/weaveworks/ignite/cmd/ignite/run"
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
	// Any mounts are copied into this directory of the VM and then mounted into
	// the application container.
	defaultMountParentDir = "/tmp"
)

var (
	// appEnvVars contains the environment variables to be set in the
	// application container.
	appEnvVars []string
	// netHost is the host networking option for the application container.
	netHost bool
	// envFile is a list of files containing environment variables.
	envFile []string
	// appCmd is the command that's passed to a container to run.
	appCmd string
	// appCmdArgs is the arguments to the command passed to a container. A
	// command must be set for passing the args, else the args will be ignored.
	appCmdArgs []string
	// mountSrcPath is the path of the source that needs to be mounted.
	mountSrcPath string
	// mountDestPath is the mount point in the application container.
	mountDestPath string
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
		appImage := args[1]

		// Get all the env vars.
		envVars, err := combinedEnvVars(appEnvVars, envFile)
		if err != nil {
			fmt.Printf("error while parsing env vars: %v", err)
		}

		if err := runApp(vmName, appImage, appCmd, appCmdArgs, envVars, netHost); err != nil {
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

func runApp(vmName string, appImage string, appcmd string, appCmdArgs []string, envVars []string, netHost bool) error {
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

	// containerMountSet is used to check if the mount flags must be set when
	// creating a container.
	containerMountSet := false

	// TODO: Add support for multiple mounts.
	if mountSrcPath != "" {
		// Ensure mount destination path is also passed.
		if mountDestPath == "" {
			return fmt.Errorf("when mounting, both --mount-src and --mount-dest must be set")
		}

		if err := copyFileToVM(vmName, mountSrcPath); err != nil {
			return fmt.Errorf("failed to copy mount-src %q into the VM", mountSrcPath)
		}
		// Set the containerMountSet to true after a successful copy of files to
		// the VM.
		containerMountSet = true
	}

	// Generate a random container app name.
	rand.Seed(time.Now().UnixNano())
	appName := fmt.Sprintf("container-app-%d", rand.Int())

	var appSetupCmd strings.Builder

	// NOTE: The order of arguments to ctr container create is important.
	// <image> <container-name> [command] [args] [--env|--net-host|...]

	// Create containerd container command.
	//   ctr -n <namespace> container create <image> <container-name>
	createContainer := fmt.Sprintf(`%s -n %s container create %s %s`, ctrPath, containerdNamespace, appImage, appName)
	appSetupCmd.WriteString(createContainer)

	// Set the container command if provided.
	if appCmd != "" {
		appSetupCmd.WriteString(fmt.Sprintf(" %s", appCmd))

		// Pass the command arguments.
		for _, arg := range appCmdArgs {
			appSetupCmd.WriteString(fmt.Sprintf(" %s", arg))
		}
	}

	// Set the container environment variables.
	for _, envVar := range envVars {
		env := fmt.Sprintf(" --env %s", envVar)
		appSetupCmd.WriteString(env)
	}

	// Enable host networking for the container if requested.
	if netHost {
		appSetupCmd.WriteString(" --net-host")
	}

	// Set mount flags. Source is the path in the VM, destination is the path
	// in the application container.
	if containerMountSet {
		mntSrc := filepath.Join(defaultMountParentDir, filepath.Base(mountSrcPath))
		mountFlag := fmt.Sprintf(" --mount=\"src=%s,dst=%s,type=bind,options=rbind:ro\"", mntSrc, mountDestPath)
		appSetupCmd.WriteString(mountFlag)
	}

	fmt.Printf("Creating container %s...\n", appName)
	fmt.Println("CMD:", appSetupCmd.String())
	if err = runCmdInVM(ip, key, appSetupCmd.String()); err != nil {
		return err
	}

	fmt.Printf("Running task %s...\n", appName)
	// Run containerd container task.
	appRunCmd := fmt.Sprintf(`ctr -n ignite task start -d %s`, appName)
	return runCmdInVM(ip, key, appRunCmd)
}

// runCmdInVM takes a VM IP, ssh key and runs the given command in the VM.
func runCmdInVM(ip, key, cmd string) error {
	cmdOut, cmdErr, err := ssh.RunSSHCommand(ip, defaultUser, key, cmd)
	if err != nil {
		return err
	}

	// Print the command stdout and stderr.
	fmt.Printf("[STDOUT]:\n%s\n", cmdOut)

	if len(cmdErr) > 0 {
		fmt.Printf("[STDERR]:\n%s\n", cmdErr)
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

// copyFileToVM copies a file into a given VM.
func copyFileToVM(vmName, source string) error {
	// Construct destination path: <vm-name>:<path-in-vm>
	destPath := filepath.Join(defaultMountParentDir, filepath.Base(source))
	dest := fmt.Sprintf("%s:%s", vmName, destPath)

	// Create ignite copy options with source and destination.
	cpFlags := igniteRun.CPFlags{}
	copyOpts, err := cpFlags.NewCPOptions(source, dest)
	if err != nil {
		return err
	}

	if err := igniteRun.CP(copyOpts); err != nil {
		return err
	}
	return nil
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
	runCmd.Flags().StringVarP(&appCmd, "cmd", "c", "", "Command passed to the container app")
	runCmd.Flags().StringArrayVarP(&appCmdArgs, "arg", "a", appCmdArgs, "Arguments to the command passed to the container app")
	runCmd.Flags().StringVar(&mountSrcPath, "mount-src", "", "local path that needs to be mounted in the application container")
	runCmd.Flags().StringVar(&mountDestPath, "mount-dest", "", "path in the application container where the source path is mounted")
}
