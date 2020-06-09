package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestoreConditionType represents a valid condition of a Restore.
type RestoreConditionType string

const (
	// RestoreScheduled means the Restore has been assigned to a Cluster
	// member for execution.
	RestoreScheduled RestoreConditionType = "SCHEDULED"
	// RestorePending means the Restore operation is pending for being submitted.
	RestorePending RestoreConditionType = "PENDING"
	// RestoreRunning means the Restore is currently being executed by a
	// Cluster member's mysql-agent side-car.
	RestoreRunning RestoreConditionType = "RUNNING"
	// RestoreComplete means the Restore has successfully executed and the
	// resulting artifact has been stored in object storage.
	RestoreCompleted RestoreConditionType = "COMPLETED"
	// RestoreFailed means the Restore has failed.
	RestoreFailed RestoreConditionType = "FAILED"
	// RestoreCanceled means the Restore operation was interrupted while being run.
	RestoreCanceled RestoreConditionType = "CANCELED"
)

func (r RestoreConditionType) IsInProgress() bool {
	return r == RestorePending || r == RestoreRunning
}

func (r RestoreConditionType) IsInError() bool {
	return r == RestoreFailed || r == RestoreCanceled
}

func (r RestoreConditionType) IsScheduled() bool {
	return r == RestoreScheduled
}

func (r RestoreConditionType) IsPending() bool {
	return r == RestorePending
}

func (r RestoreConditionType) IsRunning() bool {
	return r == RestoreRunning
}

func (r RestoreConditionType) IsCompleted() bool {
	return r == RestoreCompleted
}


// RestoreCondition describes the observed state of a Restore at a certain point.
type RestoreCondition struct {
	Type   RestoreConditionType `json:"type"`
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	Reason string `json:"reason,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

type RestorationPhaseType string

const (
	RestorationPhaseDownload RestorationPhaseType = "DOWNLOAD"
	RestorationPhaseImport RestorationPhaseType = "IMPORT"
	RestorationPhaseTruncate RestorationPhaseType = "TRUNCATE"
	RestorationPhaseCleanup RestorationPhaseType = "CLEANUP"
)

// CassandraRestoreStatus captures the current status of a Cassandra restore.
type CassandraRestoreStatus struct {
	TimeCreated metav1.Time `json:"timeCreated"`
	// TimeStarted is the time at which the restore was started.
	// +optional
	TimeStarted metav1.Time `json:"timeStarted"`
	// TimeCompleted is the time at which the restore completed.
	// +optional
	TimeCompleted metav1.Time `json:"timeCompleted"`
	// +optional
	Condition *RestoreCondition `json:"conditions,omitempty"`
	// Progress is a float32 from 0.0 to 1.0, 1.0 telling that operation is completed, either successfully or with errors
	Progress float32 `json:"conditions,omitempty"`
	//
	RestorationPhase RestorationPhaseType `json:"restorationPhase,omitempty"`
}

// CassandraRestoreSpec defines the specification for a restore of a Cassandra backup.
type CassandraRestoreSpec struct {
	// Cluster is a refeference to the Cluster to which the Restore
	// belongs.
	Cluster *corev1.LocalObjectReference `json:"cluster"`
	// Backup is a reference to the Backup object to be restored.
	Backup *corev1.LocalObjectReference `json:"backup"`
	// ScheduledMember is the Pod name of the Cluster member on which the
	// Restore will be executed.
	ScheduledMember string `json:"scheduledMember"`
	// ConcurrentConnection is the number of threads used for upload, there might be
	// at most so many uploading threads at any given time, when not set, it defaults to 10
	ConcurrentConnection *int32 `json:"concurrentConnection,omitempty"`
	// directory of Cassandra, by default it is /var/lib/cassandra, in this path, one expects there is 'data' directory
	CassandraDirectory string `json:"cassandraDirectory,omitempty"`
	// NoDeleteTruncates is flag saying if we should not delete truncated SSTables
	// after they are imported, as part of CLEANUP phase, defaults to false
	NoDeleteTruncates bool `json:"noDeleteTruncates,omitempty"`
	// SchemaVersion is the version of schema we want to restore from, upon backup, a schema version is automatically appended to snapshot name
	// and its manifest is uploaded under that name. In case we have two snapshots having same name,
	// we might distinguish between them by this schema version. If schema version is not specified,
	// we expect that there will be one and only one backup taken with respective snapshot name.
	// This schema version has to match the version of a Cassandra node we are doing restore for
	// (hence, by proxy, when global request mode is used, all nodes have to be on exact same schema version).
	SchemaVersion string `json:"schemaVersion,omitempty"`
	// ExactSchemaVersion is a flag saying if we indeed want a schema version of a running node match with schema version a snapshot is taken on.
	// There might be cases when we want to restore a table for which its CQL schema has not changed
	// but it has changed for other table / keyspace but a schema for that node has changed by doing that.
	exactSchemaVersion bool `json:"exactSchemaVersion,omitempty""`
}

// +genclient0
// +genclient:noStatus
// +resourceName=cassandrarestores
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraRestore is a Casskop Operator resource that represents the restoration of
// backup of a Cassandra cluster.
type CassandraRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   CassandraRestoreSpec   `json:"spec"`
	Status CassandraRestoreStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraRestoreList is a list of Restores.
type CassandraRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CassandraRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CassandraRestore{}, &CassandraRestoreList{})
}

// GetRestoreCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
func GetRestoreCondition(status *CassandraRestoreStatus, conditionType RestoreConditionType) *RestoreCondition {
	if status.Condition.Type == conditionType {
		return status.Condition
	}
	return nil
}

// UpdateRestoreCondition updates existing Restore condition or creates a new
// one. Sets LastTransitionTime to now if the status has changed.
// Returns true if Restore condition has changed or has been added.
func UpdateRestoreCondition(status *CassandraRestoreStatus, condition *RestoreCondition) bool {
	condition.LastTransitionTime = metav1.Now()
	// Try to find this Restore condition.
	oldCondition := GetRestoreCondition(status, condition.Type)

	if oldCondition == nil {
		// We are updating new Restore condition.
		status.Condition = condition
		return true
	}

	isEqual := condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message &&
		condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	status.Condition = condition
	// Return true if one of the fields have changed.
	return !isEqual
}
