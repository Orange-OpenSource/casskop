package v1alpha1

import (
	"encoding/json"
	csapi "github.com/instaclustr/cassandra-sidecar-go-client/pkg/cassandra_sidecar"
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
	CassandraCluster 	  string `json:"cassandracluster"`
	// Cassandra DC name to back up, used to find the cassandra nodes in the CassandraCluster
	Datacenter 		 	  string `json:"datacenter"`
	// URI for the backup target location e.g. s3 bucket, filepath
	StorageLocation  	  string `json:"storageLocation"`
	// Specify a schedule to assigned to the backup. The schedule doesn't enforce anything so if you schedule multiple
	// backups around the same time they would conflict. See https://godoc.org/github.com/robfig/cron for more information regarding the supported formats
	Schedule           	  string `json:"schedule,omitempty"`
	SnapshotTag      	  string `json:"snapshotTag"`
	// Specify a duration the backup should try to last. See https://golang.org/pkg/time/#ParseDuration for an
	// exhaustive list of the supported units. You can use values like .25h, 15m, 900s all meaning 15 minutes
	Duration         	  string `json:"duration,omitempty"`
	// Specify the bandwidth to not exceed when uploading files to the cloud. Format supported is \d+[KMG] case
	// insensitive. You can use values like 10M (meaning 10MB), 1024, 1024K, 2G, etc...
	Bandwidth        	  string `json:"bandwidth,omitempty"`
	// Maximum number of threads used to download files from the cloud. Defaults to 10
	ConcurrentConnections int32  `json:"concurrentConnections,omitempty"`
	// Database entities to backup, it might be either only keyspaces or only tables prefixed by their respective
	// keyspace, e.g. 'k1,k2' if one wants to backup whole keyspaces or 'ks1.t1,ks2.t2' if one wants to restore specific
	// tables. These formats are mutually exclusive so 'k1,k2.t2' is invalid. An empty field will backup all keyspaces
	Entities              string `json:"entities,omitempty"`
	Secret                string `json:"secret,omitempty"`
}

type BackupState string

const (
	BackupRunning   BackupState = "RUNNING"
	BackupCompleted BackupState = "COMPLETED"
	BackupFailed    BackupState = "FAILED"
)

type CassandraBackupStatus struct {
	// name of pod / node
	CoordinatorMember string `json:"coordinatorMember"`
	// State shows the status of the operation
	State BackupState `json:"state"`
	// Progress shows the percentage of the operation done
	Progress string `json:"progress"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Defines a backup operation and its details
// +k8s:openapi-gen=true
type CassandraBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraBackupSpec    `json:"spec"`
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

func (cb *CassandraBackup) ComputeLastAppliedAnnotation() (string, error) {
	lastcb := cb.DeepCopy()
	//remove unnecessary fields
	lastcb.Annotations = nil
	lastcb.ResourceVersion = ""
	lastcb.Status = nil
	lastcb.Finalizers = nil
	lastcb.ObjectMeta = metav1.ObjectMeta{Name: lastcb.Name, Namespace: lastcb.Namespace,
		CreationTimestamp: lastcb.CreationTimestamp}

	lastApplied, err := json.Marshal(lastcb)
	if err != nil {
		logrus.Errorf("[%s]: Cannot create last-applied-configuration = %v", cb.Name, err)
	}
	return string(lastApplied), err
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type CassandraBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CassandraBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CassandraBackup{}, &CassandraBackupList{})
}

func ComputeBackupStatus(backupOperationResponse *csapi.BackupOperationResponse,
	coordinatorMember string) CassandraBackupStatus{
	return CassandraBackupStatus{
		CoordinatorMember: coordinatorMember,
		Progress: ProgressPercentage(backupOperationResponse.Progress),
		State:    BackupState(backupOperationResponse.State),
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