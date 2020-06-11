package sidecar

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/antihax/optional"
	"github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
	"github.com/mitchellh/mapstructure"

	"github.com/Orange-OpenSource/casskop/pkg/common/nodestate"

	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	accept             = "Accept"
	applicationJSON    = "application/json; charset=utf-8"
	EndpointOperations = "operations"
	EndpointStatus     = "status"
)

var log = logf.Log.WithName("SidecarClient")

var DefaultSidecarClientOptions = ClientOptions{Port: 4567, Secure: false}

type Client struct {
	Host      string
	Options   *ClientOptions
	apiClient *cassandrasidecar.APIClient
}

type ClientOptions struct {
	Secure  bool
	Port    int32
	Timeout time.Duration
}

func NewSidecarClient(host string, options *ClientOptions) *Client {

	if options == nil {
		options = &ClientOptions{
			Secure: true,
		}
	}

	if options.Timeout == 0 {
		options.Timeout = 1 * time.Minute
	}

	client := &Client{}
	client.Host = host
	client.Options = options

	var protocol = "https"

	if !client.Options.Secure {
		protocol = "http"
	}

	var port = ""

	if options.Port != 0 {
		port = ":" + strconv.FormatInt(int64(options.Port), 10)
	}

	config := cassandrasidecar.NewConfiguration()
	config.BasePath = fmt.Sprintf("%s://%s:%d/", protocol, client.Host, port)
	config.HTTPClient.Timeout = client.Options.Timeout
	config.Host = client.Host

	client.apiClient = cassandrasidecar.NewAPIClient(config)

	return client
}

func ClientFromPods(podsClients map[*corev1.Pod]*Client, pod corev1.Pod) *Client {

	for key, value := range podsClients {
		if key.Name == pod.Name {
			return value
		}
	}

	return nil
}

type httpResponse struct {
	status       int
	err          error
	responseBody string
}

func (e *httpResponse) Error() string {
	if e.status != 0 {
		return fmt.Sprintf("Operation was not successful, response code %d", e.status)
	}

	return fmt.Sprintf("Operation was errorneous: %s", e.err)
}

func (client *Client) Status() (*nodestate.Status, error) {

	if r, err := client.performRequest(EndpointStatus, http.MethodGet, nil); responseInvalid(r, err) {
		return nil, err
	} else {
		body, err := readBody(r)

		if err != nil {
			return nil, err
		}

		if status, err := unmarshallBody(body, r, &nodestate.Status{}); err == nil {
			return status.(*nodestate.Status), nil
		} else {
			return nil, err
		}
	}
}

func (client *Client) GetOperation(id string) (op *operationResponse, err error) {

	if id == "" {
		return nil, fmt.Errorf("getOperation must get a non empty id")
	}

	if value, _, err :=
		client.apiClient.OperationsApi.OperationsOperationIdGet(context.Background(), id); err != nil {
		return nil, err
	} else {
		var operationResponse operationResponse
		mapstructure.Decode(value, &operationResponse)
		return &operationResponse, nil
	}
}

func (client *Client) GetOperations() (*Operations, error) {
	return client.GetFilteredOperations(nil)
}

func (client *Client) GetFilteredOperations(filter *operationsFilter) (*Operations, error) {
	endpoint := EndpointOperations
	if filter != nil {
		endpoint = filter.buildFilteredEndpoint(endpoint)
	}
	if r, err := client.performRequest(endpoint, http.MethodGet, nil); responseInvalid(r, err) {
		return nil, err
	} else {
		body, err := readBody(r)

		if err != nil {
			return nil, err
		}

		if response, err := unmarshallBody(body, r, &Operations{}); err == nil {
			return response.(*Operations), nil
		} else {
			return nil, err
		}
	}
}

func FilterOperations(ops Operations, kind Kind) (result []interface{}, err error) {

	var op interface{}

	for _, item := range ops {
		if op, err = ParseOperation(item, kind); err != nil {
			log.Error(err, "Error parsing operation", &map[string]interface{}{"Operation": op})
			continue
		}
		if op != nil {
			result = append(result, op)
		}
	}

	return result, nil
}

// func ParseOperation(operation operationResponse, kind Kind) (interface{}, error) {
// 	var op interface{}

// 	if operation["progress"] == "NaN" {
// 		operation["progress"] = 0.0
// 	}

// 	if operation["type"].(string) == _KindValueToName[kind] {

// 		switch kind {
// 		case backup:
// 			op = &BackupResponse{}
// 		case noop:
// 			return nil, fmt.Errorf("no op")
// 		}

// 		if body, err := json.Marshal(operation); err != nil {
// 			return nil, err
// 		} else if err := json.Unmarshal(body, op); err != nil {
// 			return nil, err
// 		}

// 	}
// 	return op, nil
// }

func (client *Client) StartOperation(request operationRequest) (string, error) {
	request.Init()

	var value interface{}
	var err error
	if value, _, err =
		client.apiClient.OperationsApi.OperationsPost(context.Background(),
			&cassandrasidecar.OperationsApiOperationsPostOpts{Body: optional.NewInterface(request)}); err != nil {
		return "", err
	}

	var backupOperationResponse cassandrasidecar.BackupOperationResponse
	mapstructure.Decode(value, &backupOperationResponse)
	return backupOperationResponse.Id, nil
}
