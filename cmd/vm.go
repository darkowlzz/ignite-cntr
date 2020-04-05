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
	"fmt"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
)

// vmCmd represents the vm command
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Create VM application image.",
	Long: `Create VM application image with preloaded container images. These
images can be quickly run when the VM starts.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("vm called")
		if err := runVM(); err != nil {
			fmt.Printf("fail: %v", err)
		}
	},
}

func runVM() error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	ctx := context.Background()

	containerOpts := docker.CreateContainerOptions{
		Name: "ignite-cntr-build",
		Config: &docker.Config{
			Image: imageName,
			Cmd:   []string{"sleep", "infinity"},
		},
		HostConfig: &docker.HostConfig{
			Privileged: true,
		},
		Context: ctx,
	}
	container, err := client.CreateContainer(containerOpts)
	if err != nil {
		return err
	}

	if err := client.StartContainerWithContext(container.ID, nil, ctx); err != nil {
		return err
	}
	fmt.Printf("Started build container %s", container.Name)

	// Start containerd inside the build container.
	execOpts := docker.CreateExecOptions{
		Privileged: true,
		Container:  container.ID,
		Cmd:        []string{"/usr/bin/containerd", "&"},
	}
	cntrExec, err := client.CreateExec(execOpts)
	if err != nil {
		return err
	}

	fmt.Println("Starting containerd...")

	startExecOpts := docker.StartExecOptions{
		Detach: true,
	}
	execCloser, err := client.StartExecNonBlocking(cntrExec.ID, startExecOpts)
	if execCloser != nil {
		defer execCloser.Close()
	}

	fmt.Println("Creating containerd namespace: ignite")
	// Create ignite containerd namespace.
	createNSExecOpts := execOpts
	createNSExecOpts.Cmd = []string{"/usr/bin/ctr", "namespace", "create", "ignite"}
	createNSExec, err := client.CreateExec(createNSExecOpts)
	if err != nil {
		return err
	}
	err = client.StartExec(createNSExec.ID, startExecOpts)
	if err != nil {
		return err
	}

	// Pull the application image.
	// appImage := "docker.io/library/busybox:latest"
	appImage := "quay.io/coreos/etcd:v3.4.7"
	pullExecOpts := execOpts
	pullExecOpts.Cmd = []string{"/usr/bin/ctr", "--namespace=ignite", "image", "pull", appImage}
	pullExec, err := client.CreateExec(pullExecOpts)
	if err != nil {
		return err
	}
	err = client.StartExec(pullExec.ID, startExecOpts)
	if err != nil {
		return err
	}

	fmt.Printf("waiting for the image pull to complete")
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

	// Commit the container to create an image.
	vmImage := "darkowlzz/ignite-etcd"
	vmTag := "dev"
	commitOpts := docker.CommitContainerOptions{
		Container:  container.ID,
		Repository: vmImage,
		Tag:        vmTag,
		Context:    ctx,
	}
	finalImg, err := client.CommitContainer(commitOpts)
	if err != nil {
		return err
	}

	fmt.Printf("Created VM application image: %s:%s (%s)", vmImage, vmTag, finalImg.ID)

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
}
