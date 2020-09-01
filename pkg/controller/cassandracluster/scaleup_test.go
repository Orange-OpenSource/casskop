package cassandracluster

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func stringOfSlice(a []string) string {
	q := make([]string, len(a))
	for i, s := range a {
		q[i] = fmt.Sprintf("%q", s)
	}
	return fmt.Sprintf("[%s]", strings.Join(q, ", "))
}

func registerJolokiaOperationJoiningNodes(host podName, numberOfJoiningNodes int) {
	joiningNodes := []string{}
	for i:=1; i<=numberOfJoiningNodes; i++ {
		joiningNodes = append(joiningNodes, "nodeX")
	}
	httpmock.RegisterResponder("POST", JolokiaURL(host.FullName, jolokiaPort),
		httpmock.NewStringResponder(200, fmt.Sprintf(`{"request":
											{"mbean": "org.apache.cassandra.db:type=StorageService",
											 "attribute": "joiningNodes",
											 "type": "read"},
										"value": %s,
										"timestamp": 1528850319,
										"status": 200}`, stringOfSlice(joiningNodes))))
}

func TestAddTwoNodes(t *testing.T) {
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

	cassandraCluster.Spec.NodesPerRacks = 5
	rcc.client.Update(context.TODO(), cassandraCluster)

	firstPod := podHost(stfsName, 0, rcc)
	reconcileValidation(t, rcc, *req)
	assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 0)
	assertStatefulsetReplicas(t, rcc, 3, cassandraCluster.Namespace, stfsName)

	//Reconcile adds one node at a time when it's asked to add more than one node
	for expectedReplicas:=3; expectedReplicas <= 4; expectedReplicas++ {
		//Reconcile does not update the number of nodes when there are joining nodes
		registerJolokiaOperationJoiningNodes(firstPod, 1)
		for reconcileIteration := 0; reconcileIteration <= 2; reconcileIteration++ {
			reconcileValidation(t, rcc, *req)
			assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 1)
			assertStatefulsetReplicas(t, rcc, expectedReplicas, cassandraCluster.Namespace, stfsName)
		}

		//Reconcile adds a node as soon as there are no longer joining nodes
		registerJolokiaOperationJoiningNodes(firstPod, 0)
		reconcileValidation(t, rcc, *req)
		assert.GreaterOrEqual(jolokiaCallsCount(firstPod), 1)
		assertStatefulsetReplicas(t, rcc, expectedReplicas + 1, cassandraCluster.Namespace, stfsName)
	}
}
