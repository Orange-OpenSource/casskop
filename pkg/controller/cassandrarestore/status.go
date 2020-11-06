package cassandrarestore

import (
	"context"
	"fmt"
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/controller/common"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateRestoreStatus(c client.Client, restore *api.CassandraRestore, status api.BackRestStatus,
	reqLogger *logrus.Entry) error {
	typeMeta := restore.TypeMeta

	restore.Status = status

	if err := updateRestoreStatus(c, restore); err != nil {
		if !apierrors.IsConflict(err) {
			return fmt.Errorf("Could not update CR state")
		}

		err := c.Get(context.TODO(), types.NamespacedName{Name: restore.Name, Namespace: restore.Namespace}, restore)

		if err != nil {
			return fmt.Errorf("Could not get config for updating status")
		}
		restore.Status = status

		if err = updateRestoreStatus(c, restore); err != nil {
			return fmt.Errorf("Could not update Restore state")
		}
	}
	// update loses the typeMeta of the config that's used later when setting ownerrefs
	restore.TypeMeta = typeMeta
	reqLogger.Info("Restore state updated")
	return nil
}

func updateRestoreStatus(c client.Client, restore *api.CassandraRestore) error {
	patchToApply, err := common.JsonPatch(map[string]interface{}{"status": restore.Status})
	if err != nil {
		return err
	}
	cassandraRestore := &api.CassandraRestore{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: restore.Namespace,
			Name:      restore.Name,
		}}

	if err := c.Patch(context.Background(), cassandraRestore, patchToApply); err != nil {
		return err
	}
	return nil
}
