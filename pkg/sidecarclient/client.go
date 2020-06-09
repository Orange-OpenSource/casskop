package sidecarclient

import (
	"fmt"
	"net/http"

	csapi "github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("cassandrasidecar-client")

type CassandraSidecarClient interface {

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
		cs.podsClient[pod.Spec.Hostname] = cs.newClient(podConfig)
	}
	return nil
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

	config.BasePath = fmt.Sprintf("%s://%s%d", protocol, pod.Status.PodIP, cs.opts.Port)
	config.Host = pod.Spec.Hostname

	return config
}
