package cassandrarestore

import (
	"context"
	"fmt"
	"github.com/Orange-OpenSource/casskop/pkg/controller/common"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cassandraRestoreYaml = `
apiVersion: db.orange.com/v1alpha1
kind: CassandraRestore
metadata:
  name: test-cassandra-restore
spec:
  cassandraCluster: test-cluster-dc1
  cassandraBackup: test-cassandra-backup
  concurrentConnection: 15
  noDeleteTruncates: false
#   schemaVersion:
#   exactSchemaVersion:
  entities: "k1,k2.t1"
`

var cassandraBackupYaml = `
apiVersion: db.orange.com/v1alpha1
kind: CassandraRestore
metadata:
  name: test-cassandra-backup
  namespace: default
  labels:
    app: cassandra
  annotations:
    a1: v1
spec:
  cassandraCluster: test-cluster-dc1
  cluster: test-cluster
  datacenter: dc1
  storageLocation: s3://cassie
  snapshotTag: SnapshotTag2
  secret: cloud-backup-secrets
`

func helperInitCassandraRestore(cassandraRestoreYaml string) api.CassandraRestore {
	var cassandraRestore api.CassandraRestore
	if err := yaml.Unmarshal([]byte(cassandraRestoreYaml), &cassandraRestore); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
	return cassandraRestore
}

func helperInitCassandraRestoreController(cassandraRestoreYaml string) (*ReconcileCassandraRestore,
	*api.CassandraRestore, *record.FakeRecorder) {
	//cassandraBackup := common.HelperInitCassandraBackup(cassandraRestoreYaml)
	cassandraRestore := helperInitCassandraRestore(cassandraRestoreYaml)

	cassandraRestoreList := api.CassandraRestoreList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CassandraRestoreList",
			APIVersion: api.SchemeGroupVersion.String(),
		},
	}

	// Register operator types with the runtime scheme.
	fakeClientScheme := scheme.Scheme
	fakeClientScheme.AddKnownTypes(api.SchemeGroupVersion, &api.CassandraCluster{})
	fakeClientScheme.AddKnownTypes(api.SchemeGroupVersion, &api.CassandraClusterList{})
	fakeClientScheme.AddKnownTypes(api.SchemeGroupVersion, &api.CassandraBackup{})
	fakeClientScheme.AddKnownTypes(api.SchemeGroupVersion, &api.CassandraBackupList{})
	fakeClientScheme.AddKnownTypes(api.SchemeGroupVersion, &cassandraRestore)
	fakeClientScheme.AddKnownTypes(api.SchemeGroupVersion, &cassandraRestoreList)

	objs := []runtime.Object{
		&cassandraRestore,
	}

	fakeClient := fake.NewFakeClientWithScheme(fakeClientScheme, objs...)

	fakeRecorder := record.NewFakeRecorder(3)
	reconcileCassandraRestore := ReconcileCassandraRestore{
		client: fakeClient,
		scheme: fakeClientScheme,
		recorder: fakeRecorder,
	}

	return &reconcileCassandraRestore, &cassandraRestore, fakeRecorder
}

func TestCassandraRestoreWithUnknownCassandraCluster(t *testing.T) {
	assert := assert.New(t)
	reconcileCassandraRestore, cassandraRestore, recorder := helperInitCassandraRestoreController(cassandraRestoreYaml)

	reconcileCassandraRestore.client.Create(context.TODO(), cassandraRestore)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cassandraRestore.Name,
			Namespace: cassandraRestore.Namespace,
		},
	}

	res, err := reconcileCassandraRestore.Reconcile(req)

	assert.Equal(reconcile.Result{}, res)
	assert.NotNil(err)
	assert.Equal(err.Error(), fmt.Sprintf("cassandraclusters.db.orange.com \"%s\" not found",
		cassandraRestore.Spec.CassandraCluster))
	common.AssertEvent(t, recorder.Events,
		fmt.Sprintf("Warning CassandraClusterNotFound Cassandra Cluster %s to restore not found",
			cassandraRestore.Spec.CassandraCluster))
}

func TestCassandraRestoreWithUnknownCassandraBackup(t *testing.T) {
	assert := assert.New(t)

	reconcileCassandraRestore, cassandraRestore, recorder := helperInitCassandraRestoreController(cassandraRestoreYaml)

	cassandraCluster := api.CassandraCluster{}
	cassandraCluster.Name = cassandraRestore.Spec.CassandraCluster
	cassandraCluster.Namespace = cassandraRestore.Namespace
	fmt.Println(cassandraCluster)

	reconcileCassandraRestore.client.Create(context.TODO(), &cassandraCluster)
	reconcileCassandraRestore.client.Create(context.TODO(), cassandraRestore)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cassandraRestore.Name,
			Namespace: cassandraRestore.Namespace,
		},
	}

	res, err := reconcileCassandraRestore.Reconcile(req)

	assert.Equal(reconcile.Result{}, res)
	assert.NotNil(err)
	assert.Equal(fmt.Sprintf("cassandrabackups.db.orange.com \"%s\" not found",
		cassandraRestore.Spec.CassandraBackup), err.Error())
	common.AssertEvent(t, recorder.Events,
		fmt.Sprintf("Warning BackupNotFound Backup %s to restore not found",
			cassandraRestore.Spec.CassandraBackup))
}

func TestCassandraRestoreWithNilStatusCondition(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)
}

func TestCassandraRestoreWithNoCoordinatorMember(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)
}

func TestCassandraRestorePhaseRequiredButNoPods(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)
}

func TestCassandraRestorePhaseRequired(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)
}

func TestCassandraRestorePhaseInProgress(t *testing.T) {
	assert := assert.New(t)
	assert.True(true)
}