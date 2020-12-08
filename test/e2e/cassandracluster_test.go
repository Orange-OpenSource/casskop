package e2e

import (
	goctx "context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"os"
	"testing"
	"time"

	"github.com/Orange-OpenSource/casskop/pkg/apis"
	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/k8s"
	mye2eutil "github.com/Orange-OpenSource/casskop/test/e2eutil"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

// Run all fonctional tests
func TestCassandraCluster(t *testing.T) {
	cassandracluster := &api.CassandraCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CassandraCluster",
			APIVersion: "db.orange.com/v1alpha1",
		},
	}

	err := framework.AddToFrameworkScheme(apis.AddToScheme, cassandracluster)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	// run subtests
	t.Run("group", func(t *testing.T) {
		t.Run("ClusterScaleUp", CassandraClusterTest(cassandraClusterScaleUpDC1Test))
		t.Run("ClusterScaleDown", CassandraClusterTest(cassandraClusterScaleDown2RacksFrom3NodesTo1Node))
		t.Run("ClusterScaleDownSimple", CassandraClusterTest(cassandraClusterScaleDownDC2Test))
		t.Run("RollingRestart", CassandraClusterTest(cassandraClusterRollingRestartDCTest))
		t.Run("CreateOneClusterService", CassandraClusterTest(cassandraClusterServiceTest))
		t.Run("UpdateConfigMap", CassandraClusterTest(cassandraClusterUpdateConfigMapTest))
		t.Run("ExecuteCleanup", CassandraClusterTest(cassandraClusterCleanupTest))
	})

}

func CassandraClusterTest(code func(t *testing.T, f *framework.Framework,
	ctx *framework.Context)) func(t *testing.T) {
	return func(t *testing.T) {
		ctx, f := mye2eutil.HelperInitOperator(t)
		defer ctx.Cleanup()
		code(t, f, ctx)
	}

}

//cassandraClusterRollingRestartDCTest test the rolling restart of a DC
// 1. It starts a cluster with 2dc :
//    dc1-rack1 (1 node) and dc2-rack2 (1 node)
//    We check everything went fine
// 2. We trigger a rolling restart of DC dc1-rack1
//    We check the 1st statefulset has a new version
// 3. We trigger a rolling restart of DC dc2-rack2
//    We check the 2nd statefulset has a new version
func cassandraClusterRollingRestartDCTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatalf("Could not get namespace: %v", err)
	}

	t.Log("Create the Cluster with 1 DC consisting of 1 rack of 1 node")

	cc := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC.yaml", namespace)
	t.Logf("Create CassandraCluster cassandracluster-1DC.yaml in namespace %s", namespace)

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	if err := f.Client.Create(goctx.TODO(), cc, &framework.CleanupOptions{TestContext: ctx,
		Timeout:       mye2eutil.CleanupTimeout,
		RetryInterval: mye2eutil.CleanupRetryInterval}); err != nil {
		t.Fatalf("Error Creating cassandracluster: %v", err)
	}

	if err := mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 1,
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("WaitForStatefulset got an error: %v", err)
	}

	if err := mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e",
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("WaitForStatusDone got an error: %v", err)
	}

	t.Log("Download statefulset and store current revision")
	statefulset, err := f.KubeClient.AppsV1().StatefulSets(namespace).Get(goctx.TODO(),
		"cassandra-e2e-dc1-rack1", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Could not download statefulset: %v", err)
	}

	stfsVersion := statefulset.Status.CurrentRevision

	t.Log("Download last version of CassandraCluster")
	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e",
		Namespace: namespace}, cc); err != nil {
		t.Fatalf("Could not download last version of CassandraCluster: %v", err)
	}

	t.Log("Trigger rolling restart of 1st DC by updating RollingRestart flag")

	cc.Spec.Topology.DC[0].Rack[0].RollingRestart = true

	if err := f.Client.Update(goctx.TODO(), cc); err != nil {
		t.Fatalf("Could not update CassandraCluster: %v", err)
	}

	if err := mye2eutil.WaitForStatefulset(t, f.KubeClient, namespace, "cassandra-e2e-dc1-rack1", 1,
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("WaitForStatefulset got an error: %v", err)
	}

	if err := mye2eutil.WaitForStatusDone(t, f, namespace, "cassandra-e2e",
		mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("WaitForStatusDone got an error: %v", err)
	}

	t.Log("Download statefulset and check current revision has been updated")
	statefulset, err = f.KubeClient.AppsV1().StatefulSets(namespace).Get(goctx.TODO(),
		"cassandra-e2e-dc1-rack1", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Could not download statefulset: %v", err)
	}

	assert.NotEqual(t, stfsVersion, statefulset.Status.CurrentRevision)

	t.Logf("Current labels on statefulset : %v", statefulset.Spec.Template.Labels)

	_, rollingRestartLabelExists := statefulset.Spec.Template.Labels["rolling-restart"]

	t.Log("Assert that rolling-restart label has been added to statefulset")
	assert.Equal(t, true, rollingRestartLabelExists)

	t.Log("Download last version of CassandraCluster and check RollingRestart flag has been cleaned out")

	// Delete Topology in order to get it updated as a whole
	cc.Spec.Topology.DC = nil

	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cassandra-e2e",
		Namespace: namespace}, cc); err != nil {
		t.Fatalf("Could not get CassandraCluster: %v", err)
	}

	assert.Equal(t, false, cc.Spec.Topology.DC[0].Rack[0].RollingRestart)
}

func cassandraClusterServiceTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetWatchNamespace()
	clusterName := "cassandra-e2e"
	kind := "CassandraCluster"

	logrus.Debugf("Creating cluster")
	cluster := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-2DC.yaml", namespace)

	if err := f.Client.Create(goctx.TODO(), cluster, &framework.CleanupOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("Error Creating CassandraCluster: %v", err)
	}

	waitForClusterToBeReady(cluster, f, t)

	cluster = getCassandraCluster(clusterName, namespace, f, t)

	services, err := listServices(namespace, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(map[string]string{
			"app":              "cassandracluster",
			"cassandracluster": cluster.Name,
		}),
	}, f)
	if err != nil {
		t.Errorf("Error listing services: %v", err)
	}
	assert.Equal(t, 2, len(services.Items))

	var clusterService, monitoringService v1.Service
	if val, ok := services.Items[0].Labels["k8s-app"]; ok {
		assert.Equal(t, "exporter-cassandra-jmx", val)
		clusterService = services.Items[0]
		monitoringService = services.Items[1]
	} else if val, ok := services.Items[1].Labels["k8s-app"]; ok {
		assert.Equal(t, "exporter-cassandra-jmx", val)
		clusterService = services.Items[0]
		monitoringService = services.Items[1]
	} else {
		t.Errorf("Failed to find monitoring service. Found: %v", services.Items[0])
	}

	assert.Equal(t, 1, len(clusterService.ObjectMeta.OwnerReferences))
	assert.Equal(t, kind, clusterService.ObjectMeta.OwnerReferences[0].Kind)
	assert.Equal(t, cluster.Name, clusterService.ObjectMeta.OwnerReferences[0].Name)
	assert.True(t, *clusterService.ObjectMeta.OwnerReferences[0].Controller)
	assert.Equal(t, 1, len(clusterService.Spec.Ports))

	statefulSets := getStatefulSets(cluster, f, t)
	checkResourcesConfiguration(t, statefulSets[0].Spec.Template.Spec.Containers, "1", "2Gi", "2", "3Gi")
	checkResourcesConfiguration(t, statefulSets[1].Spec.Template.Spec.Containers, "500m", "1Gi", "500m", "1Gi")

	assertServiceExposesPort(t, &clusterService, "cql", 9042)

	assert.Equal(t, 1, len(monitoringService.ObjectMeta.OwnerReferences))
	assert.Equal(t, kind, monitoringService.ObjectMeta.OwnerReferences[0].Kind)
	assert.Equal(t, cluster.Name, monitoringService.ObjectMeta.OwnerReferences[0].Name)
	assert.True(t, *monitoringService.ObjectMeta.OwnerReferences[0].Controller)
	assert.Equal(t, 1, len(monitoringService.Spec.Ports))

	assertServiceExposesPort(t, &monitoringService, "promjmx", 9500)
}


func checkResourcesConfiguration(t *testing.T, containers []v1.Container, cpuRequested string, memoryRequested string, cpuLimit string, memoryLimit string) {
	for _, c := range containers {
		if c.Name == "cassandra" {
			assert.Equal(t, resource.MustParse(cpuRequested), *c.Resources.Requests.Cpu())
			assert.Equal(t, resource.MustParse(memoryRequested), *c.Resources.Requests.Memory())
			assert.Equal(t, resource.MustParse(cpuLimit), *c.Resources.Limits.Cpu())
			assert.Equal(t, resource.MustParse(memoryLimit),  *c.Resources.Limits.Memory())
		}
	}
}

func cassandraClusterUpdateConfigMapTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatalf("Could not get namespace: %v", err)
	}

	logrus.Debug("Initializing ConfigMaps")
	mye2eutil.HelperInitCassandraConfigMap(t, f, ctx, "cassandra-configmap-v1", namespace)
	mye2eutil.HelperInitCassandraConfigMap(t, f, ctx, "cassandra-configmap-v2", namespace)

	cluster := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC.yaml", namespace)

	logrus.Debugf("Creating cluster")
	if err := f.Client.Create(goctx.TODO(), cluster, &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       mye2eutil.CleanupTimeout,
		RetryInterval: mye2eutil.CleanupRetryInterval}); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("Error Creating CassandraCluster: %v", err)
	}

	waitForClusterToBeReady(cluster, f, t)

	cluster = getCassandraCluster(cluster.Name, cluster.Namespace, f, t)
	statefulSets := getStatefulSets(cluster, f, t)

	cluster.Spec.ConfigMapName = "cassandra-configmap-v2"

	logrus.Debugf("Updating cluster.Spec.ConfigMapName to %s", cluster.Spec.ConfigMapName)
	if err := f.Client.Update(goctx.TODO(), cluster); err != nil {
		t.Fatalf("Could not update CassandraCluster: %v", err)
	}

	waitForClusterToBeReady(cluster, f, t)
	var newCurrentRevision string
	updatedStatefulSets := getStatefulSets(cluster, f, t)
	logrus.Infof("Updated Stateful Sets: %v\n", updatedStatefulSets)

	for i, updatedStatefulSet := range updatedStatefulSets {
		if newCurrentRevision == "" {
			newCurrentRevision = updatedStatefulSet.Status.CurrentRevision
		} else {
			if updatedStatefulSet.Status.CurrentRevision != newCurrentRevision {
				t.Fatalf("Expected CurrentRevion to be the same for all StatefulSets. Expected %s but found %s",
					newCurrentRevision, updatedStatefulSet.Status.CurrentRevision)
			}
			assert.NotEqual(t, updatedStatefulSet.Status.CurrentRevision, statefulSets[i].Status.CurrentRevision,
				"Expected StatefulSet.Status.CurrentRevision to be updated")
		}
	}

	updatedCluster := getCassandraCluster(cluster.Name, cluster.Namespace, f, t)

	assert.Equal(t, cluster.Spec.ConfigMapName, updatedCluster.Spec.ConfigMapName)
}

func cassandraClusterCleanupTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	namespace, _ := ctx.GetWatchNamespace()
	clusterName := "cassandra-e2e"

	logrus.Debugf("Creating cluster")

	cluster := mye2eutil.HelperInitCluster(t, f, ctx, "cassandracluster-1DC-no-autopilot.yaml", namespace)

	if err := f.Client.Create(goctx.TODO(), cluster, &framework.CleanupOptions{}); err != nil && !apierrors.IsAlreadyExists(err) {
		t.Fatalf("Error Creating CassandraCluster: %v", err)
	}

	waitForClusterToBeReady(cluster, f, t)

	cluster = getCassandraCluster(clusterName, namespace, f, t)

	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      "statefulset.kubernetes.io/pod-name",
				Operator: metav1.LabelSelectorOpIn,
				Values:   []string{"cassandra-e2e-dc1-rack1-0", "cassandra-e2e-dc1-rack2-0"},
			},
		},
	}
	opts := metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&selector),
	}

	pods := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}

	pods, err := f.KubeClient.CoreV1().Pods(namespace).List(goctx.TODO(), opts)
	if err != nil {
		t.Fatalf("Failed to list pods: %s", err)
	}
	assert.Equal(t, 2, len(pods.Items))

	for _, pod := range pods.Items {
		pod.Labels["operation-name"] = "cleanup"
		pod.Labels["operation-status"] = "ToDo"
		if err := f.Client.Update(goctx.TODO(), &pod); err != nil {
			t.Fatalf("pod update failed: %s", err)
		}
	}

	checkCleanupExecuted := func(rack string, node int, cc *api.CassandraCluster) (bool, error) {
		dcRack := "dc1-" + rack
		nodeName := fmt.Sprintf("%s-%s-%d", clusterName, dcRack, node)

		dcRackStatus, found := cc.Status.CassandraRackStatus[dcRack]
		_, found = cc.Status.CassandraRackStatus[dcRack]
		if !found {
			return false, fmt.Errorf("Did not find rack status for %s", rack)
		}

		if len(dcRackStatus.PodLastOperation.PodsOK) == 0 {
			// The operation has not completed yet
			t.Logf("The cleanup operation has not yet finished for %s", nodeName)
			return false, nil
		}

		if len(dcRackStatus.PodLastOperation.PodsOK) > 1 {
			// We only scheduled cleanup on one C* node in each rack, so PodsOK should have a length of 1.
			return false, fmt.Errorf("expected cleanup to run on one pod it ran on %d. dcRackStatus.PodLastOperation (+%v)",
				len(dcRackStatus.PodLastOperation.PodsOK), dcRackStatus.PodLastOperation)
		}
		if dcRackStatus.PodLastOperation.PodsOK[0] != nodeName {
			// Make sure the operation executed against the expected node
			return false, fmt.Errorf("expected cleanup to run on %s but it ran on %s", nodeName,
				dcRackStatus.PodLastOperation.PodsOK[0])
		}

		pod, err := f.KubeClient.CoreV1().Pods(namespace).Get(goctx.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			t.Logf("Failed to get pod %s: %s", nodeName, err)
			return false, nil
		}

		val, exists := pod.Labels["operation-name"]
		if !exists {
			t.Logf("Expected to find label operation-name on %s", nodeName)
			return false, nil
		}
		if val != "cleanup" {
			t.Logf("Expected to find label operation-name=cleanup on %s but found operation-name=%s",
				nodeName, val)
			return false, nil
		}

		val, exists = pod.Labels["operation-status"]
		if !exists {
			t.Logf("Expected to find label operation-status on %s", nodeName)
			return false, nil
		}
		if val != "Done" {
			t.Logf("Expected to find label operation-status=Done on %s but found operation-status=%s",
				nodeName, val)
			return false, nil
		}

		startTimeLabel, exists := pod.Labels["operation-start"]
		if !exists {
			t.Logf("Expected to find label operation-start on %s", nodeName)
			return false, nil
		}

		endTimeLabel, exists := pod.Labels["operation-end"]
		if !exists {
			t.Logf("Expected to find label operation-end on %s", nodeName)
			return false, nil
		}

		startTime, err := k8s.LabelTime2Time(startTimeLabel)
		if err != nil {
			t.Logf("Failed to parse operation-start label: %s", err)
			return false, nil
		}

		endTime, err := k8s.LabelTime2Time(endTimeLabel)
		if err != nil {
			t.Logf("Failed to parse operation-end label: %s", err)
			return false, nil
		}

		if endTime.Sub(startTime) < 0 {
			t.Logf("Expected endTime (%s) to be >= startTime (%s)", endTime, startTime)
			return false, nil
		}

		return true, nil
	}

	conditionFunc := func(cc *api.CassandraCluster, rack string, node int) (bool, error) {
		executed, err := checkCleanupExecuted(rack, node, cc)
		if err != nil {
			logrus.Infof("cleanup check failed: %s\n", err)
		}
		return executed, err
	}

	checkRack := func(rack string) func(cc *api.CassandraCluster) (bool, error) {
		return func(cc *api.CassandraCluster) (bool, error) { return conditionFunc(cc, rack, 0) }
	}

	logrus.Infof("Wait for cleanup to finish in rack1\n")
	err = mye2eutil.WaitForStatusChange(t, f, namespace, clusterName, 1*time.Second, 300*time.Second, checkRack("rack1"))
	if err != nil {
		t.Errorf("WaitForStatusChange failed: %s", err)
	}

	logrus.Infof("Wait for cleanup to finish in rack2\n")
	err = mye2eutil.WaitForStatusChange(t, f, namespace, clusterName, 1*time.Second, 300*time.Second, checkRack("rack2"))
	if err != nil {
		t.Errorf("WaitForStatusChange failed: %s", err)
	}
}

func listServices(namespace string, options metav1.ListOptions, f *framework.Framework) (*v1.ServiceList, error) {
	return f.KubeClient.CoreV1().Services(namespace).List(goctx.TODO(), options)
}

func assertServiceExposesPort(t *testing.T, svc *v1.Service, portName string, port int32) {
	if svcPort, err := findServicePort(portName, svc.Spec.Ports); err == nil {
		assert.Equal(t, port, svcPort.Port)
	} else {
		assert.Fail(t, fmt.Sprintf("Failed to find service port: %s", portName))
	}
}

func findServicePort(name string, ports []v1.ServicePort) (*v1.ServicePort, error) {
	for _, port := range ports {
		if port.Name == name {
			return &port, nil
		}
	}
	return nil, fmt.Errorf("Failed to find service port: %s", name)
}

func getStatefulSet(name string, namespace string, f *framework.Framework, t *testing.T) *appsv1.StatefulSet {
	statefulSet, err := f.KubeClient.AppsV1().StatefulSets(namespace).Get(goctx.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get StatefulSet %s: %v", name, err)
	}
	return statefulSet
}

func getStatefulSets(cluster *api.CassandraCluster, f *framework.Framework, t *testing.T) []*appsv1.StatefulSet {
	var statefulSets []*appsv1.StatefulSet
	for _, dc := range cluster.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			name := fmt.Sprintf("%s-%s-%s", cluster.Name, dc.Name, rack.Name)
			statefulSet := getStatefulSet(name, cluster.Namespace, f, t)
			statefulSets = append(statefulSets, statefulSet)
		}
	}
	return statefulSets
}

func getCassandraCluster(name string, namespace string, f *framework.Framework, t *testing.T) *api.CassandraCluster {
	cluster := &api.CassandraCluster{}
	if err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cluster); err != nil {
		t.Fatalf("Failed to get CassandraCluster %s: %v", name, err)
	}
	return cluster
}

func waitForClusterToBeReady(cluster *api.CassandraCluster, f *framework.Framework, t *testing.T) {
	for _, dc := range cluster.Spec.Topology.DC {
		for _, rack := range dc.Rack {
			name := fmt.Sprintf("%s-%s-%s", cluster.Name, dc.Name, rack.Name)
			logrus.Debugf("Waiting for StatefulSet %s", name)
			if err := mye2eutil.WaitForStatefulset(
				t,
				f.KubeClient,
				cluster.Namespace,
				name,
				int(cluster.Spec.NodesPerRacks),
				mye2eutil.RetryInterval,
				mye2eutil.Timeout); err != nil {
				t.Fatalf("Waiting for StatefulSet %s failed: %v", name, err)
			}
		}
	}

	if err := mye2eutil.WaitForStatusDone(t, f, cluster.Namespace, cluster.Name, mye2eutil.RetryInterval, mye2eutil.Timeout); err != nil {
		t.Fatalf("Waiting for cluster status change to Done failed: %v", err)
	}
}
