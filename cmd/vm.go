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
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
)

var (
	images    []string
	baseImage string
)

const (
	buildContainerPrefix = "ignite-cntr-build"
	ctrPath              = "/usr/bin/ctr"
	containerdNamespace  = "ignite"
)

// vmCmd represents the vm command
var vmCmd = &cobra.Command{
	Use:   "vm <vm-image-name>",
	Short: "Create VM application image.",
	Long: `Create VM application image with preloaded container images. These
images can be quickly run when the VM starts.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("require one VM image name argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		vmImage := args[0]
		if err := runVMImageBuild(vmImage, baseImage, images); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func runVMImageBuild(vmImage string, baseImage string, containerImages []string) error {
	var vmImageName, vmImageTag string

	// Separate image name and tag.
	vmImg := strings.SplitN(vmImage, ":", 2)
	vmImageName = vmImg[0]
	if len(vmImg) < 2 {
		vmImageTag = "latest"
	} else {
		vmImageTag = vmImg[1]
	}

	// Initialize a docker client.
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	ctx := context.Background()

	rand.Seed(time.Now().UnixNano())

	// Create a build container using the base image with random name. The
	// container needs to stay around, run infinite sleep.
	buildContainerName := fmt.Sprintf("%s-%d", buildContainerPrefix, rand.Int())
	containerOpts := docker.CreateContainerOptions{
		Name: buildContainerName,
		Config: &docker.Config{
			Image: baseImage,
			Cmd:   []string{"sleep", "infinity"},
		},
		HostConfig: &docker.HostConfig{
			Privileged: true,
		},
		Context: ctx,
	}
	container, err := client.CreateContainer(containerOpts)
	if err != nil {
		return fmt.Errorf("failed to create build container: %v", err)
	}
	if err := client.StartContainerWithContext(container.ID, nil, ctx); err != nil {
		return fmt.Errorf("failed to start build container: %v", err)
	}
	fmt.Printf("Started build container %s\n", container.Name)

	// Start containerd inside the build container.
	fmt.Println("Starting containerd in the build container...")
	execOpts := docker.CreateExecOptions{
		Privileged: true,
		Container:  container.ID,
		Cmd:        []string{"/usr/bin/containerd", "&"},
	}
	cntrExec, err := client.CreateExec(execOpts)
	if err != nil {
		return err
	}
	startExecOpts := docker.StartExecOptions{
		Detach: true,
	}
	execCloser, err := client.StartExecNonBlocking(cntrExec.ID, startExecOpts)
	if execCloser != nil {
		defer execCloser.Close()
	}

	// Create ignite containerd namespace.
	fmt.Printf("Creating containerd namespace: %s...\n", containerdNamespace)
	createNSExecOpts := execOpts
	createNSExecOpts.Cmd = []string{ctrPath, "namespace", "create", containerdNamespace}
	createNSExec, err := client.CreateExec(createNSExecOpts)
	if err != nil {
		return err
	}
	err = client.StartExec(createNSExec.ID, startExecOpts)
	if err != nil {
		return err
	}

	// Pull the application images.
	for _, containerImage := range containerImages {
		pullExecOpts := execOpts
		pullExecOpts.Cmd = []string{
			ctrPath, fmt.Sprintf("--namespace=%s", containerdNamespace),
			"image", "pull", containerImage,
		}
		pullExec, err := client.CreateExec(pullExecOpts)
		if err != nil {
			return err
		}
		err = client.StartExec(pullExec.ID, startExecOpts)
		if err != nil {
			return err
		}

		fmt.Printf("Waiting for %s image pull to complete", containerImage)
		for {
			inspectRes, err := client.InspectExec(pullExec.ID)
			if err != nil {
				return nil
			}

			if !inspectRes.Running {
				break
			}
			fmt.Printf(".")
			time.Sleep(3 * time.Second)
		}
		// Newline.
		fmt.Println()
	}

	// Commit the container to create an image.
	commitOpts := docker.CommitContainerOptions{
		Container:  container.ID,
		Repository: vmImageName,
		Tag:        vmImageTag,
		Context:    ctx,
	}
	finalImg, err := client.CommitContainer(commitOpts)
	if err != nil {
		return err
	}

	fmt.Printf("\nCreated VM application image: %s:%s (%s)\n", vmImageName, vmImageTag, finalImg.ID)

	// Delete the build container.
	removeContainerOpts := docker.RemoveContainerOptions{
		ID:      container.ID,
		Force:   true,
		Context: ctx,
	}
	if err := client.RemoveContainer(removeContainerOpts); err != nil {
		return err
	}

	return nil
}

func init() {
	imageCmd.AddCommand(vmCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// vmCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// vmCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	vmCmd.Flags().StringArrayVarP(&images, "image", "i", images, "Set an image to be loaded")
	vmCmd.Flags().StringVarP(&baseImage, "baseImage", "b", defaultBaseImage, "Base image of the VM image build container")
}
