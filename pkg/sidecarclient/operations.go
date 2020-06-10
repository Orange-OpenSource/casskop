package sidecarclient

import (
	"fmt"

	"github.com/antihax/optional"
	csapi "github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
)

func (cs *cassandraSidecarClient) PerformRestoreOperation(podName string, restoreOperation csapi.RestoreOperationRequest) (*csapi.RestoreOperationResponse, error) {
	client := cs.podsClient[podName]
	if client == nil {
		log.Error(ErrNoCassandraSidecarClientsAvailable, "Error during creating sidecar client")
		return nil, ErrNoCassandraSidecarClientsAvailable
	}

	inlineRsp200, rsp, err := client.OperationsApi.OperationsPost(nil, &csapi.OperationsApiOperationsPostOpts{
		Body: optional.NewInterface(restoreOperation),
	})
	fmt.Sprintf("test %s, %s, %s", inlineRsp200, rsp.Status, err)
	return nil, nil
}

func (cd *cassandraSidecarClient) GetRestoreOperation() (*csapi.RestartOperationResponse, error) {
	return nil, nil
}