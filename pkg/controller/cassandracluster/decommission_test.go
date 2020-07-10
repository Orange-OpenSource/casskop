package cassandracluster

import (
	"context"
	"fmt"
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
	rcc, req := helperCreateCassandraCluster(t, "cassandracluster-1DC.yaml")

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

func TestOneDecommission(t *testing.T) {
	rcc, req := createCassandraClusterWithNoDisruption(t, "cassandracluster-1DC.yaml")

	cassandraCluster := rcc.cc.DeepCopy()
	cassandraCluster.Spec.NodesPerRacks = 2
	rcc.client.Update(context.TODO(), cassandraCluster)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	lastPod := "cassandra-demo-dc1-rack12" + "." + rcc.cc.Name
	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
				{"mbean": "org.apache.cassandra.db:type=StorageService",
				 "attribute": "OperationMode",
				 "type": "read"},
			"value": "NORMAL",
			"timestamp": 1528850319,
			"status": 200}`))

	for i := 0; i < 2; i++ {
		otherPod := fmt.Sprintf("cassandra-demo-dc1-rack1%d.%s", i, rcc.cc.Name)

		httpmock.RegisterResponder("POST", JolokiaURL(otherPod, jolokiaPort),
			httpmock.NewNotFoundResponder(t.Fatal))

	}
	reconcileValidation(t, rcc, *req)

	stfsName := cassandraCluster.Name + "-dc1-rack1"
	stfs, _ := rcc.GetStatefulSet(cassandraCluster.Namespace, stfsName)

	assert := assert.New(t)

	assert.Equal(int32(3), *stfs.Spec.Replicas)

	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
			{"mbean": "org.apache.cassandra.db:type=StorageService",
			 "attribute": "OperationMode",
			 "type": "read"},
		"value": "LEAVING",
		"timestamp": 1528850319,
		"status": 200}`))

	reconcileValidation(t, rcc, *req)
	assert.Equal(int32(3), *stfs.Spec.Replicas)

	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
		{"mbean": "org.apache.cassandra.db:type=StorageService",
		 "attribute": "OperationMode",
		 "type": "read"},
	"value": "DECOMMISSIONED",
	"timestamp": 1528850319,
	"status": 200}`))

	reconcileValidation(t, rcc, *req)

	stfs, _ = rcc.GetStatefulSet(cassandraCluster.Namespace, stfsName)
	assert.Equal(int32(2), *stfs.Spec.Replicas)

	info := httpmock.GetCallCountInfo()
	assert.Equal(2, info["POST http://cassandra-demo-dc1-rack12.cassandra-demo:8778/jolokia/"])

	// pods, _ := rcc.ListPods(rcc.cc.Namespace, k8s.LabelsForCassandraDCRack(rcc.cc, "dc1", "rack1"))
	// fmt.Println(len(pods.Items))

	// Need to manually delete pod managed by the fake client
	rcc.client.Delete(context.TODO(), &v1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name:      stfsName + "2",
		Namespace: rcc.cc.Namespace}})

	decommissionedPod := "cassandra-demo-dc1-rack12" + "." + rcc.cc.Name
	lastPod = "cassandra-demo-dc1-rack11" + "." + rcc.cc.Name

	httpmock.RegisterResponder("POST", JolokiaURL(decommissionedPod, jolokiaPort),
		httpmock.NewNotFoundResponder(t.Fatal))
	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
				{"mbean": "org.apache.cassandra.db:type=StorageService",
				 "attribute": "OperationMode",
				 "type": "read"},
			"value": "NORMAL",
			"timestamp": 1528850319,
			"status": 200}`))

	reconcileValidation(t, rcc, *req)
}

func TestMultipleDecommissions(t *testing.T) {
	rcc, req := createCassandraClusterWithNoDisruption(t, "cassandracluster-1DC.yaml")

	cassandraCluster := rcc.cc.DeepCopy()
	cassandraCluster.Spec.NodesPerRacks = 1
	rcc.client.Update(context.TODO(), cassandraCluster)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for i := 0; i <= 1; i++ {
		nonLastPod := "cassandra-demo-dc1-rack10" + "." + rcc.cc.Name

		httpmock.RegisterResponder("POST", JolokiaURL(nonLastPod, jolokiaPort),
			httpmock.NewNotFoundResponder(t.Fatal))
	}

	lastPod := "cassandra-demo-dc1-rack12" + "." + rcc.cc.Name

	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
					{"mbean": "org.apache.cassandra.db:type=StorageService",
					 "attribute": "OperationMode",
					 "type": "read"},
				"value": "NORMAL",
				"timestamp": 1528850319,
				"status": 200}`))

	reconcileValidation(t, rcc, *req)

	stfsName := cassandraCluster.Name + "-dc1-rack1"
	stfs, _ := rcc.GetStatefulSet(cassandraCluster.Namespace, stfsName)

	assert := assert.New(t)

	assert.Equal(int32(3), *stfs.Spec.Replicas)

	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
					{"mbean": "org.apache.cassandra.db:type=StorageService",
					"attribute": "OperationMode",
					"type": "read"},
				"value": "LEAVING",
				"timestamp": 1528850319,
				"status": 200}`))

	reconcileValidation(t, rcc, *req)
	assert.Equal(int32(3), *stfs.Spec.Replicas)

	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
					{"mbean": "org.apache.cassandra.db:type=StorageService",
					"attribute": "OperationMode",
					"type": "read"},
				"value": "DECOMMISSIONED",
				"timestamp": 1528850319,
				"status": 200}`))

	reconcileValidation(t, rcc, *req)

	stfs, _ = rcc.GetStatefulSet(cassandraCluster.Namespace, stfsName)
	assert.Equal(int32(2), *stfs.Spec.Replicas)

	// Need to manually delete pod managed by the fake client
	rcc.client.Delete(context.TODO(), &v1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name:      stfsName + "2",
		Namespace: rcc.cc.Namespace}})

	previousLastPod := "cassandra-demo-dc1-rack12" + "." + rcc.cc.Name
	lastPod = "cassandra-demo-dc1-rack11" + "." + rcc.cc.Name

	httpmock.RegisterResponder("POST", JolokiaURL(previousLastPod, jolokiaPort),
		httpmock.NewNotFoundResponder(t.Fatal))
	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
					{"mbean": "org.apache.cassandra.db:type=StorageService",
					 "attribute": "OperationMode",
					 "type": "read"},
				"value": "NORMAL",
				"timestamp": 1528850319,
				"status": 200}`))
	reconcileValidation(t, rcc, *req)

	stfs, _ = rcc.GetStatefulSet(cassandraCluster.Namespace, stfsName)
	assert.Equal(int32(2), *stfs.Spec.Replicas)

	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
					{"mbean": "org.apache.cassandra.db:type=StorageService",
					"attribute": "OperationMode",
					"type": "read"},
				"value": "LEAVING",
				"timestamp": 1528850319,
				"status": 200}`))

	reconcileValidation(t, rcc, *req)
	assert.Equal(int32(2), *stfs.Spec.Replicas)

	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
					{"mbean": "org.apache.cassandra.db:type=StorageService",
					"attribute": "OperationMode",
					"type": "read"},
				"value": "DECOMMISSIONED",
				"timestamp": 1528850319,
				"status": 200}`))

	reconcileValidation(t, rcc, *req)

	stfs, _ = rcc.GetStatefulSet(cassandraCluster.Namespace, stfsName)
	assert.Equal(int32(1), *stfs.Spec.Replicas)

	// Need to manually delete pod managed by the fake client
	rcc.client.Delete(context.TODO(), &v1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name:      stfsName + "1",
		Namespace: rcc.cc.Namespace}})

	previousLastPod = "cassandra-demo-dc1-rack11" + "." + rcc.cc.Name
	lastPod = "cassandra-demo-dc1-rack10" + "." + rcc.cc.Name

	httpmock.RegisterResponder("POST", JolokiaURL(previousLastPod, jolokiaPort),
		httpmock.NewNotFoundResponder(t.Fatal))
	httpmock.RegisterResponder("POST", JolokiaURL(lastPod, jolokiaPort),
		httpmock.NewStringResponder(200, `{"request":
						{"mbean": "org.apache.cassandra.db:type=StorageService",
						 "attribute": "OperationMode",
						 "type": "read"},
					"value": "NORMAL",
					"timestamp": 1528850319,
					"status": 200}`))

	reconcileValidation(t, rcc, *req)

	// stfs, _ = rcc.GetStatefulSet(cassandraCluster.Namespace, stfsName)
	// assert.Equal(int32(1), *stfs.Spec.Replicas)

}
