package sidecarclient

import (
	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1/common"
	csapi "github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
	"github.com/jarcoal/httpmock"
	"github.com/mitchellh/mapstructure"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	hostnamePodA = "podA.ns.cluster.svc.cluster.local"
	ipPodA       = "10.110.0.2"
	namePodA     = "podA"
	hostnamePodB = "podB.ns.cluster.svc.cluster.local"
	ipPodB       = "10.110.0.3"
	namePodB     = "podB"
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

type mockCassandraSidecarClient struct {
	CassandraSidecarClient
	opts *CassandraSidecarConfig
	podsClient map[string]*csapi.APIClient

	newClient func(*csapi.Configuration) *csapi.APIClient
	failOpts bool
}

func newMockOpts() *CassandraSidecarConfig {
	return &CassandraSidecarConfig{
		UseSSL: DefaultCassandraSidecarSecure,
		Port: DefaultCassandraSidecarPort,
		Pods: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: namePodA,
				},
				Spec: corev1.PodSpec{
					Hostname: hostnamePodA,
				},
				Status: corev1.PodStatus{
					PodIP: ipPodA,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: namePodB,
				},
				Spec: corev1.PodSpec{
					Hostname: hostnamePodB,
				},
				Status: corev1.PodStatus{
					PodIP: ipPodB,
				},
			},
		},
	}
}

func newMockHttpClient(c *csapi.Configuration) *csapi.APIClient {
	client := csapi.NewAPIClient(c)
	httpmock.Activate()
	return client
}

func newMockClient() *cassandraSidecarClient {
	return &cassandraSidecarClient{
		opts:       newMockOpts(),
		newClient:  newMockHttpClient,
	}
}


func newBuildedMockClient() *cassandraSidecarClient {
	client := newMockClient()
	client.Build()
	return client
}


func NewMockCassandraSidecarClient() *mockCassandraSidecarClient {
	return &mockCassandraSidecarClient{
		opts:       newMockOpts(),
		newClient:  newMockHttpClient,
	}
}

func NewMockCassandraSidecarClientFailOps() *mockCassandraSidecarClient {
	return &mockCassandraSidecarClient{
		opts:      newMockOpts(),
		newClient: newMockHttpClient,
		failOpts:  true,
	}
}

func (m *mockCassandraSidecarClient) PerformRestoreOperation(podName string, restoreOperation csapi.RestoreOperationRequest) (*csapi.RestoreOperationResponse, error) {
	if m.failOpts {
		return nil, ErrCassandraSidecarNotReturned201
	}

	var restoreOp csapi.RestoreOperationResponse

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

func (m *mockCassandraSidecarClient) GetRestoreOperation(podName, operationId string) (*csapi.RestoreOperationResponse, error) {
	if m.failOpts {
		return nil, ErrCassandraSidecarNotReturned200
	}

	var restoreOperation csapi.RestoreOperationResponse

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