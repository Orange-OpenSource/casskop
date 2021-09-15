package cassandrarestore

import (
	"context"
	"emperror.dev/errors"
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v2"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateRestoreStatus(c client.Client, restore *api.CassandraRestore, status api.BackRestStatus,
	reqLogger *logrus.Entry) error {
	patch := client.MergeFrom(restore.DeepCopy())
	restore.Status = status

	if err := c.Patch(context.Background(), restore, patch); err != nil {
			return errors.WrapIfWithDetails(err, "could not update status for restore",
				"restore", restore)
	}

	return nil
}