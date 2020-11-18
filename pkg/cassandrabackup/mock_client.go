package cassandrabackup

import (
	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1/common"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	"github.com/jarcoal/httpmock"
	"github.com/mitchellh/mapstructure"
)

var (
	hostnamePodA = "podA.ns.cluster.svc.cluster.local"
)

const (
	state             = "PENDING"
	stateGetById      = "RUNNING"
	operationID       = "d3262073-8101-450f-9a11-c851760abd57"
	k8sSecretName     = "cloud-backup-secrets"
	snapshotTag       = "SnapshotTag1"
	storageLocation   = "gcp://bucket/clustername/dcname/nodename"
	noDeleteDownloads = false
	schemaVersion     = "test"
	concurrentConnections int32 = 15
)

type mockCassandraBackupClient struct {
	Client
	opts *Config
	podClient *icarus.APIClient

	newClient func(*icarus.Configuration) *icarus.APIClient
	failOpts bool
}

func newMockOpts() *Config {
	return &Config{
		UseSSL: DefaultCassandraBackupSecure,
		Port:   DefaultCassandraSidecarPort,
		Host:   hostnamePodA,
	}
}

func newMockHttpClient(c *icarus.Configuration) *icarus.APIClient {
	client := icarus.NewAPIClient(c)
	httpmock.Activate()
	return client
}

func newMockClient() *client {
	return &client{
		config:    newMockOpts(),
		newClient: newMockHttpClient,
	}
}


func newBuildedMockClient() *client {
	client := newMockClient()
	client.Build()
	return client
}


func NewMockCassandraBackupClient() *mockCassandraBackupClient {
	return &mockCassandraBackupClient{
		opts:       newMockOpts(),
		newClient:  newMockHttpClient,
	}
}

func NewMockCassandraBackupClientFailOps() *mockCassandraBackupClient {
	return &mockCassandraBackupClient{
		opts:      newMockOpts(),
		newClient: newMockHttpClient,
		failOpts:  true,
	}
}

func (m *mockCassandraBackupClient) PerformRestoreOperation(restoreOperation icarus.RestoreOperationRequest) (*icarus.RestoreOperationResponse, error) {
	if m.failOpts {
		return nil, ErrCassandraSidecarNotReturned201
	}

	var restoreOp icarus.RestoreOperationResponse

	mapstructure.Decode(common.MockRestoreResponse(
		restoreOperation.NoDeleteDownloads,
		restoreOperation.ConcurrentConnections,
		state,
		restoreOperation.SnapshotTag,
		operationID,
		restoreOperation.K8sSecretName,
		restoreOperation.StorageLocation,
		restoreOperation.RestorationStrategyType,
		restoreOperation.RestorationPhase,
		restoreOperation.SchemaVersion), &restoreOp)

	return &restoreOp, nil
}

func (m *mockCassandraBackupClient) RestoreOperationByID(operationId string) (*icarus.RestoreOperationResponse, error) {
	if m.failOpts {
		return nil, ErrCassandraSidecarNotReturned200
	}

	var restoreOperation icarus.RestoreOperationResponse

	mapstructure.Decode(common.MockRestoreResponse(
		noDeleteDownloads,
		concurrentConnections,
		stateGetById,
		snapshotTag,
		operationId,
		k8sSecretName,
		storageLocation,
		"HARDLINKS",
		"TRUNCATE",
		schemaVersion), &restoreOperation)
	return &restoreOperation, nil
}

