package cassandrabackup

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	csapi "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
	"github.com/mitchellh/mapstructure"
)

func (client *client) PerformRestoreOperation(restoreOperationReq csapi.RestoreOperationRequest) (*csapi.RestoreOperationResponse, error) {
	var restoreOperation csapi.RestoreOperationResponse
	podClient := client.podClient
	if podClient == nil {
		log.Error(ErrNoCassandraBackupClientAvailable, "Error during creating cassandra backup client")
		return nil, ErrNoCassandraBackupClientAvailable
	}

	body, _, err := podClient.OperationsApi.OperationsPost(context.Background(), &csapi.OperationsApiOperationsPostOpts{
		Body: optional.NewInterface(restoreOperationReq),
	},)

	if err != nil {
		return nil, err
	}

	mapstructure.Decode(body, &restoreOperation)

	return &restoreOperation, nil
}

func (client *client) GetRestoreOperation(operationId string) (*csapi.RestoreOperationResponse, error) {
	var restoreOperation csapi.RestoreOperationResponse

	if operationId == "" {
		return nil, fmt.Errorf("must get a non empty id")
	}

	podClient := client.podClient
	if podClient == nil {
		log.Error(ErrNoCassandraBackupClientAvailable, "Error during creating cassandra backup client")
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

func (client *client) GetBackupOperation(operationId string) (response *csapi.BackupOperationResponse, err error) {

	if operationId == "" {
		return nil, fmt.Errorf("must get a non empty id")
	}

	podClient := client.podClient
	if podClient == nil {
		log.Error(ErrNoCassandraBackupClientAvailable, "Error during creating cassandra backup client")
		return nil, ErrNoCassandraBackupClientAvailable
	}

	body, _, err := podClient.OperationsApi.OperationsOperationIdGet(context.Background(), operationId)

	if err != nil  {
		return nil, err
	}

	mapstructure.Decode(body, &response)
	return
}

func (client *client) PerformBackupOperation(request csapi.BackupOperationRequest) (*csapi.BackupOperationResponse, error) {
	var backupOperationResponse csapi.BackupOperationResponse

	podClient := client.podClient
	if podClient == nil {
		log.Error(ErrNoCassandraBackupClientAvailable, "Error during creating cassandra backup client")
		return nil, ErrNoCassandraBackupClientAvailable
	}

	body, _, err := podClient.OperationsApi.OperationsPost(context.Background(), &csapi.OperationsApiOperationsPostOpts{
		Body: optional.NewInterface(request),
	},)

	if err != nil {
		return nil, err
	}

	mapstructure.Decode(body, &backupOperationResponse)
	return &backupOperationResponse, nil
}