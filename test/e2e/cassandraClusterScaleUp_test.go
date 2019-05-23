package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"

	goctx "context"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	mye2eutil "github.com/Orange-OpenSource/cassandra-k8s-operator/test/e2eutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

//cassandraClusterScaleUpDC1Test test the scaleUp of a cassandracluster
// 1. it starts a cluster with 1dc and 1 rack and 1 node in the dc-rack
//    We check all is Good
// 2. We scaleUp 1 node in dc1-rack1
//    We check all is Good
func cassandraClusterScaleUpDC1Test(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	t.Logf("0. Init Operator")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	/*----
	 */
	t.Logf("1. We Create the Cluster (1dc/1rack/1node")
	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC.yaml", namespace)
	cc.Namespace = namespace
	t.Logf("Create CassandraCluster cassandracluster-1DC.yaml in namespace %s", namespace)
	err = f.Client.Create(goctx.TODO(), cc,
		&framework.CleanupOptions{TestContext: ctx,
			Timeout:       mye2eutil.CleanupTimeout,
			RetryInterval: mye2eutil.CleanupRetryInterval})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		t.Logf("Error Creating cassandracluster: %v", err)
		t.Fatal(err)
	}
	// wait for statefulset dc1-rack1
	err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 1,
		mye2eutil.RetryInterval,
		mye2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}
	// Operator status is Done when initializing the initial cluster is OK
	err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e", mye2eutil.RetryInterval, mye2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}
	//Get Updated cc
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	//locale dc-rack state is OK
	assert.Equal(t, api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	//global cluster state is OK
	assert.Equal(t, api.ClusterPhaseInitial, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)

	/*----
	 */
	t.Logf(" 2. We Request a ScaleUp (add 1 node in the first dc-rack)")
	cc.Spec.Topology.DC[0].NodesPerRacks = func(i int32) *int32 { return &i }(2)
	err = f.Client.Update(goctx.TODO(), cc)
	if err != nil {
		t.Fatal(err)
	}
	// wait for statefulset dc1-rack1
	err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 2, mye2eutil.RetryInterval,
		mye2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}
	err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e", mye2eutil.RetryInterval, mye2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}
	//Get Updated cc
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("We make some assertions")
	//locale dc-rack state is OK: Because AutoUpdateSeedList is false we stay on ScaleUp=Done status
	assert.Equal(t, api.ActionScaleUp, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	//global cluster state is OK
	assert.Equal(t, api.ActionScaleUp, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
}
