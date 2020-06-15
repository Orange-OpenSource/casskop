package sidecarclient

import (
	"fmt"

	"github.com/antihax/optional"
	csapi "github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
	"github.com/mitchellh/mapstructure"
)

func (cs *cassandraSidecarClient) PerformRestoreOperation(podName string, restoreOperationReq csapi.RestoreOperationRequest) (*csapi.RestoreOperationResponse, error) {
	var restoreOperation csapi.RestoreOperationResponse
	client := cs.podsClient[podName]
	if client == nil {
		log.Error(ErrNoCassandraSidecarClientsAvailable, "Error during creating sidecar client")
		return nil, ErrNoCassandraSidecarClientsAvailable
	}

	body, rsp, err := client.OperationsApi.OperationsPost(nil, &csapi.OperationsApiOperationsPostOpts{
		Body: optional.NewInterface(restoreOperationReq),
	},)

	if err != nil && rsp == nil {
		log.Error(err, "Could not communicate with sidecar")
		return nil, err
	}

	if rsp.StatusCode != 201 {
		log.Error(ErrCassandraSidecarNotReturned201, fmt.Sprintf("Restore cluster gracefully failed since sidecar returned non 201"))
		return nil, ErrCassandraSidecarNotReturned201
	}

	mapstructure.Decode(body, &restoreOperation)

	return &restoreOperation, nil
}

func (cs *cassandraSidecarClient) GetRestoreOperation(podName, operationId string) (*csapi.RestoreOperationResponse, error) {
	var restoreOperation csapi.RestoreOperationResponse
	client := cs.podsClient[podName]
	if client == nil {
		log.Error(ErrNoCassandraSidecarClientsAvailable, "Error during creating sidecar client")
		return nil, ErrNoCassandraSidecarClientsAvailable
	}

	body, rsp, err := client.OperationsApi.OperationsOperationIdGet(nil, operationId)

	if err != nil && rsp == nil {
		log.Error(err, "Could not communicate with sidecar")
		return nil, err
	}

	if rsp.StatusCode != 200 {
		log.Error(ErrCassandraSidecarNotReturned200, fmt.Sprintf("Restore cluster gracefully failed since sidecar returned non 200"))
		return nil, ErrCassandraSidecarNotReturned200
	}

	mapstructure.Decode(body, &restoreOperation)

	return &restoreOperation, nil
}