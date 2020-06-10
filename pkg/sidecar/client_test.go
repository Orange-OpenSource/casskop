package sidecar

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/Orange-OpenSource/casskop/pkg/common/nodestate"
	"github.com/Orange-OpenSource/casskop/pkg/common/operations"
	"github.com/google/uuid"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func getHost() string {

	host := os.Getenv("TEST_SIDECAR_HOST")

	if len(host) == 0 {
		return "127.0.0.1"
	}
	return host
}

func TestDemarshalling(t *testing.T) {

	client := NewSidecarClient(getHost(), &DefaultSidecarClientOptions)

	httpmock.ActivateNonDefault(client.restyClient.GetClient())

	defer httpmock.DeactivateAndReset()

	// Check Status()
	httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s/%s", client.restyClient.HostURL, EndpointStatus),
		httpmock.NewStringResponder(200, `{"nodeState": "NORMAL"}`))

	if status, err := client.Status(); err != nil || status == nil || status.NodeState != nodestate.NORMAL {
		t.Fail()
	}

	// Check GetOperations() and FilterOperations()
	httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s/%s", client.restyClient.HostURL, EndpointOperations),
		httpmock.NewStringResponder(200,
			`[
				{
					"type": "decommission",
					"id": "d3262073-8101-450f-9a11-c851760abd57",
					"duration": "PT1M0.613S",
					"start": "2019-06-11T03:37:15.593Z",
					"stop": "2019-06-11T03:38:16.206Z",
					"state": "FINISHED"
				},
				{
					"type": "backup",
					"id": "d3262073-8101-450f-9a11-c851760abd57",
					"duration": "PT1M0.613S",
					"start": "2019-06-11T03:37:15.593Z",
					"stop": "2019-06-11T03:38:16.206Z",
					"state": "RUNNING"
				}
			]`))

	ops, err := client.GetOperations()
	if err != nil {
		t.Error(err)
	}

	backups, err := FilterOperations(*ops, backup)
	assert := assert.New(t)
	assert.Equal(len(backups), 1)

	operationID := "d3262073-8101-450f-9a11-c851760abd57"

	httpmock.RegisterResponder(http.MethodGet,
		fmt.Sprintf("%s/%s/%s", client.restyClient.HostURL, EndpointOperations, operationID),
		httpmock.NewStringResponder(200,
			fmt.Sprintf(
				`
					{
						"type": "backup",
						"id": "%s",
						"duration": "PT1M0.613S",
						"start": "2019-06-11T03:37:15.593Z",
						"stop": "2019-06-11T03:38:16.206Z",
						"state": "RUNNING"
					}
				`, operationID)))

	// Check GetOperations(id)
	op, err := client.GetOperation(uuid.MustParse(operationID))

	assert.Nil(err)
	assert.Equal((*op)["state"], "RUNNING", "testing backing should return RUNNING state")
}

func TestClientBackupNode(t *testing.T) {

	client := NewSidecarClient(getHost(), &DefaultSidecarClientOptions)

	httpmock.ActivateNonDefault(client.restyClient.GetClient())

	defer httpmock.DeactivateAndReset()

	operationID := "d3262073-8101-450f-9a11-c851760abd57"
	backupResponse := fmt.Sprintf(
		`
			{
				"type": "backup",
				"id": "%s",
				"duration": "PT1M0.613S",
				"start": "2019-06-11T03:37:15.593Z",
				"stop": "2019-06-11T03:38:16.206Z",
				"state": "RUNNING"
			}
		`, operationID)

	responder := func(req *http.Request) (*http.Response, error) {
		body := fmt.Sprintf(`[%s]`, backupResponse)

		header := http.Header{}
		header.Set("Location", fmt.Sprintf("/%s", operationID))

		return &http.Response{
			Status:        strconv.Itoa(http.StatusOK),
			StatusCode:    http.StatusOK,
			Body:          httpmock.NewRespBodyFromString(body),
			Header:        header,
			ContentLength: -1,
		}, nil
	}

	httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s/%s", client.restyClient.HostURL, EndpointOperations),
		responder)

	httpmock.RegisterResponder(http.MethodPost, fmt.Sprintf("%s/%s", client.restyClient.HostURL, EndpointOperations),
		responder)

	httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s/%s/%s", client.restyClient.HostURL, EndpointOperations, operationID),
		httpmock.NewStringResponder(200, backupResponse))

	// List backups and ensure we get only one response
	backups, _ := client.ListBackups()
	assert := assert.New(t)

	assert.Equal(len(backups), 1)

	request := &BackupRequest{
		StorageLocation: "s3://my-bucket/cassandra-dc/test-node-0",
		SnapshotTag:     "mySnapshot",
	}

	// Trigger a backup and ensure we get an uuid
	// then use that uuid to check the backup is running
	if operationID, err := client.StartOperation(request); err != nil {
		t.Error(err.Error())
	} else if getOpResponse, err := client.GetOperation(operationID); err != nil {
		t.Error(err.Error())
	} else {
		assert.Equal((*getOpResponse)["state"], "RUNNING")
		opResponse, _ := client.FindBackup(operationID)
		response := opResponse.basicResponse
		assert.Equal(response.Id, operationID)
		assert.Equal(response.State, operations.OperationState("RUNNING"))
	}

}