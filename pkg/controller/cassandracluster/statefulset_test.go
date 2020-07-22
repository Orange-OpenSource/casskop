package cassandracluster

import (
	"fmt"
	"reflect"
	"testing"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	"github.com/stretchr/testify/assert"
)

type liveAndReadinessProbeExpected struct {
	replaceValue int32
	areEquals    bool
}

func TestStatefulSetsAreEqual(t *testing.T) {
	dcName := "dc1"
	rackName := "rack1"
	dcRackName := fmt.Sprintf("%s-%s", dcName, rackName)

	_, cc := HelperInitCluster(t, "cassandracluster-2DC.yaml")
	cc.CheckDefaults()
	labels, nodeSelector := k8s.DCRackLabelsAndNodeSelectorForStatefulSet(cc, 0, 0)
	sts, _ := generateCassandraStatefulSet(cc, &cc.Status, dcName, dcRackName, labels, nodeSelector, nil)

	testSetup := make(map[string]liveAndReadinessProbeExpected)

	// Readiness update test
	testSetup["ReadinessInitialDelaySeconds"]	= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}
	testSetup["ReadinessHealthCheckTimeout"] 	= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}
	testSetup["ReadinessHealthCheckPeriod"] 	= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}
	testSetup["ReadinessFailureThreshold"] 		= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}
	testSetup["ReadinessSuccessThreshold"] 		= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}

	// Liveness update test
	testSetup["LivenessInitialDelaySeconds"]	= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}
	testSetup["LivenessHealthCheckTimeout"] 	= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}
	testSetup["LivenessHealthCheckPeriod"] 		= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}
	testSetup["LivenessFailureThreshold"]		= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}
	testSetup["LivenessSuccessThreshold"]		= liveAndReadinessProbeExpected{areEquals:false, replaceValue: 2000}

	for field, lvExpected := range testSetup {
		ccNew := generateNewCC(cc, field, lvExpected.replaceValue)
		if ccNew == nil {
			t.Errorf("Don't succeed to find the field : %s, in the CassandraCluster.Spec struct.", field)
		} else {
			checkStatefulsetEquality(t, dcName, dcRackName, sts, ccNew, lvExpected.areEquals)
		}
	}
}

func checkStatefulsetEquality(t *testing.T, dcName, dcRackName string, stsOld *appsv1.StatefulSet, ccNew *api.CassandraCluster, expectedResult bool) {

	ccNew.CheckDefaults()
	labelsDefault, nodeSelectorDefault := k8s.DCRackLabelsAndNodeSelectorForStatefulSet(ccNew, 0, 0)
	stsDefault, _ := generateCassandraStatefulSet(ccNew, &ccNew.Status, dcName, dcRackName, labelsDefault, nodeSelectorDefault, nil)

	assert.Equal(t, expectedResult, statefulSetsAreEqual(stsOld, stsDefault))
}

func generateNewCC(cc *api.CassandraCluster, field string, newValue int32) *api.CassandraCluster {

	ccNew := cc.DeepCopy()
	v := reflect.ValueOf(&ccNew.Spec).Elem().FieldByName(field)
	if v.IsValid() {
		v.Set(reflect.ValueOf(&newValue))
		return ccNew
	}

	return nil
}
