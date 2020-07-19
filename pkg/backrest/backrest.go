package backrest

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	api "github.com/Orange-OpenSource/casskop/pkg/apis/db/v1alpha1"
	"github.com/Orange-OpenSource/casskop/pkg/cassandrabackup"
	"github.com/Orange-OpenSource/casskop/pkg/controller/common"
	csapi "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/sirupsen/logrus"

)

type Client struct {
	client            cassandrabackup.Client
	CoordinatorMember string
}

func NewClient(client client.Client, cc *api.CassandraCluster, pod *corev1.Pod) (*Client, error) {
	csClient, err := common.NewCassandraBackupConnection(client, cc, pod)
	if err != nil {
		return nil, err
	}

	return &Client{client: csClient, CoordinatorMember: pod.Name}, nil
}

func (c *Client) PerformRestore(restore *api.CassandraRestore,
	backup *api.CassandraBackup) (*api.CassandraRestoreStatus, error) {
	restoreOperationRequest := &csapi.RestoreOperationRequest {
		Type_: "restore",
		StorageLocation: backup.Spec.StorageLocation,
		SnapshotTag: backup.Spec.SnapshotTag,
		NoDeleteTruncates: restore.Spec.NoDeleteTruncates,
		ExactSchemaVersion: restore.Spec.ExactSchemaVersion,
		RestorationPhase: string(api.RestorationPhaseDownload),
		GlobalRequest: true,
		Import_: &csapi.AllOfRestoreOperationRequestImport_{
			Type_: "import",
			SourceDir: "/var/lib/cassandra/data/downloadedsstables",
		},
		Entities: restore.Spec.Entities,
		K8sSecretName: restore.Spec.Secret,
		CassandraDirectory: restore.Spec.CassandraDirectory,
		SchemaVersion: restore.Spec.SchemaVersion,
		RestorationStrategyType: restore.Spec.RestorationStrategyType,
		ConcurrentConnections: *restore.Spec.ConcurrentConnection,
	}

	if len(restore.Spec.Entities) == 0 {
		restoreOperationRequest.Entities = backup.Spec.Entities
	}

	if len(restore.Spec.Secret) == 0 {
		restoreOperationRequest.K8sSecretName = backup.Spec.Secret
	}

	restoreOperation, err := c.client.PerformRestoreOperation(*restoreOperationRequest)
	if err != nil {
		logrus.Error(err, "Restore gracefully failed")
		return nil, err
	}

	logrus.Info("Restore using CassandraBackup sidecar")
	restoreStatus := api.ComputeRestorationStatus(restoreOperation)
	return &restoreStatus, nil
}

func (c *Client) PerformBackup(backup *api.CassandraBackup) (string, error) {
	bandwidth := strings.Replace(backup.Spec.Bandwidth, " ", "", -1)
	bandwidthDataRate, err := dataRateFromBandwidth(bandwidth)

	backupOperationRequest := &csapi.BackupOperationRequest{
		Type_:                 "backup",
		StorageLocation:       backup.Spec.StorageLocation,
		SnapshotTag:           backup.Spec.SnapshotTag,
		Duration:              backup.Spec.Duration,
		Bandwidth:             bandwidthDataRate,
		ConcurrentConnections: backup.Spec.ConcurrentConnections,
		Entities:              backup.Spec.Entities,
		K8sSecretName:         backup.Spec.Secret,
		Dc:                    backup.Spec.Datacenter,
		GlobalRequest:         true,
		K8sNamespace:          backup.Namespace,
	}

	backupOperation, err := c.client.PerformBackupOperation(*backupOperationRequest)
	if err != nil {
		return "", err
	}

	return backupOperation.Id, nil
}

func (c *Client) RestoreStatusByID(id string) (*api.CassandraRestoreStatus, error) {

	restoreOperation, err := c.client.RestoreOperationByID(id)
	if err != nil  {
		logrus.WithFields(logrus.Fields{"id": id}).Error("Cannot find restore operation")
		return nil, err
	}

	status := api.ComputeRestorationStatus(restoreOperation)
	return &status, nil
}

func (c *Client) BackupStatusByID(id string) (*api.CassandraBackupStatus, error) {

	backupOperation, err := c.client.BackupOperationByID(id)
	if err != nil  {
		logrus.WithFields(logrus.Fields{"id": id}).Error("Cannot find backup operation")
		return nil, err
	}

	status := api.ComputeBackupStatus(backupOperation)
	return &status, nil
}

var regexBandwidthSupportedFormat = regexp.MustCompile(`(?i)^(?P<Value>\d+)(?P<Unit>[kmg]?)$`)

func dataRateFromBandwidth(value string) (*csapi.DataRate, error) {
	bandwidth := strings.ToUpper(strings.Replace(value, " ", "", -1))

	if bandwidth == "" {
		return nil, nil
	}

	matches := regexBandwidthSupportedFormat.FindStringSubmatch(bandwidth)
	if matches == nil {
		return nil, fmt.Errorf("Format of %s not supported", value)
	}
	dataValue, _ := strconv.Atoi(matches[1])
	return &csapi.DataRate{Value: int32(dataValue), Unit: matches[2] + "BPS"}, nil
}