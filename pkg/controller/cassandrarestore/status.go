package cassandrarestore

import (
	"context"
	"fmt"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/util"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateRestoreStatus(c client.Client, restore *api.CassandraRestore, status api.CassandraRestoreStatus, reqLogger logr.Logger) error {
	typeMeta := restore.TypeMeta

	status.Condition.LastTransitionTime = metav1.Now().Format(util.TimeStampLayout)
	restore.Status = status

	err := updateRestoreStatus(c, restore)

	if err != nil {
		if !apierrors.IsConflict(err) {
			return fmt.Errorf("could not update CR state")
		}

		err := c.Get(context.TODO(), types.NamespacedName{
			Name: restore.Name,
			Namespace: restore.Namespace,
		}, restore)

		if err != nil {
			return fmt.Errorf("could not get config for updating status")
		}
		restore.Status = status

		err = updateRestoreStatus(c, restore)

		if err != nil {
			return fmt.Errorf("could not update Restore state")
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	restore.TypeMeta = typeMeta
	reqLogger.Info("Restore state updated")
	return nil
}

func updateRestoreStatus(c client.Client, restore *api.CassandraRestore) error {
	err := c.Status().Update(context.TODO() , restore)
	if apierrors.IsNotFound(err) {
		return c.Update(context.TODO(), restore)
	}
	return nil
}
