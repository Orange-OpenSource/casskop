package e2e

import (
	goctx "context"
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis"
	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	mye2eutil "github.com/Orange-OpenSource/cassandra-k8s-operator/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Run all fonctional tests
func TestCassandraCluster(t *testing.T) {
	cassandracluster := &api.CassandraCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CassandraCluster",
			APIVersion: "db.orange.com/v1alpha1",
		},
	}

	err := framework.AddToFrameworkScheme(apis.AddToScheme, cassandracluster)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	// run subtests
	t.Run("group", func(t *testing.T) {
		t.Run("ClusterScaleUp", CassandraClusterTest(cassandraClusterScaleUpDC1Test))
		t.Run("ClusterScaleDownSimple", CassandraClusterTest(cassandraClusterScaleDownSimpleTest))
		t.Run("ClusterScaleDown", CassandraClusterTest(cassandraClusterScaleDownDC2Test))
		t.Run("RollingRestart", CassandraClusterTest(cassandraClusterRollingRestartDCTest))
	})

}

func CassandraClusterTest(code func(t *testing.T, f *framework.Framework,
	ctx *framework.TestCtx)) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, f := mye2eutil.HelperInitOperator(t)
		defer ctx.Cleanup()
		code(t, f, ctx)
	}

}

//cassandraClusterRollingRestartDCTest test the rolling restart of a DC
// 1. It starts a cluster with 2dc :
//    dc1-rack1 (1 node) and dc2-rack2 (1 node)
//    We check everything went fine
// 2. We trigger a rolling restart of DC dc1-rack1
//    We check the 1st statefulset has a new version
// 3. We trigger a rolling restart of DC dc2-rack2
//    We check the 2nd statefulset has a new version
func cassandraClusterRollingRestartDCTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("Could not get namespace: %v", err)
	}

	t.Log("Create the Cluster with 1 DC consisting of 1 rack of 1 node")

	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC.yaml", namespace)
	t.Logf("Create CassandraCluster cassandracluster-1DC.yaml in namespace %s", namespace)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	if err := f.Client.Create(goctx.TODO(), cc, &framework.CleanupOptions{TestContext: ctx,
		Timeout:       mye2eutil.CleanupTimeout,
		RetryInterval: mye2eutil.CleanupRetryInterval}); err != nil {
		t.Fatalf("Error Creating cassandracluster: %v", err)
	}

	if err := mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 1,
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("WaitForStatefulset got an error: %v", err)
	}

	if err := mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e",
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("WaitForStatusDone got an error: %v", err)
	}

	t.Log("Download statefulset and store current revision")
	statefulset, err := f.KubeClient.AppsV1().StatefulSets(namespace).Get("cassandra-e2e-dc1-rack1",
		metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		t.Fatalf("Could not download statefulset: %v", err)
	}

	stfsVersion := statefulset.Status.CurrentRevision

	t.Log("Download last version of CassandraCluster")
	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e",
		Namespace: namespace}, cc); err != nil {
		t.Fatalf("Could not download last version of CassandraCluster: %v", err)
	}

	t.Log("Trigger rolling restart of 1st DC by updating RollingRestart flag")

	cc.Spec.Topology.DC[0].Rack[0].RollingRestart = true

	if err := f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatalf("Could not update CassandraCluster: %v", err)
	}

	if err := mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 1,
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("WaitForStatefulset got an error: %v", err)
	}

	if err := mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e",
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("WaitForStatusDone got an error: %v", err)
	}

	t.Log("Download statefulset and check current revision has been updated")
	statefulset, err = f.KubeClient.AppsV1().StatefulSets(namespace).Get("cassandra-e2e-dc1-rack1",
		metav1.GetOptions{IncludeUninitialized: true})
	if err != nil {
		t.Fatalf("Could not download statefulset: %v", err)
	}

	assert.NotEqual(t, stfsVersion, statefulset.Status.CurrentRevision)

	t.Logf("Current labels on statefulset : %v", statefulset.Spec.Template.Labels)

	_, rollingRestartLabelExists := statefulset.Spec.Template.Labels["rolling-restart"]

	t.Log("Assert that rolling-restart label has been added to statefulset")
	assert.Equal(t, true, rollingRestartLabelExists)

	t.Log("Download last version of CassandraCluster and check RollingRestart flag has been cleaned out")

	// Delete Topology in order to get it updated as a whole
	cc.Spec.Topology.DC = nil

	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e",
		Namespace: namespace}, cc); err != nil {
		t.Fatalf("Could not get CassandraCluster: %v", err)
	}

	assert.Equal(t, false, cc.Spec.Topology.DC[0].Rack[0].RollingRestart)
}
