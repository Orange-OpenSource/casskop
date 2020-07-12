package cassandrabackup

import (
	"fmt"
	"net/http"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	csapi "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
	corev1 "k8s.io/api/core/v1"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	PerformRestoreOperation(restoreOperation csapi.RestoreOperationRequest) (*csapi.RestoreOperationResponse, error)
	GetRestoreOperation(operationId string) (*csapi.RestoreOperationResponse, error)
	PerformBackupOperation(request csapi.BackupOperationRequest) (*csapi.BackupOperationResponse, error)
	GetBackupOperation(id string) (response *csapi.BackupOperationResponse, err error)
	Build() error
}

type client struct {
	Client
	opts *Config
	podClient *csapi.APIClient

	newClient func(*csapi.Configuration) *csapi.APIClient
}

func New(opts *Config) Client {
	csClient := &client{
		opts: opts,
	}

	csClient.newClient = csapi.NewAPIClient
	return csClient
}

func (cs *client) Build() error {
	cs.podClient = cs.newClient( cs.getCassandraBackupPodSidecarConfig())
	return nil
}

func NewFromCluster(k8sclient controllerclient.Client, cluster *api.CassandraCluster, pod *corev1.Pod) (Client, error) {
	var client Client
	var err error

	opts := ClusterConfig(k8sclient, cluster, pod)

	client = New(opts)
	err = client.Build()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (cs *client) getCassandraBackupPodSidecarConfig() (config *csapi.Configuration) {
	config = csapi.NewConfiguration()

	protocol := "http"

	if cs.opts.UseSSL {
		config.Scheme = "HTTPS"
		transport := &http.Transport{TLSClientConfig: cs.opts.TLSConfig}
		config.HTTPClient = &http.Client{Transport: transport, Timeout: cs.opts.Timeout}
		protocol = "https"
	}

	config.BasePath = fmt.Sprintf("%s://%s:%d", protocol,cs.opts.Host, cs.opts.Port)
	//config.Host = cs.opts.Host

	return config
}