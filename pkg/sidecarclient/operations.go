package sidecarclient

import (
	"fmt"

	csapi "github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
)

func (cs *cassandraSidecarClient) PerformRestoreOperation(hostname string, restoreOperation *csapi.RestoreOperationRequest) (*csapi.RestoreOperationResponse, error) {
	client := cs.podsClient[hostname]
	if client == nil {
		log.Error(ErrNoCassandraSidecarClientsAvailable, "Error during creating sidecar client")
		return nil, ErrNoCassandraSidecarClientsAvailable
	}

	inlineRsp200, rsp, err := client.OperationsApi.OperationsPost(nil, &csapi.OperationsApiOperationsPostOpts{
		Body: restoreOperation,
	})
	fmt.Sprintf("test %s, %s, %s", inlineRsp200, rsp.Status, err)
	return nil, nil
}

func (cd *cassandraSidecarClient) GetRestoreOperation() (*csapi.RestartOperationResponse, error) {
	return nil, nil
}