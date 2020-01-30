package multicasskop

import (
	"context"
	"github.com/Orange-OpenSource/casskop/multi-casskop/pkg/controller/multi-casskop/models"

	ccv1 "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha2"
	"github.com/kylelemons/godebug/pretty"
	"github.com/sirupsen/logrus"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
)

// ReadyCassandraCluster
// return true if CassandraCluster it Done and Running
func (r *reconciler) ReadyCassandraCluster(cc *ccv1.CassandraCluster) bool {
	if cc.Status.Phase != ccv1.ClusterPhaseRunning || cc.Status.LastClusterActionStatus != ccv1.StatusDone {
		return false
	}
	return true
}

/*
func (r *reconciler) GetCassandraCluster(client *Client, cc *ccv1.CassandraCluster) (*ccv1.CassandraCluster, error) {

	storedCC := &ccv1.CassandraCluster{}
	if err := client.client.Get(context.TODO(), r.namespacedName(cc.Name, cc.Namespace), storedCC); err != nil {
		if errors.IsNotFound(err) {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace, "kubernetes": client.name}).Debug("CassandraCluster don't exists, we create it ")
			return nil, err
		}
		return storedCC, err
	}
}
*/

// CreateOrUpdateCassandraCluster
// create CassandraCluster object in target kubernetes cluster if not exists
// update it if it already exist
func (r *reconciler) CreateOrUpdateCassandraCluster(client *models.Client,
	cc *ccv1.CassandraCluster) (bool, *ccv1.CassandraCluster, error) {
	storedCC := &ccv1.CassandraCluster{}

	if err := client.Client.Get(context.TODO(), r.namespacedName(cc.Name, cc.Namespace), storedCC); err != nil {
		if errors.IsNotFound(err) {
			logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace, "kubernetes": client.Name}).Debug("CassandraCluster don't exists, we create it ")
			newCC, err := r.CreateCassandraCluster(client, cc)
			return true, newCC, err
		}
		return false, storedCC, err
	}

	needUpdate := false
	//TODO: need new way to detect changes
	if !apiequality.Semantic.DeepEqual(storedCC.Spec, cc.Spec) {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace, "kubernetes": client.Name}).
			Info("CassandraCluster is different: " + pretty.Compare(storedCC.Spec, cc.Spec))
		storedCC.Spec = cc.Spec
		needUpdate = true
	}

	//Multi-CassKop manages the Seedlist, we ensure that managed Casskop won't deal themselves with the seedlist
	cc.Spec.AutoUpdateSeedList = false

	if cc.Status.SeedList != nil &&
		!apiequality.Semantic.DeepEqual(storedCC.Status.SeedList, cc.Status.SeedList) {
		logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace, "kubernetes": client.Name}).
			Info("SeedList is different: " + pretty.Compare(storedCC.Status.SeedList, cc.Status.SeedList))
		storedCC.Status.SeedList = cc.Status.SeedList
		needUpdate = true
	}

	if needUpdate {
		newCC, err := r.UpdateCassandraCluster(client, storedCC)
		return true, newCC, err
	}
	return false, storedCC, nil
}

func (r *reconciler) CreateCassandraCluster(client *models.Client, cc *ccv1.CassandraCluster) (*ccv1.CassandraCluster, error) {
	var err error
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace, "kubernetes": client.Name}).Debug("Create CassandraCluster")
	if err = client.Client.Create(context.TODO(), cc); err != nil {
		if errors.IsAlreadyExists(err) {
			return cc, nil
		}
	}
	return cc, err
}

func (r *reconciler) UpdateCassandraCluster(client *models.Client, cc *ccv1.CassandraCluster) (*ccv1.CassandraCluster, error) {
	var err error
	logrus.WithFields(logrus.Fields{"cluster": cc.Name, "namespace": cc.Namespace, "kubernetes": client.Name}).Debug("Update CassandraCluster")
	if err = client.Client.Update(context.TODO(), cc); err != nil {
		if errors.IsAlreadyExists(err) {
			return cc, nil
		}
	}
	return cc, err
}
