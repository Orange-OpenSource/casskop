package sidecar

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/antihax/optional"
	csd "github.com/cscetbon/cassandrasidecar-go-client/pkg/cassandrasidecar"
	"github.com/mitchellh/mapstructure"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("SidecarClient")

var DefaultSidecarClientOptions = ClientOptions{Port: api.DefaultBackRestSidecarContainerPort, Secure: false}

type Client struct {
	Host      string
	Options   *ClientOptions
	apiClient *csd.APIClient
	Config    *csd.Configuration
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

	config := csd.NewConfiguration()
	config.BasePath = fmt.Sprintf("%s://%s%s", protocol, client.Host, port)
	config.HTTPClient = &http.Client{Timeout: client.Options.Timeout}
	config.Host = client.Host

	client.Config = config

	client.apiClient = csd.NewAPIClient(config)

	return client
}

func (client *Client) GetOperation(id string) (response *csd.BackupOperationResponse, err error) {

	if id == "" {
		return nil, fmt.Errorf("getOperation must get a non empty id")
	}

	var value interface{}

	if value, _, err =
		client.apiClient.OperationsApi.OperationsOperationIdGet(context.Background(), id); err != nil {
		return nil, err
	}

	mapstructure.Decode(value, &response)
	return
}

func (client *Client) StartOperation(request csd.BackupOperationRequest) (id string, err error) {
	var value interface{}
	if value, _, err =
		client.apiClient.OperationsApi.OperationsPost(context.Background(),
			&csd.OperationsApiOperationsPostOpts{Body: optional.NewInterface(request)}); err != nil {
		return "", err
	}

	var backupOperationResponse csd.BackupOperationResponse
	mapstructure.Decode(value, &backupOperationResponse)
	return backupOperationResponse.Id, nil
}
