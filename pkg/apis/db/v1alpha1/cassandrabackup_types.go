package v1alpha1

import (
	"encoding/json"
	"strings"

	"github.com/Orange-OpenSource/casskop/pkg/common/operations"
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
	Schedule              string `json:"schedule,omitempty"`
	SnapshotTag           string `json:"snapshotTag"`
	Duration              string `json:"duration,omitempty"`
	Bandwidth             string `json:"bandwidth,omitempty"`
	ConcurrentConnections int    `json:"concurrentConnections,omitempty"`
	Entities              string `json:"entities,omitempty"`
	Secret                string `json:"secret,omitempty"`
}

// CassandraBackupStatus defines the observed state of CassandraBackup
// +k8s:openapi-gen=true
type CassandraBackupStatus struct {
	// name of pod / node
	Node string `json:"node"`
	// State shows the status of the operation
	State operations.OperationState `json:"state"`
	// Progress shows the percentage of the operation done
	Progress string `json:"progress"`
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
	Status         []*CassandraBackupStatus  `json:"status,omitempty"`
	GlobalStatus   operations.OperationState `json:"globalStatus,omitempty"`
	GlobalProgress string                    `json:"globalProgress,omitempty"`
}

func (cb *CassandraBackup) ComputeLastAppliedConfiguration() (string, error) {
	lastcb := cb.DeepCopy()
	//remove unnecessary fields
	lastcb.Annotations = nil
	lastcb.ResourceVersion = ""
	lastcb.Status = make([]*CassandraBackupStatus, 0)
	lastcb.GlobalStatus = ""
	lastcb.GlobalProgress = ""

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