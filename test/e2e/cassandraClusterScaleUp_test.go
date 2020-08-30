package e2e

import (
	goctx "context"
	"testing"

	"github.com/stretchr/testify/assert"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	mye2eutil "github.com/Orange-OpenSource/casskop/test/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func cassandraClusterScaleUpDC1Test(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	t.Logf("0. Init Operator")

	namespace, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}
	t.Logf("1. We Create the Cluster (1dc/1rack/1node")
	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC.yaml", namespace)
	cc.Namespace = namespace
	t.Logf("Create CassandraCluster cassandracluster-1DC.yaml in namespace %s", namespace)
	if err = f.Client.Create(goctx.TODO(), cc,
		&framework.CleanupOptions{TestContext: ctx,
			Timeout:       mye2eutil.CleanupTimeout,
			RetryInterval: mye2eutil.CleanupRetryInterval}); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Logf("Error Creating cassandracluster: %v", err)
		t.Fatal(err)
	}
	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1",
		1, mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}
	if err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e",
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}
	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
		cc); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, api.ClusterPhaseInitial.Name, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(t, api.ClusterPhaseInitial.Name, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)

	t.Logf(" 2. We Request a ScaleUp (add 1 node in the first dc-rack)")
	cc.Spec.Topology.DC[0].NodesPerRacks = func(i int32) *int32 { return &i }(2)
	if err = f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatal(err)
	}
	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1",
		2, mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}
	if err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e",
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForPodOperationDone(t, f, namespace, "cassandra-e2e",
		"dc1-rack1", mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
		cc); err != nil {
		t.Fatal(err)
	}

	t.Logf("We make some assertions")

	assert.Equal(t, api.ActionScaleUp.Name, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)

	assert.Equal(t, api.OperationCleanup, cc.Status.CassandraRackStatus["dc1-rack1"].PodLastOperation.Name)
	assert.Equal(t, []string(nil), cc.Status.CassandraRackStatus["dc1-rack1"].PodLastOperation.Pods)
	assert.ElementsMatch(t, []string{"cassandra-e2e-dc1-rack1-0", "cassandra-e2e-dc1-rack1-1"},
		cc.Status.CassandraRackStatus["dc1-rack1"].PodLastOperation.PodsOK)

	assert.Equal(t, api.ActionScaleUp.Name, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
}
