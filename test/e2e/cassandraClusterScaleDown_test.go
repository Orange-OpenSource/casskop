package e2e

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"

	goctx "context"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	api "github.com/Orange-OpenSource/cassandra-k8s-operator/pkg/apis/db/v1alpha1"
	mye2eutil "github.com/Orange-OpenSource/cassandra-k8s-operator/test/e2eutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//cassandraClusterScaleDownSimpleTest test the scaleDown of a DC
// 1. it starts a cluster with 1dc and 2 nodes:
//    dc1-rack1 (2 nodes)
//    We check all is Good
// 2. We scaleDown to 1 the dc1-rack1
//    We check all is Good
func cassandraClusterScaleDownSimpleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	t.Logf("0. Init Operator")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	/*----
	 */
	t.Logf("1. We Create the Cluster (1dc/1rack/2node")

	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC.yaml", namespace)
	cc.Namespace = namespace
	cc.Spec.Topology.DC[0].NodesPerRacks = func(i int32) *int32 { return &i }(2)

	t.Logf("Create CassandraCluster cassandracluster-1DC.yaml in namespace %s", namespace)
	err = f.Client.Create(goctx.TODO(), cc, &framework.CleanupOptions{TestContext: ctx,
		Timeout:       mye2eutil.CleanupTimeout,
		RetryInterval: mye2eutil.CleanupRetryInterval})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		t.Logf("Error Creating cassandracluster: %v", err)
		t.Fatal(err)
	}
	// wait for statefulset dc1-rack1 2 nodes
	err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 2,
		mye2eutil.RetryInterval,
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
	//locale dc-rack state is OK
	assert.Equal(t, api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(t, api.ClusterPhaseInitial, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)

	/*----
	 */
	t.Logf("3. We Request a ScaleDown to 1")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	cc.Spec.Topology.DC[0].NodesPerRacks = func(i int32) *int32 { return &i }(1)
	err = f.Client.Update(goctx.TODO(), cc)
	if err != nil {
		t.Fatal(err)
	}
	// wait for statefulset dc1-rack1
	err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 1,
		mye2eutil.RetryInterval,
		mye2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}
	err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e", mye2eutil.RetryInterval, mye2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Get Updated cc")
	cc = &api.CassandraCluster{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	//Because AutoUpdateSeedList is false we stay on ScaleUp=Done status
	assert.Equal(t, api.ActionScaleDown, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	//Check Global state
	assert.Equal(t, api.ActionScaleDown, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
}

//cassandraClusterScaleDownDC2Test test the scaleDown of a DC
// 1. it starts a cluster with 2dc :
//    dc1-rack1 (1 node) and dc2-rack2 (1 node)
//    We check all is Good
// 2. We scaleDown to 0 the dc2-rack2
//    We check all is Good
// 3. We Remove the dc2
//    We check all is Good (check that there is not more old Pods, statefulset, services.. associated to the removes dc2
func cassandraClusterScaleDownDC2Test(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	t.Logf("0. Init Operator")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("could not get namespace: %v", err)
	}

	/*----
	 */
	t.Logf("1. We Create the Cluster (2dc/1rack/1node")

	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-2DC.yaml", namespace)
	cc.Namespace = namespace
	t.Logf("Create CassandraCluster cassandracluster-2DC.yaml in namespace %s", namespace)
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), cc, &framework.CleanupOptions{TestContext: ctx,
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
	// wait for statefulset dc1-rack1
	err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc2-rack1", 1,
		mye2eutil.RetryInterval,
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
	//locale dc-rack state is OK
	assert.Equal(t, api.ClusterPhaseInitial, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc1-rack1"].CassandraLastAction.Status)
	assert.Equal(t, api.ClusterPhaseInitial, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)

	//Check that numTokens are 256 (default) for dc1 and 32 (as specified in the crd) for dc2
	grepNumTokens := "grep num_tokens: /etc/cassandra/cassandra.yaml"
	res, _, _ := mye2eutil.ExecPodFromName(t, f, namespace, "cassandra-e2e-dc1-rack1-0", grepNumTokens)
	assert.Equal(t, "num_tokens: 256", res)
	res, _, _ = mye2eutil.ExecPodFromName(t, f, namespace, "cassandra-e2e-dc2-rack1-0", grepNumTokens)
	assert.Equal(t, "num_tokens: 32", res)
	/*----
	 */
	const Strategy1DC = "cqlsh -u cassandra -p cassandra -e \"ALTER KEYSPACE %s WITH REPLICATION = {'class" +
		"' : 'NetworkTopologyStrategy', 'dc1' : 1};\""
	const Strategy2DC = "cqlsh -u cassandra -p cassandra -e \"ALTER KEYSPACE %s WITH REPLICATION = {'class" +
		"' : 'NetworkTopologyStrategy', 'dc1' : 1, 'dc2' : 1};\""
	keyspaces := []string{
		"system_auth",
		"system_distributed",
		"system_traces",
	}
	pod := &v1.Pod{TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"}}

	t.Log("We Change Replication Topology to 2 DC ")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e-dc1-rack1-0", Namespace: namespace}, pod)
	for i := range keyspaces {
		cmd := fmt.Sprintf(Strategy2DC, keyspaces[i])
		_, _, err = mye2eutil.ExecPod(t, f, cc.Namespace, pod, []string{"bash", "-c", cmd})
		if err != nil {
			t.Logf("Error exec change keyspace %s = %v", keyspaces[i], err)
		}
	}

	t.Logf("2. Ask Scale Down to 0 but this will be refused")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	cc.Spec.Topology.DC[1].NodesPerRacks = func(i int32) *int32 { return &i }(0)
	err = f.Client.Update(goctx.TODO(), cc)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)

	//Check Result
	t.Log("Get Updated cc")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	//Operator has restore old CRD
	assert.Equal(t, api.ActionCorrectCRDConfig, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)

	/*----
	 */
	t.Logf("3. We Remove the replication to the dc2")
	//pod := &v1.Pod{TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"}}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e-dc1-rack1-0", Namespace: namespace}, pod)

	for i := range keyspaces {
		cmd := fmt.Sprintf(Strategy1DC, keyspaces[i])
		_, _, err = mye2eutil.ExecPod(t, f, cc.Namespace, pod, []string{"bash", "-c", cmd})
		if err != nil {
			t.Logf("Error exec change keyspace %s = %v", keyspaces[i], err)
		}
	}

	/*----
	 */
	t.Logf("4. We Request a ScaleDown to 0 prior to remove a DC")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	cc.Spec.Topology.DC[1].NodesPerRacks = func(i int32) *int32 { return &i }(0)
	err = f.Client.Update(goctx.TODO(), cc)
	if err != nil {
		t.Fatal(err)
	}
	// wait for statefulset dc1-rack1
	err = mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc2-rack1", 0,
		mye2eutil.RetryInterval,
		mye2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}
	err = mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e", mye2eutil.RetryInterval, mye2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Get Updated cc")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	//Because AutoUpdateSeedList is false we stay on ScaleUp=Done status
	assert.Equal(t, api.ActionScaleDown, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Name)
	assert.Equal(t, api.StatusDone, cc.Status.CassandraRackStatus["dc2-rack1"].CassandraLastAction.Status)

	statefulset, err := f.KubeClient.AppsV1().StatefulSets(namespace).Get("cassandra-e2e-dc2-rack1",
		metav1.GetOptions{IncludeUninitialized: true})
	assert.Equal(t, int32(0), statefulset.Status.CurrentReplicas)

	/*----
	 */
	t.Logf("5. We Remove the DC")
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}
	cc.Spec.Topology.DC.Remove(1)
	err = f.Client.Update(goctx.TODO(), cc)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Second)

	t.Log("Check Statefulset is deleted")
	statefulset, err = f.KubeClient.AppsV1().StatefulSets(namespace).Get("cassandra-e2e-dc2-rack1",
		metav1.GetOptions{IncludeUninitialized: true})
	//assert.Equal(t, true, apierrors.IsNotFound(err))

	t.Log("Check Service is deleted")
	svc := &v1.Pod{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}}
	names := []string{
		"cassandra-e2e-dc2",
		"cassandra-e2e-dc2-rack1",
		"cassandra-e2e-dc2-rack1-exporter-jmx",
	}
	for i := range names {
		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: names[i], Namespace: namespace}, svc)
		assert.Equal(t, true, apierrors.IsNotFound(err))
	}

	t.Log("Get Updated cc")
	cc = &api.CassandraCluster{}
	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e", Namespace: namespace}, cc)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Actual status is %s", cc.Status.LastClusterAction)
	//We have only 1 dcRack in status
	assert.Equal(t, 1, len(cc.Status.CassandraRackStatus))
	assert.Equal(t, api.ActionDeleteDC, cc.Status.LastClusterAction)
	assert.Equal(t, api.StatusDone, cc.Status.LastClusterActionStatus)
}
