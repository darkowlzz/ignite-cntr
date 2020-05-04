package cmd

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/spf13/cobra"
)

const (
	// defaultBaseImage is the base container image used as the build container. It
	// contains a pre-installed container-runtime.
	defaultBaseImage = "darkowlzz/ignite-cntr-base:dev"

	// defaultFromImage is the base container image on which the ignite-cntr
	// base image is based on. This should be an ignite compatible image.
	defaultFromImage = "weaveworks/ignite-ubuntu:18.04"
)

var (
	// baseFromImage is the flag variable to store the from image of the base.
	baseFromImage string
)

// baseCmd represents the base command
var baseCmd = &cobra.Command{
	Use:   "base [<base-image-name>]",
	Short: "Create VM base image.",
	Long: `Create base image for the VM image. This base image contains a container
runtime and other common dependencies.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("require at one or no base image name argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		targetImage := defaultBaseImage
		if len(args) == 1 {
			targetImage = args[0]
		}
		if err := runBase(baseFromImage, targetImage); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

func runBase(baseFromImage, targetImage string) error {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return err
	}

	// Write dockerfile in tar format as input to the docker build server.
	t := time.Now()
	inputbuf, outputbuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputbuf)

	dockerfile := fmt.Sprintf(`FROM %s
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
		/usr/share/local/*`, baseFromImage)

	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len(dockerfile)), ModTime: t, AccessTime: t, ChangeTime: t})
	tr.Write([]byte(dockerfile))
	tr.Close()

	opts := docker.BuildImageOptions{
		Name:         targetImage,
		InputStream:  inputbuf,
		OutputStream: outputbuf,
	}
	fmt.Println("Building image...")
	if err := client.BuildImage(opts); err != nil {
		return err
	}

	fmt.Printf("Base image built: %s\n", targetImage)
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

	baseCmd.Flags().StringVarP(&baseFromImage, "baseImage", "b", defaultFromImage, "Base image of the VM base image")
}
