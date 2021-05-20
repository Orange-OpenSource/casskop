package e2eutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"

	goctx "context"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	)

var (
	RetryInterval        = time.Second * 10
	Timeout              = time.Second * 600
	CleanupRetryInterval = time.Second * 5
	CleanupTimeout       = time.Second * 20
)

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

//HelperInitCluster goal is to create objects from the testdata/file.yaml pointed by name.
//for now we can create Secret or CassandraCluster, we may add more objects in futur if needed
func HelperInitCluster(t *testing.T, f *framework.Framework, ctx *framework.Context,
	name, namespace string) *api.CassandraCluster {
	var cc *api.CassandraCluster

	fileR := helperLoadBytes(t, name)

	fileAsString := string(fileR[:])
	sepYamlfiles := strings.Split(fileAsString, "---")
	for _, file := range sepYamlfiles {
		if file == "\n" || file == "" {
			// ignore empty cases
			continue
		}

		decode := serializer.NewCodecFactory(f.Scheme).UniversalDeserializer().Decode
		obj, groupVersionKind, err := decode([]byte(file), nil, nil)
		if err != nil {
			log.Println(fmt.Sprintf("Error while decoding YAML object. Err was: %s, %s", err, groupVersionKind))
			continue
		}
		switch o := obj.(type) {

		//for CassandraCluster we return the object and create it after (can be modified before upgrading)
		case *api.CassandraCluster:
			o.Namespace = namespace
			cc = o
		case *corev1.Secret:
			o.Namespace = namespace
			if err := f.Client.Create(goctx.TODO(), o, &framework.CleanupOptions{TestContext: ctx,
				Timeout:       CleanupTimeout,
				RetryInterval: CleanupRetryInterval}); err != nil && !apierrors.IsAlreadyExists(err) {
				t.Fatalf("Error Creating cassandracluster: %v", err)
			}
		}

	}

	return cc
}

func HelperInitOperator(t *testing.T) (*framework.Context, *framework.Framework) {
	//Comment the line below if we want to have sequential tests
	t.Parallel()
	ctx := framework.NewContext(t)

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: CleanupTimeout,
		RetryInterval: CleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetOperatorNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for casskop to be ready
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "casskop", 1, RetryInterval, Timeout)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, f

}

func WaitForStatefulset(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int,
	retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		statefulset, err := kubeclient.AppsV1().StatefulSets(namespace).Get(goctx.TODO(), name,
			metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s statefulset\n", name)
				return false, nil
			}
			return false, err
		}

		if int(statefulset.Status.ReadyReplicas) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s statefulset (%d/%d)\n", name,
			statefulset.Status.ReadyReplicas, replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Statefulset available (%d/%d)\n", replicas, replicas)
	return nil
}

type CassandraClusterConditionFunc func(cc *api.CassandraCluster) (bool, error)

// WaitForStatusChange tries a condition func until it returns true, an error, or
// the timeout is reached. Each invocation of conditionFunc will receive a newly
// fetched instance of the CassandraCluster.
func WaitForStatusChange(
	t *testing.T,
	f *framework.Framework,
	namespace string,
	name string,
	retryInterval time.Duration,
	timeout time.Duration,
	conditionFunc CassandraClusterConditionFunc) error {

	cluster := &api.CassandraCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CassandraCluster",
			APIVersion: "db.orange.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	return wait.Poll(retryInterval, timeout, func() (bool, error) {
		err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cluster)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("CassandraCluster %s in namespace %s not found: %s", name, namespace, err)
				return false, nil
			}
			return false, err
		}

		return conditionFunc(cluster)
	})
}

func WaitForStatusDone(t *testing.T, f *framework.Framework, namespace, name string,
	retryInterval, timeout time.Duration) error {

	cc2 := &api.CassandraCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CassandraCluster",
			APIVersion: "db.orange.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := wait.Poll(RetryInterval, Timeout, func() (done bool, err error) {

		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cc2)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s CassandraCluster\n", name)
				return false, nil
			}
			return false, err
		}

		if cc2.Status.LastClusterActionStatus == api.StatusDone {
			return true, nil
		}
		t.Logf("Waiting for full Operator %s to finish Action of %s=%s\n", name,
			cc2.Status.LastClusterAction,
			cc2.Status.LastClusterActionStatus)

		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Operator Status is Done (%s/%s)\n", cc2.Status.LastClusterAction,
		cc2.Status.LastClusterActionStatus)
	return nil
}

func WaitForPodOperationDone(t *testing.T, f *framework.Framework, namespace, name string, dcRackName string,
	retryInterval, timeout time.Duration) error {

	cc2 := &api.CassandraCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CassandraCluster",
			APIVersion: "db.orange.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {

		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cc2)
		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("CassandraCluster not found.. this is not good..\n", name)
				return false, nil
			}
			return false, err
		}

		if cc2.Status.CassandraRackStatus[dcRackName].PodLastOperation.Status == api.StatusDone {
			return true, nil
		}
		t.Logf("Waiting for PodOperation %s to finish\n",
			cc2.Status.CassandraRackStatus[dcRackName].PodLastOperation.Name)

		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("PodOperation is Done (%s/%s)\n", cc2.Status.CassandraRackStatus[dcRackName].PodLastOperation.Name,
		cc2.Status.CassandraRackStatus[dcRackName].PodLastOperation.Status)
	return nil
}

func ExecPodFromName(t *testing.T, f *framework.Framework, namespace string, podName string, cmd string) (string,
	string, error) {
	pod := &v1.Pod{TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"}}

	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: podName, Namespace: namespace}, pod)
	if err != nil {
		t.Logf("Error getting pod: %v", err)
	}

	stdout, stderr, err := ExecPod(f, namespace, pod, []string{"bash", "-c", cmd})
	if err != nil {
		t.Logf("Error exec pod %s = %v", podName, err)
	}
	stdout = strings.TrimSuffix(stdout, "\n")
	return stdout, stderr, err
}

func ExecPod(f *framework.Framework, namespace string, pod *corev1.Pod, cmd []string) (string, string,
	error) {

	// build the remoteexec
	req := f.KubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: "cassandra",
		Command:   cmd,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(f.KubeConfig, "POST", req.URL())
	if err != nil {
		return "", "", fmt.Errorf("could not init remote executor: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	return stdout.String(), stderr.String(), err

}

func HelperInitCassandraConfigMap(t *testing.T, f *framework.Framework, ctx *framework.Context,
	configMapName, namespace string) {
	configMapFile := helperLoadBytes(t, configMapName+".yaml")
	decode := serializer.NewCodecFactory(f.Scheme).UniversalDeserializer().Decode
	configMapString := string(configMapFile[:])
	obj, _, err := decode([]byte(configMapString), nil, nil)

	if err != nil {
		t.Fatalf("Error decoding ConfigMap: %s", err)
	}

	switch cm := obj.(type) {
	case *corev1.ConfigMap:
		cm.Name = configMapName
		cm.Namespace = namespace
		if err := f.Client.Create(goctx.TODO(), cm, &framework.CleanupOptions{TestContext: ctx, Timeout: CleanupTimeout,
			RetryInterval: CleanupRetryInterval}); err != nil && !apierrors.IsAlreadyExists(err) {
			t.Fatalf("Error creating ConfigMap: %v", err)
		}
	default:
		t.Fatalf("Expected a ConfigMap but got a %T", cm)
	}
}

func K8sGetCassandraCluster(t *testing.T, f *framework.Framework, err error, cc *api.CassandraCluster) {
	if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: cc.Name, Namespace: cc.Namespace},
		cc); err != nil {
		t.Fatal(err)
	}
}
