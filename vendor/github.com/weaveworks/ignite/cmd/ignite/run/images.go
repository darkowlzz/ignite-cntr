package run

import (
	"github.com/weaveworks/gitops-toolkit/pkg/filter"
	api "github.com/weaveworks/ignite/pkg/apis/ignite"
	"github.com/weaveworks/ignite/pkg/providers"
	"github.com/weaveworks/ignite/pkg/util"
)

type imagesOptions struct {
	allImages []*api.Image
}

func NewImagesOptions() (io *imagesOptions, err error) {
	io = &imagesOptions{}
	io.allImages, err = providers.Client.Images().FindAll(filter.NewAllFilter())
	return
}

func Images(io *imagesOptions) error {
	o := util.NewOutput()
	defer o.Flush()

	o.Write("IMAGE ID", "NAME", "CREATED", "SIZE")
	for _, image := range io.allImages {
		o.Write(image.GetUID(), image.GetName(), image.GetCreated(), image.Status.OCISource.Size.String())
	}

	return nil
}
