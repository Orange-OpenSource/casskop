package sidecarclient

import (
	"fmt"
	"net/http"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	csapi "github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("cassandrasidecar-client")

type CassandraSidecarClient interface {
	PerformRestoreOperation(podName string, restoreOperation csapi.RestoreOperationRequest) (*csapi.RestoreOperationResponse, error)
	Build() error
}

type cassandraSidecarClient struct {
	CassandraSidecarClient
	opts *CassandraSidecarConfig
	podsClient map[string]*csapi.APIClient

	newClient func(*csapi.Configuration) *csapi.APIClient
}

func New(opts *CassandraSidecarConfig) CassandraSidecarClient {
	csClient := &cassandraSidecarClient{
		opts: opts,
	}

	csClient.newClient = csapi.NewAPIClient
	return csClient
}

func (cs *cassandraSidecarClient) Build() error {

	cs.podsClient = make(map[string]*csapi.APIClient)
	for _, pod := range cs.opts.Pods {
		podConfig := cs.getCassandraPodSidecarConfig(&pod)
		cs.podsClient[pod.Name] = cs.newClient(podConfig)
	}
	return nil
}

func NewFromCluster(k8sclient client.Client, cluster *api.CassandraCluster, podList *corev1.PodList) (CassandraSidecarClient, error) {
	var client CassandraSidecarClient
	var err error

	opts := ClusterConfig(k8sclient, cluster, podList)

	client = New(opts)
	err = client.Build()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (cs *cassandraSidecarClient) getCassandraPodSidecarConfig(pod *corev1.Pod) (config *csapi.Configuration) {
	config = csapi.NewConfiguration()

	protocol := "http"

	if cs.opts.UseSSL {
		config.Scheme = "HTTPS"
		cs.opts.TLSConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: cs.opts.TLSConfig}
		config.HTTPClient = &http.Client{Transport: transport}
		protocol = "https"
	}

	config.BasePath = fmt.Sprintf("%s://%s:%d", protocol, pod.Status.PodIP, cs.opts.Port)
	config.Host = pod.Spec.Hostname

	return config
}