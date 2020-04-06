package storage

import (
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/gitops-toolkit/pkg/storage"
	"github.com/weaveworks/gitops-toolkit/pkg/storage/cache"
	"github.com/weaveworks/ignite/pkg/apis/ignite/scheme"
	"github.com/weaveworks/ignite/pkg/constants"
	"github.com/weaveworks/ignite/pkg/providers"
)

func SetGenericStorage() error {
	log.Trace("Initializing the GenericStorage provider...")
	providers.Storage = cache.NewCache(
		storage.NewGenericStorage(
			storage.NewGenericRawStorage(constants.DATA_DIR), scheme.Serializer))
	return nil
}
