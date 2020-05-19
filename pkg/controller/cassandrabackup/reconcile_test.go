package cassandrabackup

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cbyaml = `
apiVersion: db.orange.com/v1alpha1
kind: CassandraBackup
metadata:
  name: test-cassandra-backup
  namespace: default
  labels:
    app: cassandra
spec:
  cassandracluster: test-cluster-dc1
  cluster: test-cluster
  datacenter: dc1
  storageLocation: s3://cassie
  snapshotTag: SnapshotTag2
  secret: cloud-backup-secrets
`

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("../../../testdata", name)
	fmt.Println(path)
	fmt.Println("loading input from previous file")
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func helperInitCassandraBackup(cassandraBackupYaml string) api.CassandraBackup {
	var cassandraBackup api.CassandraBackup
	if err := yaml.Unmarshal([]byte(cassandraBackupYaml), &cassandraBackup); err != nil {
		log.Error(err, "error: helpInitCluster")
		os.Exit(-1)
	}
	return cassandraBackup
}

func helperInitCassandraBackupController(t *testing.T, cassandraBackupYaml string) (*ReconcileCassandraBackup, *api.CassandraBackup, *record.FakeRecorder) {
	cassandraBackup := helperInitCassandraBackup(cassandraBackupYaml)

	cassandraBackupList := api.CassandraBackupList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CassandraBackupList",
			APIVersion: api.SchemeGroupVersion.String(),
		},
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(api.SchemeGroupVersion, &api.CassandraCluster{})
	s.AddKnownTypes(api.SchemeGroupVersion, &api.CassandraClusterList{})
	s.AddKnownTypes(api.SchemeGroupVersion, &cassandraBackup)
	s.AddKnownTypes(api.SchemeGroupVersion, &cassandraBackupList)

	//Create Fake client
	//Objects to track in the Fake client
	objs := []runtime.Object{
		&cassandraBackup,
	}
	fakeClient := fake.NewFakeClient(objs...)

	// Create a ReconcileCassandraBackup object with the scheme and fake client.
	fakeRecorder := record.NewFakeRecorder(3)
	reconcileCassandraBackup := ReconcileCassandraBackup{client: fakeClient, scheme: s, recorder: fakeRecorder}

	return &reconcileCassandraBackup, &cassandraBackup, fakeRecorder
}

func TestCassandraBackupJustCreate(t *testing.T) {
	// When instance.JustCreate Reconcile stops
	rcb, cb, _ := helperInitCassandraBackupController(t, cbyaml)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cb.Name,
			Namespace: cb.Namespace,
		},
	}

	cb.JustCreate = true

	rcb.client.Update(context.TODO(), cb)

	res, err := rcb.Reconcile(req)

	assert := assert.New(t)

	// Confirms that Reconcile stops and does not return any error
	assert.Equal(reconcile.Result{}, res)
	assert.Nil(err)

}

func TestCassandraBackupAlreadyExists(t *testing.T) {
	// When same CassandraBackup is recreated Reconcile stops and an event is created
	rcb, cb, recorder := helperInitCassandraBackupController(t, cbyaml)

	oldBackup := cb.DeepCopy()
	oldBackup.Status = []*api.CassandraBackupStatus{
		{
			Node:     "node1",
			State:    "COMPLETED",
			Progress: "Done",
		},
	}
	oldBackup.Name = "prev-test-cassandra-backup"
	rcb.client.Create(context.TODO(), oldBackup)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cb.Name,
			Namespace: cb.Namespace,
		},
	}

	res, err := rcb.Reconcile(req)

	assert := assert.New(t)

	// Confirms that Reconcile stops and does not return any error
	assert.Equal(reconcile.Result{}, res)
	assert.Nil(err)

	msg := <-recorder.Events
	assert.Contains(msg, "SnapshotTag2 because such backup already exists")
}

func TestCassandraBackupSecretNotFound(t *testing.T) {
	// When secret does not exist Reconcile stops and an event is created
	rcb, cb, recorder := helperInitCassandraBackupController(t, cbyaml)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cb.Name,
			Namespace: cb.Namespace,
		},
	}

	res, err := rcb.Reconcile(req)

	assert := assert.New(t)

	// Confirms that Reconcile stops and does not return any error
	assert.Equal(reconcile.Result{}, res)
	assert.Nil(err)

	msg := <-recorder.Events
	assert.Contains(msg, "Secret cloud-backup-secrets used for backups was not found")
}

func TestCassandraBackupIncorrectAwsCreds(t *testing.T) {
	// awsregion must be set when both awssecretaccesskey and awsaccesskeyid are set
	cb := helperInitCassandraBackup(cbyaml)

	var reqLogger = logf.Log.WithName("controller_cassandrabackup")
	secret := &corev1.Secret{
		Data: map[string][]byte{
			"awssecretaccesskey": []byte("a secret"),
			"awsaccesskeyid":     []byte("an access key"),
		},
	}

	resp := validateBackupSecret(secret, &cb, reqLogger)

	assert := assert.New(t)
	assert.Contains(resp.Error(), "there is no awsregion property while you have set both awssecretaccesskey and awsaccesskeyid")

	secret.Data["awsregion"] = []byte("a region")

	resp = validateBackupSecret(secret, &cb, reqLogger)
	assert.Nil(resp)
}

func TestCassandraBackupDatacenterNotFound(t *testing.T) {
	// when datacenter does not exist Reconcile returns an error
	rcb, cb, recorder := helperInitCassandraBackupController(t, cbyaml)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cb.Spec.Secret,
			Namespace: cb.Namespace,
		},
		Data: map[string][]byte{
			"awssecretaccesskey": []byte("a secret"),
			"awsaccesskeyid":     []byte("an access key"),
			"awsregion":          []byte("a region"),
		},
	}

	rcb.client.Create(context.TODO(), secret)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cb.Name,
			Namespace: cb.Namespace,
		},
	}

	res, err := rcb.Reconcile(req)

	assert := assert.New(t)

	// Confirms that Reconcile stops and does not return any error
	assert.Equal(reconcile.Result{}, res)
	assert.Nil(err)

	msg := <-recorder.Events
	assert.Contains(msg, "Datacenter dc1 of cluster test-cluster-dc1 to backup not found")
}
