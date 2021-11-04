package cassandrabackup

import (
	"fmt"
	"net/http"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v2"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	corev1 "k8s.io/api/core/v1"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Client interface {
	PerformRestoreOperation(restoreOperation icarus.RestoreOperationRequest) (*icarus.RestoreOperationResponse, error)
	RestoreOperationByID(operationId string) (*icarus.RestoreOperationResponse, error)
	PerformBackupOperation(request icarus.BackupOperationRequest) (*icarus.BackupOperationResponse, error)
	BackupOperationByID(id string) (response *icarus.BackupOperationResponse, err error)
	Build() error
}

type client struct {
	Client
	config    *Config
	podClient *icarus.APIClient

	newClient func(*icarus.Configuration) *icarus.APIClient
}

func New(config *Config) Client {
	return &client{config: config, newClient: icarus.NewAPIClient}
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

func (cs *client) cassandraBackupSidecarConfig() (config *icarus.Configuration) {
	config = icarus.NewConfiguration()

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