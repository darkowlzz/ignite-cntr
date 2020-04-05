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
	"archive/tar"
	"bytes"
	"fmt"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
)

var imageName = "darkowlzz/ignite-cntr-base:dev"

// baseCmd represents the base command
var baseCmd = &cobra.Command{
	Use:   "base",
	Short: "Create VM base image.",
	Long: `Create base image for the VM image. This base image contains a container
runtime and other common dependencies.`,
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("base called")
		if err := runBase(); err != nil {
			fmt.Printf("fail: %v", err)
		}
	},
}

func runBase() error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	t := time.Now()
	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputbuf)

	dockerfile := `FROM weaveworks/ignite-ubuntu:latest
RUN apt-get update -y \
	&& apt-get install -y --no-install-recommends containerd \
	&& apt-get clean -y \
	&& rm -rf \
		/var/cache/debconf/* \
		/var/lib/apt/lists/* \
		/var/log/* \
		/tmp/* \
		/var/tmp/* \
		/usr/share/doc/* \
		/usr/share/man/* \
		/usr/share/local/*`

	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(dockerfile)), ModTime: t, AccessTime: t, ChangeTime: t})
	tr.Write([]byte(dockerfile))
	tr.Close()

	opts := docker.BuildImageOptions{
		Name:         imageName,
		InputStream:  inputbuf,
		OutputStream: outputbuf,
	}
	if err := client.BuildImage(opts); err != nil {
		return err
	}

	fmt.Printf("Base image built: %s\n", imageName)
	return nil
}

func init() {
	imageCmd.AddCommand(baseCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// baseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// baseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
