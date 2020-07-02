package sidecar

import (
	"fmt"
	"net/http"
	"testing"

	csd "github.com/cscetbon/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

var host, port = "127.0.0.1", ":4567"

func TestDemarshalling(t *testing.T) {

	client := NewSidecarClient(host, &DefaultSidecarClientOptions)
	httpmock.ActivateNonDefault(client.Config.HTTPClient)

	defer httpmock.DeactivateAndReset()

	operationID := "d3262073-8101-450f-9a11-c851760abd57"
	url := fmt.Sprintf("http://%s%s/%s/%s", host, port, "operations", operationID)

	httpmock.RegisterResponder(http.MethodGet, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, map[string]interface{}{
				"type":     "backup",
				"id":       operationID,
				"duration": "PT1M0.613S",
				"start":    "2019-06-11T03:37:15.593Z",
				"stop":     "2019-06-11T03:38:16.206Z",
				"state":    "RUNNING",
				"progress": 10.0,
			})
		})

	response, err := client.GetOperation(operationID)

	assert := assert.New(t)
	assert.Nil(err)
	assert.NotNil(response)
	assert.Equal(response.State, "RUNNING")
	assert.Equal(response.Progress, 10.0)
}

func TestClientBackupNode(t *testing.T) {

	client := NewSidecarClient(host, &DefaultSidecarClientOptions)
	httpmock.ActivateNonDefault(client.Config.HTTPClient)

	defer httpmock.DeactivateAndReset()

	operationID := "d3262073-8101-450f-9a11-c851760abd57"
	url := fmt.Sprintf("http://%s%s/%s", host, port, "operations")

	request := csd.BackupOperationRequest{
		ConcurrentConnections: 10,
		GlobalRequest:         true,
		K8sNamespace:          "default",
		K8sSecretName:         "cloud-backup-secrets",
		SnapshotTag:           "SnapshotTag1",
		StorageLocation:       "s3://cassie/test-cluster/dc1/cassandra-test-cluster-dc1-rack1-0",
	}

	httpmock.RegisterResponder(http.MethodPost, url,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewJsonResponse(200, map[string]interface{}{
				"cassandraDirectory":    "/var/lib/cassandra",
				"concurrentConnections": request.ConcurrentConnections,
				"creationTime":          "2020-06-10T04:53:05.976Z",
				"entities":              "",
				"globalRequest":         request.GlobalRequest,
				"id":                    operationID,
				"k8sNamespace":          request.K8sNamespace,
				"k8sSecretName":         request.K8sSecretName,
				"progress":              0.0,
				"snapshotTag":           request.SnapshotTag,
				"startTime":             "2020-06-10T04:53:05.985Z",
				"state":                 "RUNNING",
				"storageLocation":       request.StorageLocation,
				"type":                  "backup",
			})
		})

	id, err := client.StartOperation(request)

	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal(id, operationID)

}
