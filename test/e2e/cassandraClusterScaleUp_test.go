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

	t.Logf("Create Cluster with 1 DC of 1 rack of 1 node")
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

	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1",
		1); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e"); err != nil {
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

	t.Logf("Add 1 node to first statefulset)")
	cc.Spec.Topology.DC[0].NodesPerRacks = func(i int32) *int32 { return &i }(2)
	err = f.Client.Update(goctx.TODO(), cc)
	if err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1",
		2); err != nil {
		t.Fatal(err)
	}
	if err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e"); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForPodOperationDone(t, f, namespace, "cassandra-e2e",
		"dc1-rack1"); err != nil {
		t.Fatal(err)
	}

	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
		cc); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, api.ActionScaleUp.Name, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)

	assert.Equal(t, api.OperationCleanup, cc.Status.CassandraRackStatus["dc1-rack1"].PodLastOperation.Name)
	assert.Equal(t, []string(nil), cc.Status.CassandraRackStatus["dc1-rack1"].PodLastOperation.Pods)
	assert.ElementsMatch(t, []string{"cassandra-e2e-dc1-rack1-0", "cassandra-e2e-dc1-rack1-1"}, cc.Status.CassandraRackStatus["dc1-rack1"].PodLastOperation.PodsOK)

	assert.Equal(t, api.ActionScaleUp.Name, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
}
