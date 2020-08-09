package e2e

import (
	"fmt"
	"strconv"
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

func cassandraClusterScaleDown2RacksFrom3NodesTo1Node(t *testing.T, f *framework.Framework, ctx *framework.Context) {
	namespace, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC.yaml", namespace)
	cc.Namespace = namespace
	cc.Spec.NodesPerRacks = int32(3)
	DC := &cc.Spec.Topology.DC[0]
	DC.Rack = append(DC.Rack, api.Rack{Name: "rack2"})

	if err = f.Client.Create(goctx.TODO(), cc, &framework.CleanupOptions{TestContext: ctx,
		Timeout:       mye2eutil.CleanupTimeout,
		RetryInterval: mye2eutil.CleanupRetryInterval}); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Logf("Error Creating cassandracluster: %v", err)
		t.Fatal(err)
	}

	for _, rack := range DC.Rack {
		if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace,
			fmt.Sprintf("%s-%s-%s", cc.Name, DC.Name, rack.Name),
			int(cc.Spec.NodesPerRacks), mye2eutil.RetryInterval, 2 * mye2eutil.Timeout); err != nil {
			t.Fatal(err)
		}
	}

	if err = mye2eutil.WaitForStatusDone(t, f, namespace, cc.Name, mye2eutil.RetryInterval,
		mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

	for _, rack := range DC.Rack {
		DCRackName := fmt.Sprintf("%s-%s", DC.Name, rack.Name)
		cassandraLastAction := cc.Status.CassandraRackStatus[DCRackName].CassandraLastAction
		assert.Equal(t, api.ClusterPhaseInitial.Name, cassandraLastAction.Name)
		assert.Equal(t, api.StatusDone, cassandraLastAction.Status)
	}

	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
	assert.Equal(t, api.ClusterPhaseInitial.Name, cc.Status.LastClusterAction)

	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

	cc.Spec.NodesPerRacks = int32(1)

	if err = f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatal(err)
	}

	for _, rack := range DC.Rack {
		if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace,
			fmt.Sprintf("%s-%s-%s", cc.Name, DC.Name, rack.Name),
			int(cc.Spec.NodesPerRacks), mye2eutil.RetryInterval, 2 * mye2eutil.Timeout); err != nil {
			t.Fatal(err)
		}
	}

	if err = mye2eutil.WaitForStatusDone(t, f, namespace, cc.Name, mye2eutil.RetryInterval,
		mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	numberOfNodesSeenCmd := "nodetool status|grep -ic rack"
	numberOfNodesSeen, _, _ := mye2eutil.ExecPodFromName(t, f, namespace,
		fmt.Sprintf("%s-%s-%s", cc.Name, DC.Name, DC.Rack[0].Name), numberOfNodesSeenCmd)
	assert.Equal(t, strconv.Itoa(int(2 * cc.Spec.NodesPerRacks)), numberOfNodesSeen)
}

func cassandraClusterScaleDownDC2Test(t *testing.T, f *framework.Framework, ctx *framework.Context) {
	namespace, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-2DC.yaml", namespace)
	cc.Namespace = namespace

	if err = f.Client.Create(goctx.TODO(), cc, &framework.CleanupOptions{TestContext: ctx,
		Timeout:       mye2eutil.CleanupTimeout,
		RetryInterval: mye2eutil.CleanupRetryInterval}); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Logf("Error Creating cassandracluster: %v", err)
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 1,
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc2-rack1", 1,
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatusDone(t, f, namespace, cc.Name, mye2eutil.RetryInterval,
		mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

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

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e-dc1-rack1-0", Namespace: namespace}, pod)
	for _, keyspace := range keyspaces {
		cmd := fmt.Sprintf(Strategy2DC, keyspace)
		if _, _, err = mye2eutil.ExecPod(t, f, cc.Namespace, pod, []string{"bash", "-c", cmd}); err != nil {
			t.Fatalf("Error exec change keyspace %s = %v", keyspace, err)
		}
	}
	time.Sleep(2 * time.Second)

	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

	nodesPerRack := int32(0)
	cc.Spec.Topology.DC[1].NodesPerRacks = &nodesPerRack

	if err = f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Second)

	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

	assert.Equal(t, api.ActionCorrectCRDConfig.Name, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)

	f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e-dc1-rack1-0", Namespace: namespace}, pod)

	for i := range keyspaces {
		cmd := fmt.Sprintf(Strategy1DC, keyspaces[i])
		_, _, err = mye2eutil.ExecPod(t, f, cc.Namespace, pod, []string{"bash", "-c", cmd})
		if err != nil {
			t.Fatalf("Error exec change keyspace %s{%s} = %v", keyspaces[i], cmd, err)
		}
	}
	time.Sleep(2 * time.Second)

	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

	cc.Spec.Topology.DC[1].NodesPerRacks = &nodesPerRack

	if err = f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc2-rack1", 0,
		mye2eutil.RetryInterval, 2 * mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	if err = mye2eutil.WaitForStatusDone(t, f, namespace, cc.Name, mye2eutil.RetryInterval,
		mye2eutil.Timeout); err != nil {
		t.Fatal(err)
	}

	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

	assert.Equal(t, api.ActionScaleDown.Name, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	statefulset, err := f.KubeClient.AppsV1().StatefulSets(namespace).Get(goctx.TODO(),
		"cassandra-e2e-dc2-rack1", metav1.GetOptions{})
	assert.Equal(t, int32(0), statefulset.Status.CurrentReplicas)

	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

	cc.Spec.Topology.DC.Remove(1)
	if err = f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Second)

	statefulset, err = f.KubeClient.AppsV1().StatefulSets(namespace).Get(goctx.TODO(),
		"cassandra-e2e-dc2-rack1", metav1.GetOptions{})

	svc := &v1.Pod{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}}
	names := []string{"cassandra-e2e-dc2"}
	for _, name := range names {
		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, svc)
		assert.Equal(t, true, apierrors.IsNotFound(err))
	}

	cc = &api.CassandraCluster{}
	mye2eutil.K8sGetCassandraCluster(t, f, err, cc)

	assert.Equal(t, 1, len(cc.Status.CassandraRackStatus))
	assert.Equal(t, api.ActionDeleteDC.Name, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
}

