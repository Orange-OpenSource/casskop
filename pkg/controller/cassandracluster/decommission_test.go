package cassandracluster

import (
	"context"
	"fmt"
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"strconv"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func reconcileValidation(t *testing.T, rcc *ReconcileCassandraCluster, request reconcile.Request) {
	if res, err := rcc.Reconcile(request); err != nil {
		t.Fatalf("reconcile: (%v)", err)
	} else if !res.Requeue && res.RequeueAfter == 0 {
		t.Error("reconcile did not requeue request as expected")
	}
}

func createCassandraClusterWithNoDisruption(t *testing.T, cassandraClusterFileName string) (*ReconcileCassandraCluster,
	*reconcile.Request) {
	rcc, req := helperCreateCassandraCluster(t, cassandraClusterFileName)

	pdb := &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rcc.cc.Name,
			Namespace: rcc.cc.Namespace,
		},
	}

	rcc.client.Get(context.TODO(), req.NamespacedName, pdb)
	// No disruption
	pdb.Status.DisruptionsAllowed = 1
	rcc.client.Update(context.TODO(), pdb)

	return rcc, req
}

func registerJolokiaOperationModeResponder(host podName, op operationMode) {
	httpmock.RegisterResponder("POST", JolokiaURL(host.FullName, jolokiaPort),
		httpmock.NewStringResponder(200, fmt.Sprintf(`{"request":
											{"mbean": "org.apache.cassandra.db:type=StorageService",
											 "attribute": "OperationMode",
											 "type": "read"},
										"value": "%s",
										"timestamp": 1528850319,
										"status": 200}`, string(op))))
}

func registerFatalJolokiaResponder(t *testing.T, host podName) {
	httpmock.RegisterResponder("POST", JolokiaURL(host.FullName, jolokiaPort),
		httpmock.NewNotFoundResponder(t.Fatal))
}

func jolokiaCallsCount(name podName) int {
	info := httpmock.GetCallCountInfo()
	return info[fmt.Sprintf("POST http://%s:8778/jolokia/", name.FullName)]
}

type podName struct {
	Name string
	FullName string
}

func podHost(stfsName string, id int8, rcc *ReconcileCassandraCluster) podName {
	name := stfsName + strconv.Itoa(int(id))
	return podName{name, name + "." + rcc.cc.Name}
}

func deletePodNotDeletedByFakeClient(rcc *ReconcileCassandraCluster, host podName) {
	// Need to manually delete pod managed by the fake client
	rcc.client.Delete(context.TODO(), &v1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name:      host.Name,
		Namespace: rcc.cc.Namespace}})
}

func TestOneDecommission(t *testing.T) {
	rcc, req := createCassandraClusterWithNoDisruption(t, "cassandracluster-1DC.yaml")

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	assert := assert.New(t)

	assert.Equal(int32(3), rcc.cc.Spec.NodesPerRacks)

	cassandraCluster := rcc.cc.DeepCopy()

	datacenters := cassandraCluster.Spec.Topology.DC
	assert.Equal(1, len(datacenters))
	assert.Equal(1, len(datacenters[0].Rack))

	dc := datacenters[0]
	stfsName := cassandraCluster.Name + fmt.Sprintf("-%s-%s", dc.Name, dc.Rack[0].Name)

	cassandraCluster.Spec.NodesPerRacks = 2
	rcc.client.Update(context.TODO(), cassandraCluster)

	lastPod := podHost(stfsName, 2, rcc)

	registerFatalJolokiaResponder(t, podHost(stfsName, int8(1), rcc))
	registerJolokiaOperationModeResponder(lastPod, NORMAL)
	reconcileValidation(t, rcc, *req)
	assert.GreaterOrEqual(jolokiaCallsCount(lastPod), 1)
	assertStatefulsetReplicas(t, rcc, 3, cassandraCluster.Namespace, stfsName)

	registerJolokiaOperationModeResponder(lastPod, LEAVING)
	reconcileValidation(t, rcc, *req)
	assert.GreaterOrEqual(jolokiaCallsCount(lastPod), 1)
	assertStatefulsetReplicas(t, rcc, 3, cassandraCluster.Namespace, stfsName)

	registerJolokiaOperationModeResponder(lastPod, DECOMMISSIONED)
	reconcileValidation(t, rcc, *req)
	assert.GreaterOrEqual(jolokiaCallsCount(lastPod), 1)
	assertStatefulsetReplicas(t, rcc, 2, cassandraCluster.Namespace, stfsName)

	deletedPod := podHost(stfsName, 2, rcc)
	assert.Equal(1, jolokiaCallsCount(deletedPod))

	lastPod = podHost(stfsName, 1, rcc)
	deletePodNotDeletedByFakeClient(rcc, deletedPod)
	stfs, _ := rcc.GetStatefulSet(namespace, stfsName)
	stfs.Status.ReadyReplicas = 2
	rcc.client.Update(context.TODO(), stfs)

	registerFatalJolokiaResponder(t, deletedPod)
	registerJolokiaOperationModeResponder(lastPod, NORMAL)
	reconcileValidation(t, rcc, *req)
	assert.Equal(0, jolokiaCallsCount(lastPod))
	assert.Equal(api.StatusDone, rcc.cc.Status.CassandraRackStatus["dc1-rack1"].PodLastOperation.Status)

	reconcileValidation(t, rcc, *req)
	assert.Equal(0, jolokiaCallsCount(lastPod))
}

func assertStatefulsetReplicas(t *testing.T, rcc *ReconcileCassandraCluster, expected int, namespace, stfsName string){
	assert := assert.New(t)
	stfs, _ := rcc.GetStatefulSet(namespace, stfsName)
	assert.Equal(int32(expected), *stfs.Spec.Replicas)
}


