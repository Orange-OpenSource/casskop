package v1alpha1

import (
	"testing"

	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1/common"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
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

func TestComputeStatusFromRestoreOperation(t *testing.T) {
	assert := assert.New(t)

	var restoreOperation icarus.RestoreOperationResponse

	mapstructure.Decode(common.MockRestoreResponse(
		noDeleteDownloads,
		concurrentConnections,
		stateGetById,
		snapshotTag,
		operationID,
		k8sSecretName,
		storageLocation,
		"TRUNCATE",
		schemaVersion), &restoreOperation)

	cs := ComputeRestorationStatus(&restoreOperation)
	assert.Equal(BackRestStatus{
		TimeCreated:   "2020-06-10T04:53:05.976Z",
		TimeStarted:   "2020-06-10T05:53:05.976Z",
		TimeCompleted: "2020-06-10T06:53:05.976Z",
		Condition:     &BackRestCondition{
			Type: string(RestoreRunning),
			LastTransitionTime: cs.Condition.LastTransitionTime,
		},
		Progress:      "10%",
		ID:            operationID,
	}, cs)
}