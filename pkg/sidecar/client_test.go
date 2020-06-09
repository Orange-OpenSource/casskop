package sidecar

import (
<<<<<<< HEAD
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
=======
	"fmt"
	"net/http"
	"os"
	"strconv"
>>>>>>> ca75c4c553a8444d3010e61eb47f01684a347719
	"testing"

	"github.com/Orange-OpenSource/casskop/pkg/common/nodestate"
	"github.com/Orange-OpenSource/casskop/pkg/common/operations"
<<<<<<< HEAD
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"

	"gotest.tools/assert"
=======
	"github.com/google/uuid"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
>>>>>>> ca75c4c553a8444d3010e61eb47f01684a347719
)

func getHost() string {

	host := os.Getenv("TEST_SIDECAR_HOST")

	if len(host) == 0 {
		return "127.0.0.1"
	}
	return host
}

func TestDemarshalling(t *testing.T) {

<<<<<<< HEAD
	// status

	client := NewSidecarClient(getHost(), &DefaultSidecarClientOptions)

	client.testMode = true

	client.testResponse = &resty.Response{
		RawResponse: &http.Response{
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "nodeState": "NORMAL"}`))),
			Status:     "200 OK",
			StatusCode: 200,
		},
	}
=======
	client := NewSidecarClient(getHost(), &DefaultSidecarClientOptions)

	httpmock.ActivateNonDefault(client.restyClient.GetClient())

	defer httpmock.DeactivateAndReset()

	// Check Status()
	httpmock.RegisterResponder(http.MethodGet, fmt.Sprintf("%s/%s", client.restyClient.HostURL, EndpointStatus),
		httpmock.NewStringResponder(200, `{"nodeState": "NORMAL"}`))
>>>>>>> ca75c4c553a8444d3010e61eb47f01684a347719

	if status, err := client.Status(); err != nil || status == nil || status.NodeState != nodestate.NORMAL {
		t.Fail()
	}

<<<<<<< HEAD
	// list operations

	client.testResponse = &resty.Response{
		RawResponse: &http.Response{
			Body: ioutil.NopCloser(bytes.NewReader([]byte(
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
				]`))),
			Status:     "200 OK",
			StatusCode: 200,
		},
	}

	// get operations

	ops, err := client.GetOperations()
	if err != nil {
		t.Error(err)
	}

	backups, err := FilterOperations(*ops, backup)
	assert.Assert(t, len(backups) == 1)

	decommissions, err := FilterOperations(*ops, decommission)
	assert.Assert(t, len(decommissions) == 1)

	// get operation

	client.testResponse = &resty.Response{
		RawResponse: &http.Response{
			Body: ioutil.NopCloser(bytes.NewReader([]byte(
				`{
=======
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
>>>>>>> ca75c4c553a8444d3010e61eb47f01684a347719
					"type": "backup",
					"id": "d3262073-8101-450f-9a11-c851760abd57",
					"duration": "PT1M0.613S",
					"start": "2019-06-11T03:37:15.593Z",
					"stop": "2019-06-11T03:38:16.206Z",
					"state": "RUNNING"
<<<<<<< HEAD
				}`))),
			Status:     "200 OK",
			StatusCode: 200,
		},
	}

	op, err := client.GetOperation(uuid.MustParse("d3262073-8101-450f-9a11-c851760abd57"))

=======
				}
			]`))

	ops, err := client.GetOperations()
>>>>>>> ca75c4c553a8444d3010e61eb47f01684a347719
	if err != nil {
		t.Error(err)
	}

<<<<<<< HEAD
	if (*op)["state"] != "RUNNING" {
		t.Errorf("testing backing should return RUNNING state")
	}
}

func TestSidecarClient_GetStatus(t *testing.T) {

	client := NewSidecarClient(getHost(), &DefaultSidecarClientOptions)

	status, e := client.Status()

	if e != nil {
		t.Errorf(e.Error())
	}

	if status == nil {
		t.Errorf("Status endpoint has not returned error but its status is not set.")
	}

	if status.NodeState != nodestate.NORMAL {
		t.Fatalf("Expected NORMAL operation mode but received %v", status.NodeState)
	}
}

func TestClient_DecommissionNode(t *testing.T) {

	client := NewSidecarClient(getHost(), &DefaultSidecarClientOptions)

	// first decommissioning

	if operationID, err := client.StartOperation(&decommissionRequest{}); err != nil {
		t.Errorf(err.Error())
	} else if operationID == uuid.Nil {
		t.Errorf("there is not any error from decommission endpoint but operationId is nil")
	} else if getOpResponse, err := client.GetOperation(operationID); err != nil {
		t.Errorf(err.Error())
	} else {
		assert.Assert(t, (*getOpResponse)["state"] == operations.RUNNING)
	}

	// second decommissioning on the same node

	operationID, err := client.StartOperation(&decommissionRequest{})

	if err == nil || operationID != uuid.Nil {
		t.Errorf("Decommissioning of already decomissioned node should fail.")
	}
}

func TestClient_BackupNode(t *testing.T) {

	client := NewSidecarClient(getHost(), &DefaultSidecarClientOptions)

	backups, _ := client.ListBackups()
	fmt.Println(backups[0])
=======
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
>>>>>>> ca75c4c553a8444d3010e61eb47f01684a347719

	request := &BackupRequest{
		StorageLocation: "s3://my-bucket/cassandra-dc/test-node-0",
		SnapshotTag:     "mySnapshot",
	}

<<<<<<< HEAD
	if operationID, err := client.StartOperation(request); err != nil {
		t.Errorf(err.Error())
	} else if getOpResponse, err := client.GetOperation(operationID); err != nil {
		t.Errorf(err.Error())
	} else {
		assert.Assert(t, (*getOpResponse)["state"] == "RUNNING")
	}
}

func TestClient_CleanupNode(t *testing.T) {

	client := NewSidecarClient(getHost(), &DefaultSidecarClientOptions)

	cleanups, _ := client.ListCleanups()
	fmt.Print(cleanups[0])

	request := &cleanupRequest{
		Keyspace: "test",
		Tables:   []string{"mytable", "mytable2"},
	}

	if operationID, err := client.StartOperation(request); err != nil {
		t.Errorf(err.Error())
	} else if operationID == uuid.Nil {
		t.Errorf("there is not any error from cleanup endpoint but operationID is nil")
	} else if getOpResponse, err := client.GetOperation(operationID); err != nil {
		t.Errorf(err.Error())
	} else {
		assert.Assert(t, (*getOpResponse)["state"] == "COMPLETED")
	}
=======
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

>>>>>>> ca75c4c553a8444d3010e61eb47f01684a347719
}
