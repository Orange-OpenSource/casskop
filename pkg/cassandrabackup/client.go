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
	RestoreOperationByID(operationId string) (*csapi.RestoreOperationResponse, error)
	PerformBackupOperation(request csapi.BackupOperationRequest) (*csapi.BackupOperationResponse, error)
	BackupOperationByID(id string) (response *csapi.BackupOperationResponse, err error)
	Build() error
}

type client struct {
	Client
	config    *Config
	podClient *csapi.APIClient

	newClient func(*csapi.Configuration) *csapi.APIClient
}

func New(config *Config) Client {
	return &client{config: config, newClient: csapi.NewAPIClient}
}

func (cs *client) Build() error {
	cs.podClient = cs.newClient( cs.cassandraBackupSidecarConfig())
	return nil
}

func ClientFromCluster(k8sClient controllerclient.Client, cluster *api.CassandraCluster,
	pod *corev1.Pod) (Client, error) {
	config := ClusterConfig(k8sClient, cluster, pod)

	client := New(config)
	err := client.Build()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (cs *client) cassandraBackupSidecarConfig() (config *csapi.Configuration) {
	config = csapi.NewConfiguration()

	protocol := "http"

	if cs.config.UseSSL {
		config.Scheme = "HTTPS"
		transport := &http.Transport{TLSClientConfig: cs.config.TLSConfig}
		config.HTTPClient = &http.Client{Transport: transport, Timeout: cs.config.Timeout}
		protocol = "https"
	}

	config.BasePath = fmt.Sprintf("%s://%s:%d", protocol,cs.config.Host, cs.config.Port)
	//config.Host = cs.config.Host

	return config
}