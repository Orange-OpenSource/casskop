// Copyright 2019 Orange
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// 	See the License for the specific language governing permissions and
// limitations under the License.

// Provide K8S Client to be able to operate directly with K8S. For example to do exec cmd on a pod.
// Usage sample :
//   clientset, cfg := k8s.MustNewKubeClientAndConfig()
//	 sdtout, stderr,err := k8s.ExecPodFromName(clientset.(*kubernetes.Clientset),cfg,capi.Namespace,nodename,killCmd)
//	 if err!=nil {
//		logrus.Errorf("Error when run cmd to pod %s", nodename)
//		return fmt.Errorf("failed execute command on pods: %v", err)
//	 }
package k8s

import (
	"bytes"
	"fmt"
	"net"
	"os"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

//var clientset *kubernetes.Clientset
var clientset kubernetes.Interface
var cfg *rest.Config

//InitClient allow to setup an additional client to kubernetes API while operator-sdk don't gives us access to oit
func InitClient() {
	if clientset == nil {
		clientset, cfg = MustNewKubeClientAndConfig()
	}
}

// Copy from k8sclient because not yet public
// MustNewKubeClientAndConfig returns the in-cluster config and kubernetes client
// or if KUBERNETES_CONFIG is given an out of cluster config and client
func MustNewKubeClientAndConfig() (kubernetes.Interface, *rest.Config) {
	//var cfg *rest.Config
	var err error
	if os.Getenv(k8sutil.KubeConfigEnvVar) != "" {
		cfg, err = outOfClusterConfig()
	} else {
		cfg, err = inClusterConfig()
	}
	if err != nil {
		panic(err)
	}
	return kubernetes.NewForConfigOrDie(cfg), cfg
}

// inClusterConfig returns the in-cluster config accessible inside a pod
func inClusterConfig() (*rest.Config, error) {
	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// See https://github.com/coreos/etcd-operator/issues/731#issuecomment-283804819
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		addrs, err := net.LookupHost("kubernetes.default.svc")
		if err != nil {
			return nil, err
		}
		os.Setenv("KUBERNETES_SERVICE_HOST", addrs[0])
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", "443")
	}
	return rest.InClusterConfig()
}

func outOfClusterConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv(k8sutil.KubeConfigEnvVar)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	return config, err
}

//inspiration
//https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/exec.go
//func ExecPodFromName(clientset *kubernetes.Clientset, cfg *rest.Config, namespace string, name string, cmd []string) (string, string, error) {
func ExecPodFromName(namespace string, name string, cmd []string) (string, string, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", "", fmt.Errorf("could not get pod info: %v", err)
	}

	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return "", "", fmt.Errorf("cannot exec into a container in a completed pod; current phase is %s", pod.Status.Phase)
	}

	//	return ExecPod(clientset, cfg, namespace, pod, cmd)
	return ExecPod(namespace, pod, cmd)

}

//inspiration
//https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/cmd/exec.go
//func ExecPod(clientset *kubernetes.Clientset, cfg *rest.Config, namespace string, pod *corev1.Pod, cmd []string) (string, string, error) {
func ExecPod(namespace string, pod *corev1.Pod, cmd []string) (string, string, error) {

	if len(pod.Spec.Containers) != 1 {
		return "", "", fmt.Errorf("could not determine which container to use")
	}

	// build the remoteexec
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: pod.Spec.Containers[0].Name,
		Command:   cmd,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
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
