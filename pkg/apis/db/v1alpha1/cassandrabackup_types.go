package v1alpha1

import (
	"encoding/json"
	"strings"

	csd "github.com/erdrix/cassandrasidecar-go-client/pkg/cassandrasidecar"
	cron "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidateSchedule valides that schedule uses a correct format
func (backupSpec CassandraBackupSpec) ValidateSchedule() error {
	if _, err := cron.ParseStandard(backupSpec.Schedule); err != nil {
		return err
	}

	return nil
}

// CassandraBackupSpec defines the desired state of CassandraBackup
// +k8s:openapi-gen=true
type CassandraBackupSpec struct {
	CassandraCluster string `json:"cassandracluster"`
	// Cassandra DC name to back up. Used to find the pods in the CassandraCluster
	Datacenter string `json:"datacenter"`
	// The uri for the backup target location e.g. s3 bucket, filepath
	StorageLocation string `json:"storageLocation"`
	// The snapshot tag for the backup
	Schedule              string        `json:"schedule,omitempty"`
	SnapshotTag           string        `json:"snapshotTag"`
	Duration              string        `json:"duration,omitempty"`
	Bandwidth             *csd.DataRate `json:"bandwidth,omitempty"`
	ConcurrentConnections int32         `json:"concurrentConnections,omitempty"`
	Entities              string        `json:"entities,omitempty"`
	Secret                string        `json:"secret,omitempty"`
}

type BackupState string

const (
	BackupPending   BackupState = "PENDING"
	BackupRunning   BackupState = "RUNNING"
	BackupCompleted BackupState = "COMPLETED"
	BackupCanceled  BackupState = "CANCELED"
	BackupFailed    BackupState = "FAILED"
)

// CassandraBackupStatus defines the observed state of CassandraBackup
// +k8s:openapi-gen=true
type CassandraBackupStatus struct {
	// name of pod / node
	Node string `json:"node"`
	// State shows the status of the operation
	State BackupState `json:"state"`
	// Progress shows the percentage of the operation done
	Progress string `json:"progress"`
}

func (status *CassandraBackupStatus) SetBackupStatusState(state string) {
	switch state {
	case string(BackupPending):
		status.State = BackupPending
	case string(BackupRunning):
		status.State = BackupRunning
	case string(BackupCompleted):
		status.State = BackupCompleted
	case string(BackupCanceled):
		status.State = BackupCanceled
	case string(BackupFailed):
		status.State = BackupFailed
	}
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraBackup is the Schema for the cassandrabackups API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".globalStatus",description="Restore operation status"
// +kubebuilder:printcolumn:name="Progress",type="string",JSONPath=".globalProgress",description="Restore operation progress"
type CassandraBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec CassandraBackupSpec `json:"spec"`
	// +listType
	Status *CassandraBackupStatus `json:"status,omitempty"`
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
	return cb.Status != nil
}

func (cb *CassandraBackup) ComputeLastAppliedConfiguration() (string, error) {
	lastcb := cb.DeepCopy()
	//remove unnecessary fields
	lastcb.Annotations = nil
	lastcb.ResourceVersion = ""
	lastcb.Status = nil
	lastcb.Finalizers = nil

	lastApplied, err := json.Marshal(lastcb)
	if err != nil {
		logrus.Errorf("[%s]: Cannot create last-applied-configuration = %v", cb.Name, err)
	}
	return string(lastApplied), err
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraBackupList contains a list of CassandraBackup
type CassandraBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CassandraBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CassandraBackup{}, &CassandraBackupList{})
}

// IsS3Backup returns true if the backup type is Amazon S3, otherwise returns false
func (backupSpec *CassandraBackup) IsS3Backup() bool {
	return strings.HasPrefix(backupSpec.Spec.StorageLocation, "s3://")
}

// IsAzureBackup returns true if the backup type is Azure, otherwise returns false
func (backupSpec *CassandraBackup) IsAzureBackup() bool {
	return strings.HasPrefix(backupSpec.Spec.StorageLocation, "azure://")
}

// IsGcpBackup returns true if the backup type is GCP, otherwise returns false
func (backupSpec *CassandraBackup) IsGcpBackup() bool {
	return strings.HasPrefix(backupSpec.Spec.StorageLocation, "gcp://")
}