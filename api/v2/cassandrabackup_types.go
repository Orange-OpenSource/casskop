package v2

import (
	"encoding/json"
	"github.com/Orange-OpenSource/casskop/pkg/util"
	icarus "github.com/instaclustr/instaclustr-icarus-go-client/pkg/instaclustr_icarus"
	"strings"

	cron "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (backupSpec CassandraBackupSpec) ValidateScheduleFormat() error {
	if _, err := cron.ParseStandard(backupSpec.Schedule); err != nil {
		return err
	}

	return nil
}

type CassandraBackupSpec struct {
	// Name of the CassandraCluster to backup
	CassandraCluster string `json:"cassandraCluster"`
	// Cassandra DC name to back up, used to find the cassandra nodes in the CassandraCluster
	Datacenter string `json:"datacenter,omitempty"`
	// URI for the backup target location e.g. s3 bucket, filepath
	StorageLocation string `json:"storageLocation"`
	// Specify a schedule to assigned to the backup. The schedule doesn't enforce anything so if you schedule multiple
	// backups around the same time they would conflict. See https://godoc.org/github.com/robfig/cron for more information regarding the supported formats
	Schedule string `json:"schedule,omitempty"`
	// name of snapshot to make so this snapshot will be uploaded to storage location. If not specified, the name of
	// snapshot will be automatically generated and it will have name 'autosnap-milliseconds-since-epoch'
	SnapshotTag string `json:"snapshotTag"`
	// Specify a duration the backup should try to last. See https://golang.org/pkg/time/#ParseDuration for an
	// exhaustive list of the supported units. You can use values like .25h, 15m, 900s all meaning 15 minutes
	Duration string `json:"duration,omitempty"`
	// Specify the bandwidth to not exceed when uploading files to the cloud. Format supported is \d+[KMG] case
	// insensitive. You can use values like 10M (meaning 10MB), 1024, 1024K, 2G, etc...
	Bandwidth string `json:"bandwidth,omitempty"`
	// Maximum number of threads used to download files from the cloud. Defaults to 10
	ConcurrentConnections int32 `json:"concurrentConnections,omitempty"`
	// Database entities to backup, it might be either only keyspaces or only tables prefixed by their respective
	// keyspace, e.g. 'k1,k2' if one wants to backup whole keyspaces or 'ks1.t1,ks2.t2' if one wants to restore specific
	// tables. These formats are mutually exclusive so 'k1,k2.t2' is invalid. An empty field will backup all keyspaces
	Entities string `json:"entities,omitempty"`
	// Name of Secret to use when accessing cloud storage providers
	Secret string `json:"secret,omitempty"`
}

type BackupConditionType string

const (
	BackupRunning   BackupConditionType = "RUNNING"
	BackupCompleted BackupConditionType = "COMPLETED"
	BackupFailed    BackupConditionType = "FAILED"
)

func (b BackupConditionType) IsRunning() bool {
	return b == BackupRunning
}

func (b BackupConditionType) IsCompleted() bool {
	return b == BackupCompleted
}

func (b BackupConditionType) HasFailed() bool {
	return b == BackupFailed
}

// +kubebuilder:object:root=true

// CassandraBackup is the Schema for the cassandrabackups API
// +k8s:openapi-gen=true
type CassandraBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraBackupSpec `json:"spec"`
	Status BackRestStatus      `json:"status,omitempty"`
}

func (cb *CassandraBackup) PreventBackupDeletion(value bool) {
	if value {
		cb.SetFinalizers([]string{"kubernetes.io/unschedule-needed"})
		return
	}
	cb.SetFinalizers([]string{})
}

func (cb *CassandraBackup) IsScheduled() bool {
	return cb.Spec.Schedule != ""
}
func (cb *CassandraBackup) Ran() bool {
	return cb.Status.Condition != nil
}

func (cb *CassandraBackup) ComputeLastAppliedAnnotation() (string, error) {
	lastcb := cb.DeepCopy()
	//remove unnecessary fields
	lastcb.Annotations = nil
	lastcb.ResourceVersion = ""
	lastcb.Status = BackRestStatus{}
	lastcb.Finalizers = nil
	lastcb.ObjectMeta = metav1.ObjectMeta{Name: lastcb.Name, Namespace: lastcb.Namespace,
		CreationTimestamp: lastcb.CreationTimestamp}

	lastApplied, err := json.Marshal(lastcb)
	if err != nil {
		logrus.Errorf("[%s]: Cannot create last-applied-configuration = %v", cb.Name, err)
	}
	return string(lastApplied), err
}

// +kubebuilder:object:root=true

// CassandraBackupList contains a list of CassandraBackup
type CassandraBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CassandraBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CassandraBackup{}, &CassandraBackupList{})
}
func ComputeBackupStatus(backupOperationResponse *icarus.BackupOperationResponse,
	coordinatorMember string) BackRestStatus {
	logrus.Infof("backupOperationResponse object: %+v", backupOperationResponse)

	return BackRestStatus{
		Progress:          ProgressPercentage(backupOperationResponse.Progress),
		ID:                backupOperationResponse.Id,
		TimeCreated:       backupOperationResponse.CreationTime,
		TimeStarted:       backupOperationResponse.StartTime,
		TimeCompleted:     backupOperationResponse.CompletionTime,
		CoordinatorMember: coordinatorMember,
		Condition: &BackRestCondition{
			LastTransitionTime: metav1.Now().Format(util.TimeStampLayout),
			Type:               backupOperationResponse.State,
			FailureCause:       failureCause(backupOperationResponse.Errors),
		},
	}
}

func (backupSpec *CassandraBackup) IsS3Backup() bool {
	return strings.HasPrefix(backupSpec.Spec.StorageLocation, "s3://")
}

func (backupSpec *CassandraBackup) IsAzureBackup() bool {
	return strings.HasPrefix(backupSpec.Spec.StorageLocation, "azure://")
}

func (backupSpec *CassandraBackup) IsGcpBackup() bool {
	return strings.HasPrefix(backupSpec.Spec.StorageLocation, "gcp://")
}

func (backupSpec *CassandraBackup) IsFileBackup() bool {
	return strings.HasPrefix(backupSpec.Spec.StorageLocation, "file://")
}
