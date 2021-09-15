package cassandrabackup

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Orange-OpenSource/casskop/pkg/apis/db/v2/common"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)



func TestPerformRestoreOperation(t *testing.T) {
	assert := assert.New(t)

	restore, err := performRestoreMock(201)
	assert.Nil(err)
	assert.NotNil(restore)

	restore, err = performRestoreMock(500)
	assert.NotNil(err)
	assert.Nil(restore)

}

func TestGetRestoreOperation(t *testing.T) {
	assert := assert.New(t)

	restore, err := getRestoreMock(200)
	assert.Nil(err)
	assert.NotNil(restore)

	restore, err = getRestoreMock(500)
	assert.NotNil(err)
	assert.Nil(restore)
}


func performRestoreMock(codeStatus int) (*icarus.RestoreOperationResponse, error) {
	client := newBuildedMockClient()
	defer httpmock.DeactivateAndReset()

	sourceDir         := "/var/lib/cassandra/downloadedsstables"

	restoreOperationRequest := icarus.RestoreOperationRequest {
		Type_: "restore",
		StorageLocation: storageLocation,
		SnapshotTag: snapshotTag,
		NoDeleteDownloads: noDeleteDownloads,
		SchemaVersion: schemaVersion,
		ExactSchemaVersion: false,
		RestorationPhase: "DOWNLOAD",
		GlobalRequest: true,
		Import_: &icarus.AllOfRestoreOperationRequestImport_{
			Type_: "import",
			SourceDir: sourceDir,
		},
		K8sSecretName: k8sSecretName,
		ConcurrentConnections: concurrentConnections,
	}

	url := fmt.Sprintf("http://%s:%d/operations", hostnamePodA, DefaultCassandraSidecarPort)

	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				codeStatus,
				common.MockRestoreResponse(
					restoreOperationRequest.NoDeleteDownloads,
					restoreOperationRequest.ConcurrentConnections,
					state,
					restoreOperationRequest.SnapshotTag,
					operationID,
					restoreOperationRequest.K8sSecretName,
					restoreOperationRequest.StorageLocation,
					"INIT",
					restoreOperationRequest.SchemaVersion))
		})


	return client.PerformRestoreOperation(restoreOperationRequest)
}

func getRestoreMock(codeStatus int) (*icarus.RestoreOperationResponse, error) {
	client := newBuildedMockClient()
	defer httpmock.DeactivateAndReset()

	url := fmt.Sprintf("http://%s:%d/operations/%s", hostnamePodA, DefaultCassandraSidecarPort, operationID)

	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(
				codeStatus,
				common.MockRestoreResponse(
					noDeleteDownloads,
					concurrentConnections,
					state,
					snapshotTag,
					operationID,
					k8sSecretName,
					storageLocation,
					"INIT",
					schemaVersion))
		})

	return client.RestoreOperationByID(operationID)
}

