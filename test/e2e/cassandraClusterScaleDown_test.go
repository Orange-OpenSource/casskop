package e2e

import (
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"

	goctx "context"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	mye2eutil "github.com/Orange-OpenSource/casskop/test/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func cassandraClusterScaleDownSimpleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	t.Logf("Create Cluster with 1 DC of 1 rack of 2 nodes")
	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC.yaml", namespace)
	cc.Namespace = namespace
	cc.Spec.Topology.DC[0].NodesPerRacks = func(idx int32) *int32 { return &idx }(2)
	err = f.Client.Create(goctx.TODO(), cc, &framework.CleanupOptions{TestContext: ctx,
		Timeout:       mye2eutil.CleanupTimeout,
		RetryInterval: mye2eutil.CleanupRetryInterval})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		t.Logf("Error Creating cassandracluster: %v", err)
		t.Fatal(err)
	}
	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1",
		2); err != nil {
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

	t.Logf("Scale down DC1 to 1 node")
	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
		cc); err != nil {
		t.Fatal(err)
	}
	cc.Spec.Topology.DC[0].NodesPerRacks = func(idx int32) *int32 { return &idx }(1)
	if err = f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1",
		1); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e"); err != nil {
		t.Fatal(err)
	}

	cc = &api.CassandraCluster{}
	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
		cc); err != nil {
		t.Fatal(err)
	}
	//Because AutoUpdateSeedList is false we stay on ScaleUp=Done status
	assert.Equal(t, api.ActionScaleDown.Name, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)

	assert.Equal(t, api.ActionScaleDown.Name, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
}

func cassandraClusterDeleteSecondDC(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	t.Logf("Create Cluster with 2DCs of 1 rack of 1 node each")

	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-2DC.yaml", namespace)
	cc.Namespace = namespace

	t.Logf("Create CassandraCluster cassandracluster-2DC.yaml in namespace %s", namespace)
	err = f.Client.Create(goctx.TODO(), cc, &framework.CleanupOptions{TestContext: ctx,
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

	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc2-rack1",
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

	grepNumTokens := "grep num_tokens: /etc/cassandra/cassandra.yaml"
	res, _, _ := mye2eutil.ExecPodFromName(t, f, namespace, "cassandra-e2e-dc1-rack1-0", grepNumTokens)
	assert.Equal(t, "num_tokens: 256", res)
	res, _, _ = mye2eutil.ExecPodFromName(t, f, namespace, "cassandra-e2e-dc2-rack1-0", grepNumTokens)
	assert.Equal(t, "num_tokens: 32", res)

	const Strategy1DC = "cqlsh -u cassandra -p cassandra -e \"ALTER KEYSPACE %s WITH REPLICATION = {'class" +
		"' : 'NetworkTopologyStrategy', 'dc1' : 1};\""
	const Strategy2DC = "cqlsh -u cassandra -p cassandra -e \"ALTER KEYSPACE %s WITH REPLICATION = {'class" +
		"' : 'NetworkTopologyStrategy', 'dc1' : 1, 'dc2' : 1};\""
	keyspaces := []string{"system_auth", "system_distributed", "system_traces"}
	pod := &v1.Pod{TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"}}

	t.Log("Change replication topology to 2 DCs ")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e-dc1-rack1-0", Namespace: namespace}, pod)
	for keyspace := range keyspaces {
		cmd := fmt.Sprintf(Strategy2DC, keyspaces[keyspace])
		_, _, err = mye2eutil.ExecPod(t, f, cc.Namespace, pod, []string{"bash", "-c", cmd})
		if err != nil {
			t.Fatalf("Error exec change keyspace %s = %v", keyspaces[keyspace], err)
		}
	}
	time.Sleep(2 * time.Second)

	t.Logf("Attempt to scale down DC2 to 0 nodes which should fail as cassandra is still replicating to DC2")
	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
		cc); err != nil {
		t.Fatal(err)
	}
	cc.Spec.Topology.DC[1].NodesPerRacks = func(idx int32) *int32 { return &idx }(0)
	if err = f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Second)

	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
		cc); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, api.ActionCorrectCRDConfig.Name, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)

	t.Logf("Remove DC2 from keyspace's replication")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e-dc1-rack1-0", Namespace: namespace}, pod)

	for keyspace := range keyspaces {
		cmd := fmt.Sprintf(Strategy1DC, keyspaces[keyspace])
		_, _, err = mye2eutil.ExecPod(t, f, cc.Namespace, pod, []string{"bash", "-c", cmd})
		if err != nil {
			t.Fatalf("Error exec change keyspace %s{%s} = %v", keyspaces[keyspace], cmd, err)
		}
	}
	time.Sleep(2 * time.Second)

	t.Logf("ScaleDown DC2 to 0 nodes")
	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
	cc); err != nil {
		t.Fatal(err)
	}
	cc.Spec.Topology.DC[1].NodesPerRacks = func(idx int32) *int32 { return &idx }(0)
	if err = f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc2-rack1",
		0); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e"); err != nil {
		t.Fatal(err)
	}
	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace},
		cc); err != nil {
		t.Fatal(err)
	}
	//Because AutoUpdateSeedList is false we stay on ScaleUp=Done status
	assert.Equal(t, api.ActionScaleDown.Name, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	statefulset, err := f.KubeClient.AppsV1().StatefulSets(namespace).Get(goctx.TODO(),
		"cassandra-e2e-dc2-rack1", metav1.GetOptions{})
	assert.Equal(t, int32(0), statefulset.Status.CurrentReplicas)


	t.Logf("Remove DC2")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	cc.Spec.Topology.DC.Remove(1)
	err = f.Client.Update(goctx.TODO(), cc)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Second)

	t.Log("Check Statefulset is deleted")
	statefulset, err = f.KubeClient.AppsV1().StatefulSets(namespace).Get(goctx.TODO(),
		"cassandra-e2e-dc2-rack1", metav1.GetOptions{})

	t.Log("Check Service is deleted")
	svc := &v1.Pod{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}}
	names := []string{
		"cassandra-e2e-dc2",
	}
	for name := range names {
		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: names[name], Namespace: namespace}, svc)
		assert.Equal(t, true, apierrors.IsNotFound(err))
	}

	cc = &api.CassandraCluster{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, len(cc.Status.CassandraRackStatus))
	assert.Equal(t, api.ActionDeleteDC.Name, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
}
