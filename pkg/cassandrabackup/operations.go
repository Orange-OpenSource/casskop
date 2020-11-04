package cassandrabackup

import (
	"context"
	"fmt"
	"github.com/antihax/optional"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	"github.com/mitchellh/mapstructure"
)

func (client *client) PerformRestoreOperation(restoreOperationReq icarus.RestoreOperationRequest) (
	*icarus.RestoreOperationResponse, error) {
	var restoreOperation icarus.RestoreOperationResponse
	podClient := client.podClient
	if podClient == nil {
		return nil, ErrNoCassandraBackupClientAvailable
	}

	body, _, err := podClient.OperationsApi.OperationsPost(context.Background(), &icarus.OperationsApiOperationsPostOpts{
		Body: optional.NewInterface(restoreOperationReq),
	},)

	if err != nil {
		return nil, err
	}

	mapstructure.Decode(body, &restoreOperation)
	return &restoreOperation, nil
}

func (client *client) RestoreOperationByID(operationId string) (*icarus.RestoreOperationResponse, error) {
	var restoreOperation icarus.RestoreOperationResponse

	if operationId == "" {
		return nil, fmt.Errorf("must get a non empty id")
	}

	podClient := client.podClient
	if podClient == nil {
		return nil, ErrNoCassandraBackupClientAvailable
	}

	body, _, err :=
		podClient.OperationsApi.OperationsOperationIdGet(context.Background(), operationId)

	if err != nil  {
		return nil, err
	}

	mapstructure.Decode(body, &restoreOperation)
	return &restoreOperation, nil
}

func (client *client) BackupOperationByID(operationId string) (response *icarus.BackupOperationResponse, err error) {

	if operationId == "" {
		return nil, fmt.Errorf("must get a non empty id")
	}

	podClient := client.podClient
	if podClient == nil {
		return nil, ErrNoCassandraBackupClientAvailable
	}

	body, _, err := podClient.OperationsApi.OperationsOperationIdGet(context.Background(), operationId)

	if err != nil  {
		return nil, err
	}

	mapstructure.Decode(body, &response)
	return
}

func (client *client) PerformBackupOperation(request icarus.BackupOperationRequest) (
	*icarus.BackupOperationResponse, error) {
	var backupOperationResponse icarus.BackupOperationResponse

	podClient := client.podClient
	if podClient == nil {
		return nil, ErrNoCassandraBackupClientAvailable
	}

	body, _, err := podClient.OperationsApi.OperationsPost(context.Background(), &icarus.OperationsApiOperationsPostOpts{
		Body: optional.NewInterface(request),
	},)

	if err != nil {
		return nil, err
	}

	mapstructure.Decode(body, &backupOperationResponse)
	return &backupOperationResponse, nil
}